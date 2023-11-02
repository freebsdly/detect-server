package detector

import (
	"context"
	"detect-server/log"
	"fmt"
	"github.com/go-ping/ping"
	"time"
)

type IcmpDetect struct {
	Target  string
	Count   int
	Timeout int
}

type IcmpDetectorOptions struct {
	DefaultTimeout     int
	DefaultCount       int
	MaxRunnerCount     int
	MaxTaskBufferSize  int
	MaxResultQueueSize int
}

// IcmpDetector detect target use icmp protocol
type IcmpDetector struct {
	options          IcmpDetectorOptions
	taskBuffer       chan Task[IcmpDetect]
	resultQueue      chan Result[IcmpDetect, *ping.Statistics]
	parentCtx        context.Context
	parentCancelFunc context.CancelFunc
	controllers      *Controller
}

func NewIcmpDetector(options IcmpDetectorOptions) *IcmpDetector {
	if options.DefaultTimeout <= 0 {
		options.DefaultTimeout = 1000
	}
	if options.DefaultCount <= 0 {
		options.DefaultCount = 3
	}
	if options.MaxTaskBufferSize <= 0 {
		options.MaxTaskBufferSize = 10
	}
	if options.MaxRunnerCount <= 0 {
		options.MaxRunnerCount = 10
	}
	var ctx, cancelFunc = context.WithCancel(context.TODO())
	var detector = &IcmpDetector{
		options:          options,
		taskBuffer:       make(chan Task[IcmpDetect], options.MaxTaskBufferSize),
		resultQueue:      make(chan Result[IcmpDetect, *ping.Statistics], options.MaxResultQueueSize),
		parentCtx:        ctx,
		parentCancelFunc: cancelFunc,
		controllers:      NewController(),
	}

	return detector
}

func (detector *IcmpDetector) Start() error {
	if detector.resultQueue == nil {
		return fmt.Errorf("result queue can not be nil")
	}

	if detector.controllers.RunningCount() > 0 {
		return fmt.Errorf("can not start detector, it's running")
	}

	if detector.taskBuffer == nil {
		detector.taskBuffer = make(chan Task[IcmpDetect], detector.options.MaxTaskBufferSize)
	}

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
	var err = detector.controllers.Add(name, &RunnerController{
		ctx:        ctx,
		cancelFunc: cancelFunc,
	})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Logger.Infof("stopping runner %s", name)
			return nil
		case task := <-detector.taskBuffer:
			log.Logger.Debugf("start run task: %v", task)
			for _, target := range task.Targets {
				var result = detector.Detect(target)
				detector.resultQueue <- result
			}
		}
	}
}

func (detector *IcmpDetector) Stop() error {
	for length := len(detector.taskBuffer); length > 0; length = len(detector.taskBuffer) {
		time.Sleep(time.Second)
	}
	log.Logger.Infof("current length of target queue is 0, stopping detector")
	detector.parentCancelFunc()
	detector.controllers = NewController()
	return nil
}

func (detector *IcmpDetector) stopRunner(name string) error {
	controller, exist := detector.controllers.Get(name)
	if !exist {
		return fmt.Errorf("can not found runner %s", name)
	}
	controller.cancelFunc()
	return nil

}

func (detector *IcmpDetector) Detect(detect IcmpDetect) Result[IcmpDetect, *ping.Statistics] {
	var result = Result[IcmpDetect, *ping.Statistics]{
		Detect: detect,
	}
	pinger, err := ping.NewPinger(detect.Target)
	pinger.SetPrivileged(true)
	if err != nil {
		result.Error = err
		return result
	}
	if detect.Count <= 0 {
		detect.Count = detector.options.DefaultCount
	}
	pinger.Count = detect.Count

	if detect.Timeout <= 0 {
		detect.Timeout = detector.options.DefaultTimeout
	}
	pinger.Timeout = time.Duration(detect.Timeout) * time.Millisecond
	if err = pinger.Run(); err != nil {
		result.Error = err
		return result
	}
	result.Data = pinger.Statistics()
	return result
}

// Detects detect target async
func (detector *IcmpDetector) Detects() chan<- Task[IcmpDetect] {
	return detector.taskBuffer
}

func (detector *IcmpDetector) Results() <-chan Result[IcmpDetect, *ping.Statistics] {
	return detector.resultQueue
}
