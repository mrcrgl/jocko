package listoffset

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.OffsetsKey,
		scheme.APIVersion{protocol.OffsetsMinVersion, protocol.OffsetsMaxVersion},
		OffsetsHandler,
	)
}
