package metadata

import (
	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/protocol"
)

func APIVersionsHandler(_ protocol.APIKeyVersion, _ *protocol.ByteDecoder, _ jocko.Broker) (protocol.ResponseBody, error) {
	resp := new(APIVersionsResponse)

	resp.APIVersions = make([]APIVersion, api.Scheme.Len())

	var cur int
	for _, av := range api.Scheme.Map() {
		resp.APIVersions[cur] = APIVersion{
			APIKey:     av.Key,
			MinVersion: av.Version.Min(),
			MaxVersion: av.Version.Max(),
		}
		cur++
	}

	return resp, nil
}
