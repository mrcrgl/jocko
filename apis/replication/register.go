package replication

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.LeaderAndISRKey,
		scheme.APIVersion{protocol.LeaderAndISRMinVersion, protocol.LeaderAndISRMaxVersion},
		LeaderAndISRHandler,
	)
}
