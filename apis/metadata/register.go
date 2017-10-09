package metadata

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.APIVersionsKey,
		scheme.APIVersion{protocol.APIVersionsMinVersion, protocol.APIVersionsMaxVersion},
		APIVersionsHandler,
	)

	api.Scheme.AddFunction(
		protocol.MetadataKey,
		scheme.APIVersion{protocol.MetadataMinVersion, protocol.MetadataMaxVersion},
		MetadataHandler,
	)
}
