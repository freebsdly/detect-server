package detector

import (
	"context"
	"detect-server/log"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/spf13/viper"
	"time"
)

type IcmpOptions struct {
}

type IcmpDetectorOptions struct {
	DefaultTimeout      int
	DefaultCount        int
	MaxRunnerCount      int
	MaxDetectBufferSize int
	MaxResultQueueSize  int
}

func NewIcmpDetectorOptions() IcmpDetectorOptions {
	var options = IcmpDetectorOptions{
		DefaultTimeout:      viper.GetInt("detector.icmp.detect.timeout"),
		DefaultCount:        viper.GetInt("detector.icmp.detect.count"),
		MaxRunnerCount:      viper.GetInt("detector.icmp.runner.count"),
		MaxDetectBufferSize: viper.GetInt("detector.icmp.detect.buffer.size"),
		MaxResultQueueSize:  viper.GetInt("detector.icmp.detect.result.queue.size"),
	}

	if options.DefaultTimeout <= 0 {
		options.DefaultTimeout = 1000
	}
	if options.DefaultCount <= 0 {
		options.DefaultCount = 3
	}
	if options.MaxDetectBufferSize <= 0 {
		options.MaxDetectBufferSize = 256
	}
	if options.MaxRunnerCount <= 0 {
		options.MaxRunnerCount = 10
	}

	return options
}

// IcmpDetector detect target use icmp protocol
type IcmpDetector struct {
	options          IcmpDetectorOptions
	detectBuffer     chan DetectTarget[IcmpOptions]
	resultQueue      chan DetectResult[IcmpOptions, *ping.Statistics]
	parentCtx        context.Context
	parentCancelFunc context.CancelFunc
}

func NewIcmpDetector(options IcmpDetectorOptions) Detector[IcmpOptions, *ping.Statistics] {
	var detector = &IcmpDetector{
		options:      options,
		detectBuffer: make(chan DetectTarget[IcmpOptions], options.MaxDetectBufferSize),
		resultQueue:  make(chan DetectResult[IcmpOptions, *ping.Statistics], options.MaxResultQueueSize),
	}

	return detector
}

func (detector *IcmpDetector) Start() error {
	if detector.resultQueue == nil {
		return fmt.Errorf("result queue can not be nil")
	}

	if detector.detectBuffer == nil {
		return fmt.Errorf("detect queue can not be nil")
	}
	detector.parentCtx, detector.parentCancelFunc = context.WithCancel(context.Background())
	for i := 1; i <= detector.options.MaxRunnerCount; i++ {
		go func(idx int) {
			var runnerName = fmt.Sprintf("runner%d", idx)
			log.Logger.Debugf("start icmp detector runner %s", runnerName)
			err := detector.startRunner(runnerName)
			if err != nil {
				log.Logger.Errorf("%s", err)
			}
		}(i)
	}
	return nil
}

func (detector *IcmpDetector) startRunner(name string) error {
	var ctx, cancelFunc = context.WithCancel(detector.parentCtx)
	defer cancelFunc()

	for {
		select {
		case <-ctx.Done():
			log.Logger.Infof("stopping runner %s", name)
			return nil
		case detect := <-detector.detectBuffer:
			var result = detector.Detect(detect)
			detector.resultQueue <- result

		}
	}
}

func (detector *IcmpDetector) Stop() error {
	for length := len(detector.detectBuffer); length > 0; length = len(detector.detectBuffer) {
		time.Sleep(time.Second)
	}
	log.Logger.Infof("current length of target queue is 0, stopping detector")
	detector.parentCancelFunc()
	return nil
}

func (detector *IcmpDetector) Detect(target DetectTarget[IcmpOptions]) DetectResult[IcmpOptions, *ping.Statistics] {
	var result = DetectResult[IcmpOptions, *ping.Statistics]{
		Target: target,
	}
	pinger, err := ping.NewPinger(target.Target)
	pinger.SetPrivileged(true)
	if err != nil {
		result.Error = err
		return result
	}
	if target.Options.Count <= 0 {
		target.Options.Count = detector.options.DefaultCount
	}
	pinger.Count = target.Options.Count

	if target.Options.Timeout <= 0 {
		target.Options.Timeout = detector.options.DefaultTimeout
	}
	pinger.Timeout = time.Duration(target.Options.Timeout) * time.Millisecond
	result.Error = pinger.Run()
	result.Result = pinger.Statistics()
	return result
}

// Detects detect target async
func (detector *IcmpDetector) Detects() chan<- DetectTarget[IcmpOptions] {
	return detector.detectBuffer
}

func (detector *IcmpDetector) Results() <-chan DetectResult[IcmpOptions, *ping.Statistics] {
	return detector.resultQueue
}
