package listoffset

import "github.com/travisjeffery/jocko/protocol"

type PartitionResponse struct {
	Partition int32
	ErrorCode int16
	Timestamp int64
	Offsets   []int64
}

type OffsetResponse struct {
	Topic              string
	PartitionResponses []*PartitionResponse
}

type OffsetsResponse struct {
	Responses []*OffsetResponse
}

func (r *OffsetsResponse) Encode(e protocol.PacketEncoder, version int16) error {
	e.PutArrayLength(len(r.Responses))
	for _, r := range r.Responses {
		e.PutString(r.Topic)
		e.PutArrayLength(len(r.PartitionResponses))
		for _, p := range r.PartitionResponses {
			e.PutInt32(p.Partition)
			e.PutInt16(p.ErrorCode)
			if version == 1 {
				e.PutInt64(p.Timestamp)
			}
			e.PutInt64Array(p.Offsets)
		}
	}
	return nil
}

func (r *OffsetsResponse) Decode(d protocol.PacketDecoder, version int16) error {
	var err error
	l, err := d.ArrayLength()
	if err != nil {
		return err
	}
	r.Responses = make([]*OffsetResponse, l)
	for i := range r.Responses {
		resp := new(OffsetResponse)
		r.Responses[i] = resp
		resp.Topic, err = d.String()
		if err != nil {
			return err
		}
		pl, err := d.ArrayLength()
		if err != nil {
			return err
		}
		ps := make([]*PartitionResponse, pl)
		for j := range ps {
			p := new(PartitionResponse)
			p.Partition, err = d.Int32()
			if err != nil {
				return err
			}
			p.ErrorCode, err = d.Int16()
			if err != nil {
				return err
			}
			if version == 1 {
				p.Timestamp, err = d.Int64()
				if err != nil {
					return err
				}
			}

			p.Offsets, err = d.Int64Array()
			/*p.Offset, err = d.Int64()
			if err != nil {
				return err
			}*/
			ps[j] = p
		}
		resp.PartitionResponses = ps
	}
	if err != nil {
		return err
	}
	return nil
}