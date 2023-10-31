package detector

import (
	"context"
	"detect-server/log"
	"fmt"
	"github.com/go-ping/ping"
	"sync"
	"time"
)

type IcmpDetectorOptions struct {
	DefaultTimeout int
	DefaultCount   int
	MaxRunnerCount int
	MaxTargetQueue int
}

type DetectOptions struct {
	Ip      string
	Count   int
	Timeout int
}

type RunnerController struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type icmpDetectorController struct {
	sync.RWMutex
	controls map[string]*RunnerController
}

func (controller *icmpDetectorController) Add(name string, runner *RunnerController) error {
	controller.Lock()
	defer controller.Unlock()
	var _, exist = controller.controls[name]
	if exist {
		return fmt.Errorf("runner name %s already exist", name)
	}
	controller.controls[name] = runner
	return nil
}

func (controller *icmpDetectorController) Get(name string) (*RunnerController, bool) {
	controller.RLock()
	defer controller.RUnlock()
	var runner, exist = controller.controls[name]
	return runner, exist
}

type IcmpDetector struct {
	Options          IcmpDetectorOptions
	targetQueue      chan DetectOptions
	parentCtx        context.Context
	parentCancelFunc context.CancelFunc
	controllers      *icmpDetectorController
}

func NewIcmpDetector(options IcmpDetectorOptions) *IcmpDetector {
	if options.DefaultTimeout <= 0 {
		options.DefaultTimeout = 1000
	}
	if options.DefaultCount <= 0 {
		options.DefaultCount = 3
	}
	if options.MaxTargetQueue <= 0 {
		options.MaxTargetQueue = 10
	}
	if options.MaxRunnerCount <= 0 {
		options.MaxRunnerCount = 10
	}
	var ctx, cancelFunc = context.WithCancel(context.TODO())
	var detector = &IcmpDetector{
		Options:          options,
		targetQueue:      make(chan DetectOptions, options.MaxTargetQueue),
		parentCtx:        ctx,
		parentCancelFunc: cancelFunc,
		controllers: &icmpDetectorController{
			controls: make(map[string]*RunnerController),
		},
	}

	return detector
}

func (detector *IcmpDetector) Start() error {
	if detector.targetQueue == nil {
		detector.targetQueue = make(chan DetectOptions, detector.Options.MaxTargetQueue)
	}
	for i := 1; i <= detector.Options.MaxRunnerCount; i++ {
		go func(idx int) {
			err := detector.startRunner(fmt.Sprintf("runner%d", idx))
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
		case target := <-detector.targetQueue:
			stat, err := detector.Detect(target)
			if err != nil {
				log.Logger.Errorf("%s", err)
				continue
			}
			log.Logger.Infof("%s stat: %v", target.Ip, *stat)
		}
	}
}

func (detector *IcmpDetector) Stop() error {
	for length := len(detector.targetQueue); length > 0; length = len(detector.targetQueue) {
		time.Sleep(time.Second)
	}
	log.Logger.Infof("current length of target queue is 0, stopping detector")
	detector.parentCancelFunc()
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

func (detector *IcmpDetector) Detect(target DetectOptions) (*ping.Statistics, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Logger.Errorf("ping %s failed. %s\n", target.Ip, err)
		}
	}()

	pinger, err := ping.NewPinger(target.Ip)
	pinger.SetPrivileged(true)
	if err != nil {
		return nil, err
	}
	if target.Count <= 0 {
		target.Count = detector.Options.DefaultCount
	}
	pinger.Count = target.Count

	if target.Timeout <= 0 {
		target.Timeout = detector.Options.DefaultTimeout
	}
	pinger.Timeout = time.Duration(target.Timeout) * time.Millisecond
	if err = pinger.Run(); err != nil {
		return nil, err
	}

	return pinger.Statistics(), nil
}

func (detector *IcmpDetector) Detects(targets []DetectOptions) {
	for _, target := range targets {
		detector.targetQueue <- target
	}
}
