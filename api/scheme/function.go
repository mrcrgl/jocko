package scheme

import (
	"errors"

	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
	"github.com/travisjeffery/jocko"
)

var (
	ErrHandlerNotSetUp = errors.New("api handler function not set up")
)

func NewFunction(key int16, v APIVersion, h machinery.Handler) APIFunction {
	return APIFunction{enabled: true, Key: key, Version: v, handler: h}
}

type APIFunction struct {
	enabled bool
	Key     int16
	Version APIVersion
	handler machinery.Handler
}

func (af APIFunction) Handle(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	if af.handler == nil {
		return nil, ErrHandlerNotSetUp
	}
	return af.handler(kv, decoder, broker)
}
