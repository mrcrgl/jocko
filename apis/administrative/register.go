package administrative

import (
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
	"github.com/travisjeffery/jocko/protocol"
)

func init() {
	api.Scheme.AddFunction(
		protocol.DescribeGroupsKey,
		scheme.APIVersion{protocol.DescribeGroupsMinVersion, protocol.DescribeGroupsMaxVersion},
		func() (protocol.ResponseBody, error) {
			return nil, nil
		},
	)
}
