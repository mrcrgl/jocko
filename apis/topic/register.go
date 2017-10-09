package topic

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.CreateTopicsKey,
		scheme.APIVersion{protocol.CreateTopicsMinVersion, protocol.CreateTopicsMaxVersion},
		CreateTopicHandler,
	)
	api.Scheme.AddFunction(
		protocol.DeleteTopicsKey,
		scheme.APIVersion{protocol.DeleteTopicsMinVersion, protocol.DeleteTopicsMaxVersion},
		DeleteTopicHandler,
	)
}
