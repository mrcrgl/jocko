package groupmembership

type SyncGroupResponse struct {
	ErrorCode        int16
	MemberAssignment []byte
}

func (r *SyncGroupResponse) Encode(e PacketEncoder, _ int16) error {
	e.PutInt16(r.ErrorCode)
	return e.PutBytes(r.MemberAssignment)
}

func (r *SyncGroupResponse) Decode(d PacketDecoder, _ int16) (err error) {
	if r.ErrorCode, err = d.Int16(); err != nil {
		return err
	}
	r.MemberAssignment, err = d.Bytes()
	return err
}

func (r *SyncGroupResponse) Key() int16 {
	return 14
}

func (r *SyncGroupResponse) Version() int16 {
	return 0
}
