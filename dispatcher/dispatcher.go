package dispatcher

import (
	"context"
	"detect-server/connector"
	"detect-server/detector"
	"detect-server/log"
	"fmt"
)

type Dispatcher[T detector.DetectInput, R detector.DetectOutput, F MessageOutput] interface {
	Start() error
	Stop() error
	AddReceiver(connector.Receiver[Task[T]])
	AddDetector(detector.Detector[T, R])
	AddPublisher(connector.Publisher[any])
	AddProcessor(processor Processor[T, R, F])
}

// Task may contain many targets
type Task[T detector.DetectInput] interface {
	Name() string
	Targets() []detector.DetectTarget[T]
}

type task[T detector.DetectInput] struct {
	name    string
	targets []detector.DetectTarget[T]
}

func (t *task[T]) Name() string {
	return t.name
}

func (t *task[T]) Targets() []detector.DetectTarget[T] {
	return t.targets
}

func NewTask[T detector.DetectInput](name string, targets []detector.DetectTarget[T]) Task[T] {
	return &task[T]{
		name:    name,
		targets: targets,
	}
}

type Options struct {
}

func NewOptions() Options {
	return Options{}
}

type commonDispatcher[T detector.DetectInput, R detector.DetectOutput, F MessageOutput] struct {
	options    Options
	ctx        context.Context
	cancelFunc context.CancelFunc
	receiver   connector.Receiver[Task[T]]
	detector   detector.Detector[T, R]
	publisher  connector.Publisher[any]
	processor  Processor[T, R, F]
}

func NewDispatcher[T detector.DetectInput, R detector.DetectOutput, F MessageOutput](options Options) Dispatcher[T, R, F] {
	return &commonDispatcher[T, R, F]{
		options: options,
	}
}

func (dispatch *commonDispatcher[T, R, F]) AddReceiver(receiver connector.Receiver[Task[T]]) {
	dispatch.receiver = receiver
}

func (dispatch *commonDispatcher[T, R, F]) AddDetector(detector detector.Detector[T, R]) {
	dispatch.detector = detector
}

func (dispatch *commonDispatcher[T, R, F]) AddPublisher(publisher connector.Publisher[any]) {
	dispatch.publisher = publisher
}

func (dispatch *commonDispatcher[T, R, F]) AddProcessor(processor Processor[T, R, F]) {
	dispatch.processor = processor
}

func (dispatch *commonDispatcher[T, R, F]) Start() error {
	if dispatch.receiver == nil {
		return fmt.Errorf("receiver is invalid")
	}
	if dispatch.detector == nil {
		return fmt.Errorf("detector is invalid")
	}
	if dispatch.publisher == nil {
		return fmt.Errorf("publisher is invalid")
	}
	if dispatch.processor == nil {
		return fmt.Errorf("processor is invalid")
	}
	dispatch.ctx, dispatch.cancelFunc = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Logger.Infof("stop dispatcher send to detector")
				return
			case task := <-dispatch.receiver.Receive():
				for _, target := range task.Targets() {
					dispatch.detector.Detects() <- target
				}
			}
		}
	}(dispatch.ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Logger.Infof("stop dispatcher send to sender")
				return
			case result := <-dispatch.detector.Results():
				log.Logger.Debugf("detect result: %v", result)
				dispatch.publisher.Publish() <- dispatch.processor.Process(result)

			}
		}
	}(dispatch.ctx)
	return nil
}

func (dispatch *commonDispatcher[T, R, F]) Stop() error {
	if dispatch.cancelFunc != nil {
		dispatch.cancelFunc()
	}
	return nil
}
