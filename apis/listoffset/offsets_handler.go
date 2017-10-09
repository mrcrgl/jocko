package listoffset

import (
	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = OffsetsHandler

func OffsetsHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	req := new(OffsetsRequest)

	if err := req.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	resp := new(OffsetsResponse)
	resp.Responses = make([]*OffsetResponse, len(req.Topics))
	for i, t := range req.Topics {
		resp.Responses[i] = new(OffsetResponse)
		resp.Responses[i].Topic = t.Topic
		resp.Responses[i].PartitionResponses = make([]*PartitionResponse, len(t.Partitions))
		for j, p := range t.Partitions {
			pResp := new(PartitionResponse)
			pResp.Partition = p.Partition

			partition, err := broker.Partition(t.Topic, p.Partition)
			if err != protocol.ErrNone {
				pResp.ErrorCode = err.Code()
				continue
			}

			var offset int64
			if p.Timestamp == -2 {
				offset = partition.LowWatermark()
			} else {
				offset = partition.HighWatermark()
			}
			pResp.Offsets = []int64{offset}

			resp.Responses[i].PartitionResponses[j] = pResp
		}
	}

	return resp, nil
}
