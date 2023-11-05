package sender

import (
	"detect-server/connector"
)

type Sender[T any] interface {
	Start() error
	Stop() error
	AddReceiver(connector.Receiver[T])
}
