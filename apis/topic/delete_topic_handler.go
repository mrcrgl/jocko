package topic

import (
	"fmt"

	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = DeleteTopicHandler

func DeleteTopicHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	reqs := new(DeleteTopicsRequest)

	if err := reqs.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	resp := new(DeleteTopicsResponse)
	resp.TopicErrorCodes = make([]*TopicErrorCode, len(reqs.Topics))
	isController := broker.IsController()

	var err error
	for i, topic := range reqs.Topics {
		if !isController {
			resp.TopicErrorCodes[i] = &TopicErrorCode{
				Topic:     topic,
				ErrorCode: protocol.ErrNotController.Code(),
			}
			continue
		}

		if err = broker.DeleteTopic(topic); err != nil {
			//s.logger.Info("failed to delete topic %s: %v", topic, err)
			return nil, fmt.Errorf("failed to delete topic %s: %v", topic, err)
		}
		resp.TopicErrorCodes[i] = &TopicErrorCode{
			Topic:     topic,
			ErrorCode: protocol.ErrNone.Code(),
		}
	}

	return resp, nil
}
