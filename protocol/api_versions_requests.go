package protocol

type APIVersionsRequest struct{}

func (c *APIVersionsRequest) Encode(_ PacketEncoder, _ int16) error {
	return nil
}

func (c *APIVersionsRequest) Decode(_ PacketDecoder, _ int16) error {
	return nil
}

func (c *APIVersionsRequest) Key() int16 {
	return APIVersionsKey
}

func (c *APIVersionsRequest) Version() int16 {
	return 0
}
