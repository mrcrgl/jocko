package replication

import (
	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/jocko/server/machinery"
)

var _ machinery.Handler = LeaderAndISRHandler

func LeaderAndISRHandler(kv protocol.APIKeyVersion, decoder *protocol.ByteDecoder, broker jocko.Broker) (protocol.ResponseBody, error) {
	req := new(LeaderAndISRRequest)

	if err := req.Decode(decoder, kv.Version); err != nil {
		return nil, err
	}

	resp := &LeaderAndISRResponse{
		Partitions: make([]*LeaderAndISRPartition, len(req.PartitionStates)),
	}
	setErr := func(i int, p *PartitionState, err protocol.Error) {
		resp.Partitions[i] = &LeaderAndISRPartition{
			ErrorCode: err.Code(),
			Partition: p.Partition,
			Topic:     p.Topic,
		}
	}
	for i, p := range req.PartitionStates {
		partition, err := broker.Partition(p.Topic, p.Partition)
		// TODO: seems ok to have protocol.ErrUnknownTopicOrPartition here?
		if err != protocol.ErrNone {
			setErr(i, p, err)
			continue
		}
		if partition == nil {
			partition = &jocko.Partition{
				Topic:                   p.Topic,
				ID:                      p.Partition,
				Replicas:                p.Replicas,
				ISR:                     p.ISR,
				Leader:                  p.Leader,
				PreferredLeader:         p.Leader,
				LeaderAndISRVersionInZK: p.ZKVersion,
			}
			if err := broker.StartReplica(partition); err != protocol.ErrNone {
				setErr(i, p, err)
				continue
			}
		}
		if p.Leader == broker.ID() && !partition.IsLeader(broker.ID()) {
			// is command asking this broker to be the new leader for p and this broker is not already the leader for
			if err := broker.BecomeLeader(partition.Topic, partition.ID, p); err != protocol.ErrNone {
				setErr(i, p, err)
				continue
			}
		} else if contains(p.Replicas, broker.ID()) && !partition.IsFollowing(p.Leader) {
			// is command asking this broker to follow leader who it isn't a leader of already
			if err := broker.BecomeFollower(partition.Topic, partition.ID, p); err != protocol.ErrNone {
				setErr(i, p, err)
				continue
			}
		}
	}

	return resp, nil
}

func contains(rs []int32, r int32) bool {
	for _, ri := range rs {
		if ri == r {
			return true
		}
	}
	return false
}
