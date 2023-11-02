package dispatcher

import (
	"context"
	"detect-server/connector"
	"detect-server/detector"
	"detect-server/log"
	"detect-server/sender"
	"fmt"
	"github.com/go-ping/ping"
)

type Dispatcher struct {
	options      Options
	icmpReceiver connector.Receiver[detector.Task[detector.IcmpDetect]]
	icmpDetector detector.Detector[detector.IcmpDetect, *ping.Statistics]
	kafkaSender  *sender.KafkaSender
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

type Options struct {
	MaxIcmpResultQueueSize int
	IcmpReceiverOptions    connector.Options
}

func NewDispatcher(options Options) *Dispatcher {
	var ctx, cancelFunc = context.WithCancel(context.TODO())
	return &Dispatcher{
		options:    options,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
}

func (dis *Dispatcher) AddIcmpReceiver(receiver connector.Receiver[detector.Task[detector.IcmpDetect]]) {
	dis.icmpReceiver = receiver
}

func (dis *Dispatcher) AddIcmpDetector(detector detector.Detector[detector.IcmpDetect, *ping.Statistics]) {
	dis.icmpDetector = detector
}

func (dis *Dispatcher) AddSender(sender *sender.KafkaSender) {
	dis.kafkaSender = sender
}

func (dis *Dispatcher) icmpConnect() error {
	if dis.icmpReceiver == nil {
		return fmt.Errorf("icmp connector is not setup")
	}
	if dis.icmpDetector == nil {
		return fmt.Errorf("icmp detector is not setup")
	}
	var ctx, cancelFunc = context.WithCancel(dis.ctx)
	defer cancelFunc()
	for {
		select {
		case <-ctx.Done():
			log.Logger.Infof("stop connecting icmp connector and detector")
			return nil
		case task := <-dis.icmpReceiver.Receive():
			log.Logger.Debugf("send task to icmp detector. %v", task)
			dis.icmpDetector.Detects() <- task
		case result := <-dis.icmpDetector.Results():
			if result.Error != nil {
				log.Logger.Errorf("detect target %s failed. %s", result.Detect.Target, result.Error)
			}
			err := dis.kafkaSender.SendMessage(result)
			if err != nil {
				log.Logger.Errorf("send message to kafka failed. %s", err)
			}
		}
	}
}

func (dis *Dispatcher) Start() error {
	dis.ctx, dis.cancelFunc = context.WithCancel(context.TODO())

	go func() {
		var err = dis.icmpConnect()
		if err != nil {
			log.Logger.Errorf("connect icmp connector and detector failed. %s", err)
		}
	}()

	return nil
}

func (dis *Dispatcher) Stop() error {
	if dis.cancelFunc == nil {
		return fmt.Errorf("cancel is unable")
	}
	dis.cancelFunc()
	return nil
}
