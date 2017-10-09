package metadata

import (
	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = MetadataHandler

func MetadataHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	req := new(MetadataRequest)

	if err := req.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	cluster := broker.Cluster()
	brokers := make([]*Broker, 0, len(cluster))
	for _, b := range cluster {
		brokers = append(brokers, &Broker{
			NodeID: b.ID,
			Host:   b.IP,
			Port:   int32(b.Port),
		})
	}
	var topicMetadata []*TopicMetadata
	topicMetadataFn := func(topic string, partitions []*jocko.Partition, err protocol.Error) *TopicMetadata {
		partitionMetadata := make([]*PartitionMetadata, len(partitions))
		for i, p := range partitions {
			partitionMetadata[i] = &PartitionMetadata{
				ParititionID: p.ID,
			}
		}
		return &TopicMetadata{
			TopicErrorCode:    err.Code(),
			Topic:             topic,
			PartitionMetadata: partitionMetadata,
		}
	}
	if len(req.Topics) == 0 {
		// Respond with metadata for all topics
		topics := broker.Topics()
		topicMetadata = make([]*TopicMetadata, len(topics))
		idx := 0
		for topic, partitions := range topics {
			topicMetadata[idx] = topicMetadataFn(topic, partitions, protocol.ErrNone)
			idx++
		}
	} else {
		topicMetadata = make([]*TopicMetadata, len(req.Topics))
		for i, topic := range req.Topics {
			partitions, err := broker.TopicPartitions(topic)
			topicMetadata[i] = topicMetadataFn(topic, partitions, err)
		}
	}
	resp := &MetadataResponse{
		Brokers:       brokers,
		TopicMetadata: topicMetadata,
	}

	return resp, nil
}
