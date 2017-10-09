package groupmembership

type HeartbeatResponse struct {
	ErrorCode int16
}

func (r *HeartbeatResponse) Encode(e PacketEncoder, _ int16) error {
	e.PutInt16(r.ErrorCode)
	return nil
}

func (r *HeartbeatResponse) Decode(d PacketDecoder, _ int16) (err error) {
	r.ErrorCode, err = d.Int16()
	return err
}

func (r *HeartbeatResponse) Key() int16 {
	return 12
}

func (r *HeartbeatResponse) Version() int16 {
	return 0
}
