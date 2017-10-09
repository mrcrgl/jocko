package fetch

import (
	"io"
	"time"

	"bytes"

	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = FetchHandler

func FetchHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	req := new(FetchRequest)

	if err := req.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	resp := &FetchResponses{
		Responses: make([]*FetchResponse, len(req.Topics)),
	}
	received := time.Now()

	for i, topic := range req.Topics {
		fr := &FetchResponse{
			Topic:              topic.Topic,
			PartitionResponses: make([]*FetchPartitionResponse, len(topic.Partitions)),
		}

		for j, p := range topic.Partitions {
			partition, err := broker.Partition(topic.Topic, p.Partition)
			if err != protocol.ErrNone {
				fr.PartitionResponses[j] = &FetchPartitionResponse{
					Partition: p.Partition,
					ErrorCode: err.Code(),
				}
				continue
			}
			if !broker.IsLeaderOfPartition(partition.Topic, partition.ID, partition.LeaderID()) {
				fr.PartitionResponses[j] = &FetchPartitionResponse{
					Partition: p.Partition,
					ErrorCode: protocol.ErrNotLeaderForPartition.Code(),
				}
				continue
			}
			rdr, rdrErr := partition.NewReader(p.FetchOffset, p.MaxBytes)
			if rdrErr != nil {
				fr.PartitionResponses[j] = &FetchPartitionResponse{
					Partition: p.Partition,
					ErrorCode: protocol.ErrUnknown.Code(),
				}
				continue
			}
			b := new(bytes.Buffer)
			var n int32
			for n < req.MinBytes {
				if req.MaxWaitTime != 0 && int32(time.Since(received).Nanoseconds()/1e6) > req.MaxWaitTime {
					break
				}
				// TODO: copy these bytes to outer bytes
				nn, err := io.Copy(b, rdr)
				if err != nil && err != io.EOF {
					fr.PartitionResponses[j] = &FetchPartitionResponse{
						Partition: p.Partition,
						ErrorCode: protocol.ErrUnknown.Code(),
					}
					break
				}
				n += int32(nn)
				if err == io.EOF {
					break
				}
			}

			fr.PartitionResponses[j] = &FetchPartitionResponse{
				Partition:     p.Partition,
				ErrorCode:     protocol.ErrNone.Code(),
				HighWatermark: partition.HighWatermark(),
				RecordSet:     b.Bytes(),
			}
		}

		resp.Responses[i] = fr
	}

	return resp, nil
}
