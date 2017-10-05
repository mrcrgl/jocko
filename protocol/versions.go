package protocol

// Supported versions.

const (
	ProduceMaxVersion = 2
	ProduceMinVersion = 0

	FetchMaxVersion = 1
	FetchMinVersion = 0

	OffsetsMaxVersion = 1
	OffsetsMinVersion = 0

	MetadataMaxVersion = 0
	MetadataMinVersion = 0

	LeaderAndISRMaxVersion = 0
	LeaderAndISRMinVersion = 0

	StopReplicaMaxVersion = 0
	StopReplicaMinVersion = 0

	// TODO: add these when supported
	// UpdateMetadataMaxVersion = 0
	// UpdateMetadataMinVersion = 0

	// ControlledShutdownMaxVersion = 0
	// ControlledShutdownMinVersion = 0

	// OffsetCommitMaxVersion = 0
	// OffsetCommitMinVersion = 0

	// OffsetFetchMaxVersion = 0
	// OffsetFetchMinVersion = 0

	GroupCoordinatorMaxVersion = 0
	GroupCoordinatorMinVersion = 0

	JoinGroupMaxVersion = 1
	JoinGroupMinVersion = 0

	HeartbeatMaxVersion = 0
	HeartbeatMinVersion = 0

	LeaveGroupMaxVersion = 0
	LeaveGroupMinVersion = 0

	SyncGroupMaxVersion = 0
	SyncGroupMinVersion = 0

	DescribeGroupsMaxVersion = 0
	DescribeGroupsMinVersion = 0

	ListGroupsMaxVersion = 0
	ListGroupsMinVersion = 0

	SaslHandshakeMaxVersion = 0
	SaslHandshakeMinVersion = 0

	APIVersionsMaxVersion = 0
	APIVersionsMinVersion = 0

	CreateTopicsMaxVersion = 0
	CreateTopicsMinVersion = 0

	DeleteTopicsMaxVersion = 0
	DeleteTopicsMinVersion = 0
)

func supportVersion(version int16) (version1, version2, version3 bool) {
	switch version {
	case 0:
		break
	case 1:
		version1 = true
		break
	case 2:
		version2 = true
		break
	case 3:
		version3 = true
		break
	}

	return
}
