package protocol

type ListGroupsRequest struct {
}

func (r *ListGroupsRequest) Encode(e PacketEncoder, _ int16) error {
	return nil
}

func (r *ListGroupsRequest) Decode(d PacketDecoder, _ int16) (err error) {
	return nil
}

func (r *ListGroupsRequest) Key() int16 {
	return ListGroupsKey
}

func (r *ListGroupsRequest) Version() int16 {
	return 0
}
