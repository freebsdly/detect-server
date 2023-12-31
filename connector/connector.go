package connector

// Receiver as the interface, any api can send detect job
// to Receiver
type Receiver[T any] interface {
	Receive() <-chan T
}

type Publisher[T any] interface {
	Publish() chan<- T
}

type Connector[T any] interface {
	Receiver[T]
	Publisher[T]
}

type Options struct {
	MaxBufferSize int
}

type chanConnector[T any] struct {
	buffer  chan T
	options Options
}

func NewChanConnector[T any](options Options) Connector[T] {
	var connector = &chanConnector[T]{
		buffer:  make(chan T, options.MaxBufferSize),
		options: options,
	}

	return connector
}

func (receiver *chanConnector[T]) Publish() chan<- T {
	if receiver.buffer == nil {
		receiver.buffer = make(chan T, receiver.options.MaxBufferSize)
	}

	return receiver.buffer
}

func (receiver *chanConnector[T]) Receive() <-chan T {
	if receiver.buffer == nil {
		receiver.buffer = make(chan T, receiver.options.MaxBufferSize)
	}

	return receiver.buffer
}
