package receiver

// Receiver as the interface, any api can send detect job
// to Receiver
type Receiver[T any] interface {
	Publish() chan<- T
	Subscribe() <-chan T
}

type Options struct {
	MaxBufferSize int
}

type commonReceiver[T any] struct {
	buffer  chan T
	options Options
}

func NewReceiver[T any](options Options) Receiver[T] {
	var receiver = &commonReceiver[T]{
		buffer:  make(chan T, options.MaxBufferSize),
		options: options,
	}

	return receiver
}

func (receiver *commonReceiver[T]) Publish() chan<- T {
	if receiver.buffer == nil {
		receiver.buffer = make(chan T, receiver.options.MaxBufferSize)
	}

	return receiver.buffer
}

func (receiver *commonReceiver[T]) Subscribe() <-chan T {
	if receiver.buffer == nil {
		receiver.buffer = make(chan T, receiver.options.MaxBufferSize)
	}

	return receiver.buffer
}
