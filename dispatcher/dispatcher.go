package dispatcher

import (
	"context"
	"detect-server/detector"
	"detect-server/log"
	"detect-server/receiver"
	"fmt"
	"github.com/go-ping/ping"
)

type Dispatcher struct {
	options      Options
	icmpReceiver receiver.Receiver[detector.Task[detector.IcmpDetect]]
	icmpDetector detector.Detector[detector.IcmpDetect, *ping.Statistics]
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

type Options struct {
	MaxIcmpResultQueueSize int
	IcmpReceiverOptions    receiver.Options
}

func NewDispatcher(options Options) *Dispatcher {
	var (
		icmpReceiver = receiver.NewReceiver[detector.Task[detector.IcmpDetect]](options.IcmpReceiverOptions)
	)
	return &Dispatcher{
		options:      options,
		icmpReceiver: icmpReceiver,
	}
}

func (dis *Dispatcher) Start() error {
	dis.ctx, dis.cancelFunc = context.WithCancel(context.TODO())

	for {
		select {
		case <-dis.ctx.Done():
			log.Logger.Infof("stop dispatcher")
			return nil
		case task := <-dis.icmpReceiver.Subscribe():
			dis.icmpDetector.Detects() <- task
		case result := <-dis.icmpDetector.Results():
			if result.Error != nil {
				log.Logger.Errorf("detect target %s failed. %s", result.Detect.Target, result.Error)
			}
		}
	}
}

func (dis *Dispatcher) Stop() error {
	if dis.cancelFunc == nil {
		return fmt.Errorf("cancel is unable")
	}
	dis.cancelFunc()
	return nil
}

func (dis *Dispatcher) GetIcmpReceiver() chan<- detector.Task[detector.IcmpDetect] {
	return dis.icmpReceiver.Publish()
}
