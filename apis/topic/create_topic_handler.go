package topic

import (
	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = CreateTopicHandler

func CreateTopicHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	reqs := new(CreateTopicRequests)

	if err := reqs.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	resp := new(CreateTopicsResponse)
	resp.TopicErrorCodes = make([]*TopicErrorCode, len(reqs.Requests))
	isController := broker.IsController()
	for i, req := range reqs.Requests {
		if !isController {
			resp.TopicErrorCodes[i] = &TopicErrorCode{
				Topic:     req.Topic,
				ErrorCode: protocol.ErrNotController.Code(),
			}
			continue
		}
		if req.ReplicationFactor > int16(len(broker.Cluster())) {
			resp.TopicErrorCodes[i] = &TopicErrorCode{
				Topic:     req.Topic,
				ErrorCode: protocol.ErrInvalidReplicationFactor.Code(),
			}
			continue
		}
		err := broker.CreateTopic(req.Topic, req.NumPartitions, req.ReplicationFactor)
		resp.TopicErrorCodes[i] = &TopicErrorCode{
			Topic:     req.Topic,
			ErrorCode: err.Code(),
		}
	}

	return resp, nil
}
