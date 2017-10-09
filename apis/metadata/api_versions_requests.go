package metadata

import "github.com/travisjeffery/jocko/protocol"

type APIVersionsRequest struct{}

func (c *APIVersionsRequest) Encode(_ protocol.PacketEncoder, _ int16) error {
	return nil
}

func (c *APIVersionsRequest) Decode(_ protocol.PacketDecoder, _ int16) error {
	return nil
}

func (c *APIVersionsRequest) Key() int16 {
	return protocol.APIVersionsKey
}

func (c *APIVersionsRequest) Version() int16 {
	return 0
}
