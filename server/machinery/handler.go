package machinery

import (
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko"
)

type Logger interface {
	Info(msg string, args... interface{}) error
	Debug(msg string, args... interface{}) error
}

type Handler func (kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error)
