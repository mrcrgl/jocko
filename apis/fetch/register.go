package fetch

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.FetchKey,
		scheme.APIVersion{protocol.FetchMinVersion, protocol.FetchMaxVersion},
		FetchHandler,
	)
}
