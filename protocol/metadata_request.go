package protocol

type MetadataRequest struct {
	Topics []string
}

func (r *MetadataRequest) Encode(e PacketEncoder, _ int16) error {
	e.PutStringArray(r.Topics)
	return nil
}

func (r *MetadataRequest) Decode(d PacketDecoder, _ int16) (err error) {
	r.Topics, err = d.StringArray()
	return err
}

func (r *MetadataRequest) Key() int16 {
	return MetadataKey
}

func (r *MetadataRequest) Version() int16 {
	return 0
}
