package dispatcher

import "detect-server/detector"

type DefaultMessage struct {
	Type   detector.DetectType
	Target string
	Count  int
	Detail any
	Error  error
}

type MessageOutput interface {
	DefaultMessage
}

type Processor[T detector.DetectInput, R detector.DetectOutput, F MessageOutput] interface {
	Process(result detector.DetectResult[T, R]) F
}

type defaultProcessor[T detector.DetectInput, R detector.DetectOutput, F MessageOutput] struct {
}

func (process *defaultProcessor[T, R, F]) Process(in detector.DetectResult[T, R]) F {
	var out = F(DefaultMessage{
		Type:   in.Target.Type,
		Target: in.Target.Target,
		Count:  in.Target.Options.Count,
		Detail: in.Result,
		Error:  in.Error,
	})
	return out
}

func NewDefaultProcessor[T detector.DetectInput, R detector.DetectOutput, F MessageOutput]() Processor[T, R, F] {
	return &defaultProcessor[T, R, F]{}
}
