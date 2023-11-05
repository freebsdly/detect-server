package detector

import (
	"github.com/go-ping/ping"
)

// Detector support sync and async method to detect target
type Detector[T DetectInput, R DetectOutput] interface {
	// Start Detector
	Start() error

	// Stop Detector, it will wait for all detect target finished
	Stop() error
	// Detect detect single target and return result sync
	Detect(target DetectTarget[T]) DetectResult[T, R]
	// Detects detect many targets
	Detects() chan<- DetectTarget[T]
	Results() <-chan DetectResult[T, R]
}

type DetectType = string

const (
	ICMPDetect DetectType = "icmp"
	TCPDetect  DetectType = "tcp"
	UDPDetect  DetectType = "udp"
)

type DetectStatus = string

type DetectInput interface {
	IcmpOptions | TcpOptions | UdpOptions | HttpOptions
}

type DetectOutput interface {
	*ping.Statistics
}

type DetectTarget[T DetectInput] struct {
	Type    DetectType
	Target  string
	Options DetectOptions[T]
}

type DetectOptions[T DetectInput] struct {
	Count   int
	Timeout int
	Options T
}

type DetectResult[T DetectInput, R DetectOutput] struct {
	Target DetectTarget[T]
	Result R
	Error  error
}

func NewDetectTarget[T DetectInput](t DetectType, target string, options DetectOptions[T]) DetectTarget[T] {
	return DetectTarget[T]{
		Type:    t,
		Target:  target,
		Options: options,
	}
}

func NewDetectResult[T DetectInput, R DetectOutput](target DetectTarget[T], result R, err error) DetectResult[T, R] {
	return DetectResult[T, R]{
		Target: target,
		Result: result,
		Error:  err,
	}
}
