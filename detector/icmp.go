package detector

import (
	"context"
	"fmt"
	"log"
)

type IcmpDetectorOptions struct {
	DefaultTimeout int
	DefaultCount   int
	MaxRunnerCount int
	MaxTargetQueue int
}

type DetectOptions struct {
	Ip string
}

type IcmpDetector struct {
	Options     IcmpDetectorOptions
	targetQueue chan DetectOptions
	ctx         context.Context
	cancelFuncs map[string]context.CancelFunc
}

func (detector *IcmpDetector) Start() error {
	if detector.targetQueue == nil {
		detector.targetQueue = make(chan DetectOptions, detector.Options.MaxTargetQueue)
	}
	for i := 1; i <= detector.Options.MaxRunnerCount; i++ {
		go func(idx int) {
			err := detector.startRunner(fmt.Sprintf("runner%d", idx))
			if err != nil {
				log.Println(err)
			}
		}(i)
	}
	return nil
}

func (detector *IcmpDetector) startRunner(name string) error {
	ctx, cancelFunc := context.WithCancel(detector.ctx)
	_, exist := detector.cancelFuncs[name]
	if exist {
		cancelFunc()
		return fmt.Errorf("runner name %s realdy exist", name)
	}
	detector.cancelFuncs[name] = cancelFunc
	for {
		select {
		case <-ctx.Done():
			log.Println("stop")
			return nil
		case target := <-detector.targetQueue:
			err := detector.Detect(target)
			if err != nil {

			}
		}
	}
}

func (detector *IcmpDetector) Stop() error {
	return nil
}

func (detector *IcmpDetector) Detect(target DetectOptions) error {
	return nil
}
