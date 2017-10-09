package metadata

import "github.com/travisjeffery/jocko/protocol"

type MetadataRequest struct {
	Topics []string
}

func (r *MetadataRequest) Encode(e protocol.PacketEncoder, _ int16) error {
	e.PutStringArray(r.Topics)
	return nil
}

func (r *MetadataRequest) Decode(d protocol.PacketDecoder, _ int16) (err error) {
	r.Topics, err = d.StringArray()
	return err
}

func (r *MetadataRequest) Key() int16 {
	return protocol.MetadataKey
}

func (r *MetadataRequest) Version() int16 {
	return 0
}
