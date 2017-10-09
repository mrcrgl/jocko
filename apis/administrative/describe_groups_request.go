package administrative

type DescribeGroupsRequest struct {
	GroupIDs []string
}

func (r *DescribeGroupsRequest) Encode(e PacketEncoder, _ int16) error {
	return e.PutStringArray(r.GroupIDs)
}

func (r *DescribeGroupsRequest) Decode(d PacketDecoder, _ int16) (err error) {
	r.GroupIDs, err = d.StringArray()
	return err
}

func (r *DescribeGroupsRequest) Key() int16 {
	return DescribeGroupsKey
}

func (r *DescribeGroupsRequest) Version() int16 {
	return 0
}
