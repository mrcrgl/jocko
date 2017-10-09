package produce

import (
	"time"

	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = ProduceHandler

func ProduceHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	req := new(ProduceRequest)

	if err := req.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	resp := new(ProduceResponses)
	resp.Responses = make([]*ProduceResponse, len(req.TopicData))
	for i, td := range req.TopicData {
		presps := make([]*ProducePartitionResponse, len(td.Data))
		for j, p := range td.Data {
			partition := jocko.NewPartition(td.Topic, p.Partition)
			presp := &ProducePartitionResponse{}
			partition, err := broker.Partition(td.Topic, p.Partition)
			if err != protocol.ErrNone {
				presp.ErrorCode = err.Code()
			}
			if !broker.IsLeaderOfPartition(partition.Topic, partition.ID, partition.LeaderID()) {
				presp.ErrorCode = protocol.ErrNotLeaderForPartition.Code()
				// break ?
			}
			offset, appendErr := partition.Append(p.RecordSet)
			if appendErr != nil {
				//s.logger.Info("commitlog/append failed: %s", err)
				presp.ErrorCode = protocol.ErrUnknown.Code()
			}
			presp.Partition = p.Partition
			presp.BaseOffset = offset
			presp.Timestamp = time.Now().Unix()
			presps[j] = presp
		}
		resp.Responses[i] = &ProduceResponse{
			Topic:              td.Topic,
			PartitionResponses: presps,
		}
	}

	return resp, nil
}
