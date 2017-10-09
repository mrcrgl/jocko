package produce

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.ProduceKey,
		scheme.APIVersion{protocol.ProduceMinVersion, protocol.ProduceMaxVersion},
		ProduceHandler,
	)
}
