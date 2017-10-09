package offset

type GroupCoordinatorRequest struct {
	GroupID string
}

func (r *GroupCoordinatorRequest) Encode(e PacketEncoder, _ int16) error {
	return e.PutString(r.GroupID)
}

func (r *GroupCoordinatorRequest) Decode(d PacketDecoder, _ int16) (err error) {
	r.GroupID, err = d.String()
	return err
}

func (r *GroupCoordinatorRequest) Version() int16 {
	return 0
}

func (r *GroupCoordinatorRequest) Key() int16 {
	return GroupCoordinatorKey
}