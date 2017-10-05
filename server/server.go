package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/travisjeffery/jocko"
	"github.com/travisjeffery/jocko/protocol"
	"github.com/travisjeffery/simplelog"
)

// Server is used to handle the TCP connections, decode requests,
// defer to the broker, and encode the responses.
type Server struct {
	addr       string
	ln         *net.TCPListener
	mu         sync.Mutex
	logger     *simplelog.Logger
	broker     jocko.Broker
	shutdownCh chan struct{}
}

// New creates a new Server instance.
func New(addr string, broker jocko.Broker, logger *simplelog.Logger) *Server {
	return &Server{
		addr:       addr,
		broker:     broker,
		logger:     logger,
		shutdownCh: make(chan struct{}),
	}
}

// Start starts the service.
func (s *Server) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		panic(err)
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	s.ln = ln

	r := mux.NewRouter()
	r.Methods("POST").Path("/join").HandlerFunc(s.handleJoin)
	r.PathPrefix("").HandlerFunc(s.handleNotFound)
	http.Handle("/", r)

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	server := http.Server{
		Handler: loggedRouter,
	}

	go func() {
		for {
			select {
			case <-s.shutdownCh:
				break
			default:
				conn, err := s.ln.Accept()
				if err != nil {
					s.logger.Debug("listener accept failed: %v", err)
					continue
				}

				go s.handleRequest(conn)
			}
		}
	}()

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			s.logger.Info("serve failed: %v", err)
		}
	}()

	return nil
}

// Close closes the service.
func (s *Server) Close() {
	close(s.shutdownCh)
	s.ln.Close()
	return
}

func (s *Server) handleRequest(conn net.Conn) {
	defer conn.Close()

	header := new(protocol.RequestHeader)
	p := make([]byte, 4)

	for {
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			s.logger.Info("read deadline failed: %s", err)
			continue
		}
		_, err = io.ReadFull(conn, p[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Info("conn read failed: %s", err)
			break
		}

		size := protocol.Encoding.Uint32(p)
		if size == 0 {
			break // TODO: should this even happen?
		}

		b := make([]byte, size+4) //+4 since we're going to copy the size into b
		copy(b, p)

		_, err = io.ReadFull(conn, b[4:])
		if err != nil {
			panic(err)
		}

		d := protocol.NewDecoder(b)
		header.Decode(d)
		s.logger.Debug("request: correlation id [%d], client id [%s], request size [%d], key [%d]", header.CorrelationID, header.ClientID, size, header.APIKey)

		switch header.APIKey {
		case protocol.APIVersionsKey:
			req := &protocol.APIVersionsRequest{}
			s.decode(header, req, d)
			if err = s.handleAPIVersions(conn, header, req); err != nil {
				s.logger.Info("api Versions failed: %s", err)
			}
		case protocol.ProduceKey:
			req := &protocol.ProduceRequest{}
			s.decodeVersioned(header, req, d)
			if err = s.handleProduce(conn, header, req); err != nil {
				s.logger.Info("produce failed: %s", err)
			}
		case protocol.FetchKey:
			req := &protocol.FetchRequest{}
			s.decodeVersioned(header, req, d)
			if err = s.handleFetch(conn, header, req); err != nil {
				s.logger.Info("fetch failed: %s", err)
			}
		case protocol.OffsetsKey:
			req := &protocol.OffsetsRequest{}
			s.decodeVersioned(header, req, d)
			if err = s.handleOffsets(conn, header, req); err != nil {
				s.logger.Info("offsets failed: %s", err)
			}
		case protocol.MetadataKey:
			req := &protocol.MetadataRequest{}
			s.decodeVersioned(header, req, d)
			if err = s.handleMetadata(conn, header, req); err != nil {
				s.logger.Info("metadata request failed: %s", err)
			}
		case protocol.CreateTopicsKey:
			req := &protocol.CreateTopicRequests{}
			s.decode(header, req, d)
			if err = s.handleCreateTopic(conn, header, req); err != nil {
				s.logger.Info("create topic failed: %s", err)
			}
		case protocol.DeleteTopicsKey:
			req := &protocol.DeleteTopicsRequest{}
			s.decode(header, req, d)
			if err = s.handleDeleteTopics(conn, header, req); err != nil {
				s.logger.Info("delete topic failed: %s", err)
			}
		case protocol.LeaderAndISRKey:
			req := &protocol.LeaderAndISRRequest{}
			s.decodeVersioned(header, req, d)
			if err = s.handleLeaderAndISR(conn, header, req); err != nil {
				s.logger.Info("handle leader and ISR failed: %s", err)
			}
		}
	}
}

func (s *Server) decode(header *protocol.RequestHeader, req protocol.Decoder, d protocol.PacketDecoder) error {
	err := req.Decode(d, header.APIVersion)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) decodeVersioned(header *protocol.RequestHeader, req protocol.Decoder, d protocol.PacketDecoder) error {
	err := req.Decode(d, header.APIVersion)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleAPIVersions(conn net.Conn, header *protocol.RequestHeader, req *protocol.APIVersionsRequest) (err error) {
	resp := new(protocol.APIVersionsResponse)

	resp.APIVersions = []protocol.APIVersion{
		{APIKey: protocol.ProduceKey, MinVersion: protocol.ProduceMinVersion, MaxVersion: protocol.ProduceMaxVersion},
		{APIKey: protocol.FetchKey, MinVersion: protocol.FetchMinVersion, MaxVersion: protocol.FetchMaxVersion},
		{APIKey: protocol.OffsetsKey, MinVersion: protocol.OffsetsMinVersion, MaxVersion: protocol.OffsetsMaxVersion},
		{APIKey: protocol.MetadataKey, MinVersion: protocol.MetadataMinVersion, MaxVersion: protocol.MetadataMaxVersion},
		{APIKey: protocol.LeaderAndISRKey, MinVersion: protocol.LeaderAndISRMinVersion, MaxVersion: protocol.LeaderAndISRMaxVersion},
		{APIKey: protocol.StopReplicaKey, MinVersion: protocol.StopReplicaMinVersion, MaxVersion: protocol.StopReplicaMaxVersion},
		{APIKey: protocol.GroupCoordinatorKey, MinVersion: protocol.GroupCoordinatorMinVersion, MaxVersion: protocol.GroupCoordinatorMaxVersion},
		{APIKey: protocol.JoinGroupKey, MinVersion: protocol.JoinGroupMinVersion, MaxVersion: protocol.JoinGroupMaxVersion},
		{APIKey: protocol.HeartbeatKey, MinVersion: protocol.HeartbeatMinVersion, MaxVersion: protocol.HeartbeatMaxVersion},
		{APIKey: protocol.LeaveGroupKey, MinVersion: protocol.LeaveGroupMinVersion, MaxVersion: protocol.LeaveGroupMaxVersion},
		{APIKey: protocol.SyncGroupKey, MinVersion: protocol.SyncGroupMinVersion, MaxVersion: protocol.SyncGroupMaxVersion},
		{APIKey: protocol.DescribeGroupsKey, MinVersion: protocol.DescribeGroupsMinVersion, MaxVersion: protocol.DescribeGroupsMaxVersion},
		{APIKey: protocol.ListGroupsKey, MinVersion: protocol.ListGroupsMinVersion, MaxVersion: protocol.ListGroupsMaxVersion},
		{APIKey: protocol.APIVersionsKey, MinVersion: protocol.APIVersionsMinVersion, MaxVersion: protocol.APIVersionsMaxVersion},
		{APIKey: protocol.CreateTopicsKey, MinVersion: protocol.CreateTopicsMinVersion, MaxVersion: protocol.CreateTopicsMaxVersion},
		{APIKey: protocol.DeleteTopicsKey, MinVersion: protocol.DeleteTopicsMinVersion, MaxVersion: protocol.DeleteTopicsMaxVersion},
	}

	r := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          resp,
	}
	return s.write(conn, header, r)
}

func (s *Server) handleCreateTopic(conn net.Conn, header *protocol.RequestHeader, reqs *protocol.CreateTopicRequests) error {
	resp := new(protocol.CreateTopicsResponse)
	resp.TopicErrorCodes = make([]*protocol.TopicErrorCode, len(reqs.Requests))
	isController := s.broker.IsController()
	if isController {
		for i, req := range reqs.Requests {
			if req.ReplicationFactor > int16(len(s.broker.Cluster())) {
				resp.TopicErrorCodes[i] = &protocol.TopicErrorCode{
					Topic:     req.Topic,
					ErrorCode: protocol.ErrInvalidReplicationFactor.Code(),
				}
				continue
			}
			err := s.broker.CreateTopic(req.Topic, req.NumPartitions, req.ReplicationFactor)
			resp.TopicErrorCodes[i] = &protocol.TopicErrorCode{
				Topic:     req.Topic,
				ErrorCode: err.Code(),
			}
		}
	} else {
		// TODO: forward req to controller
		s.logger.Info("failed to create topic(s): %s", protocol.ErrNotController)
		// TODO: could have these topic error code structs have protocol.Error
		// set as the field instead of the code directly
		for i, req := range reqs.Requests {
			resp.TopicErrorCodes[i] = &protocol.TopicErrorCode{
				Topic:     req.Topic,
				ErrorCode: protocol.ErrNotController.Code(),
			}
		}
	}
	r := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          resp,
	}
	return s.write(conn, header, r)
}

func (s *Server) handleDeleteTopics(conn net.Conn, header *protocol.RequestHeader, reqs *protocol.DeleteTopicsRequest) (err error) {
	resp := new(protocol.DeleteTopicsResponse)
	resp.TopicErrorCodes = make([]*protocol.TopicErrorCode, len(reqs.Topics))
	isController := s.broker.IsController()
	for i, topic := range reqs.Topics {
		if !isController {
			resp.TopicErrorCodes[i] = &protocol.TopicErrorCode{
				Topic:     topic,
				ErrorCode: protocol.ErrNotController.Code(),
			}
			continue
		}
		err = s.broker.DeleteTopic(topic)
		if err != nil {
			s.logger.Info("failed to delete topic %s: %v", topic, err)
			return
		}
		resp.TopicErrorCodes[i] = &protocol.TopicErrorCode{
			Topic:     topic,
			ErrorCode: protocol.ErrNone.Code(),
		}
	}
	r := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          resp,
	}
	return s.write(conn, header, r)
}

func (s *Server) handleLeaderAndISR(conn net.Conn, header *protocol.RequestHeader, req *protocol.LeaderAndISRRequest) (err error) {
	body := &protocol.LeaderAndISRResponse{}
	for _, p := range req.PartitionStates {
		partition, err := s.broker.Partition(p.Topic, p.Partition)
		if err != protocol.ErrNone {
			return err
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
			if err := s.broker.StartReplica(partition); err != nil {
				return err
			}
		}
		if p.Leader == s.broker.ID() && !partition.IsLeader(s.broker.ID()) {
			// is command asking this broker to be the new leader for p and this broker is not already the leader for
			if err := s.broker.BecomeLeader(partition.Topic, partition.ID, p); err != nil {
				return err
			}
		} else if contains(p.Replicas, s.broker.ID()) && !partition.IsFollowing(p.Leader) {
			// is command asking this broker to follow leader who it isn't a leader of already
			if err := s.broker.BecomeFollower(partition.Topic, partition.ID, p); err != nil {
				return err
			}
		}
	}
	r := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          body,
	}
	return s.write(conn, header, r)
}

func contains(rs []int32, r int32) bool {
	for _, ri := range rs {
		if ri == r {
			return true
		}
	}
	return false
}

func zero(p []byte) {
	for i := range p {
		p[i] = 0
	}
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	b := new(jocko.ClusterMember)
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: change join to take a broker
	if _, err := s.broker.Join(b.IP); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleMetadata(conn net.Conn, header *protocol.RequestHeader, req *protocol.MetadataRequest) error {
	brokers := make([]*protocol.Broker, 0, len(s.broker.Cluster()))
	for _, b := range s.broker.Cluster() {
		brokers = append(brokers, &protocol.Broker{
			NodeID: b.ID,
			Host:   b.IP,
			Port:   int32(b.Port),
		})
	}
	var topicMetadata []*protocol.TopicMetadata
	topicMetadataFn := func(topic string, partitions []*jocko.Partition, err protocol.Error) *protocol.TopicMetadata {
		partitionMetadata := make([]*protocol.PartitionMetadata, len(partitions))
		for i, p := range partitions {
			partitionMetadata[i] = &protocol.PartitionMetadata{
				ParititionID: p.ID,
			}
		}
		return &protocol.TopicMetadata{
			TopicErrorCode:    err.Code(),
			Topic:             topic,
			PartitionMetadata: partitionMetadata,
		}
	}
	if len(req.Topics) == 0 {
		// Respond with metadata for all topics
		topics := s.broker.Topics()
		topicMetadata = make([]*protocol.TopicMetadata, len(topics))
		idx := 0
		for topic, partitions := range topics {
			topicMetadata[idx] = topicMetadataFn(topic, partitions, protocol.ErrNone)
			idx++
		}
	} else {
		topicMetadata = make([]*protocol.TopicMetadata, len(req.Topics))
		for i, topic := range req.Topics {
			partitions, err := s.broker.TopicPartitions(topic)
			topicMetadata[i] = topicMetadataFn(topic, partitions, err)
		}
	}
	resp := &protocol.MetadataResponse{
		Brokers:       brokers,
		TopicMetadata: topicMetadata,
	}
	r := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          resp,
	}
	return s.write(conn, header, r)
}

func (s *Server) write(conn net.Conn, header *protocol.RequestHeader, e protocol.Encoder) error {
	s.logger.Debug("response: correlation id [%d], key [%d]", header.CorrelationID, header.APIKey)
	b, err := protocol.EncodeVersioned(e, header.APIVersion)
	if err != nil {
		return err
	}
	_, err = conn.Write(b)
	return err
}

func (s *Server) handleOffsets(conn net.Conn, header *protocol.RequestHeader, req *protocol.OffsetsRequest) error {
	oResp := new(protocol.OffsetsResponse)
	oResp.Responses = make([]*protocol.OffsetResponse, len(req.Topics))
	for i, t := range req.Topics {
		oResp.Responses[i] = new(protocol.OffsetResponse)
		oResp.Responses[i].Topic = t.Topic
		oResp.Responses[i].PartitionResponses = make([]*protocol.PartitionResponse, len(t.Partitions))
		for j, p := range t.Partitions {
			pResp := new(protocol.PartitionResponse)
			pResp.Partition = p.Partition

			partition, err := s.broker.Partition(t.Topic, p.Partition)
			if err != protocol.ErrNone {
				pResp.ErrorCode = err.Code()
				continue
			}

			var offset int64
			if p.Timestamp == -2 {
				offset = partition.LowWatermark()
			} else {
				offset = partition.HighWatermark()
			}
			pResp.Offsets = []int64{offset}

			oResp.Responses[i].PartitionResponses[j] = pResp
		}
	}
	resp := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          oResp,
	}
	return s.write(conn, header, resp)
}

func (s *Server) handleProduce(conn net.Conn, header *protocol.RequestHeader, req *protocol.ProduceRequest) error {
	resp := new(protocol.ProduceResponses)
	resp.Responses = make([]*protocol.ProduceResponse, len(req.TopicData))
	for i, td := range req.TopicData {
		presps := make([]*protocol.ProducePartitionResponse, len(td.Data))
		for j, p := range td.Data {
			partition := jocko.NewPartition(td.Topic, p.Partition)
			presp := &protocol.ProducePartitionResponse{}
			partition, err := s.broker.Partition(td.Topic, p.Partition)
			if err != protocol.ErrNone {
				presp.ErrorCode = err.Code()
			}
			if !s.broker.IsLeaderOfPartition(partition.Topic, partition.ID, partition.LeaderID()) {
				presp.ErrorCode = protocol.ErrNotLeaderForPartition.Code()
				// break ?
			}
			offset, appendErr := partition.Append(p.RecordSet)
			if appendErr != nil {
				s.logger.Info("commitlog/append failed: %s", err)
				presp.ErrorCode = protocol.ErrUnknown.Code()
			}
			presp.Partition = p.Partition
			presp.BaseOffset = offset
			presp.Timestamp = time.Now().Unix()
			presps[j] = presp
		}
		resp.Responses[i] = &protocol.ProduceResponse{
			Topic:              td.Topic,
			PartitionResponses: presps,
		}
	}
	r := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          resp,
	}
	return s.write(conn, header, r)
}

func (s *Server) handleFetch(conn net.Conn, header *protocol.RequestHeader, r *protocol.FetchRequest) error {
	fresp := &protocol.FetchResponses{
		Responses: make([]*protocol.FetchResponse, len(r.Topics)),
	}
	received := time.Now()

	for i, topic := range r.Topics {
		fr := &protocol.FetchResponse{
			Topic:              topic.Topic,
			PartitionResponses: make([]*protocol.FetchPartitionResponse, len(topic.Partitions)),
		}

		for j, p := range topic.Partitions {
			partition, err := s.broker.Partition(topic.Topic, p.Partition)
			if err != protocol.ErrNone {
				fr.PartitionResponses[j] = &protocol.FetchPartitionResponse{
					Partition: p.Partition,
					ErrorCode: err.Code(),
				}
				continue
			}
			if !s.broker.IsLeaderOfPartition(partition.Topic, partition.ID, partition.LeaderID()) {
				fr.PartitionResponses[j] = &protocol.FetchPartitionResponse{
					Partition: p.Partition,
					ErrorCode: protocol.ErrNotLeaderForPartition.Code(),
				}
				continue
			}
			rdr, rdrErr := partition.NewReader(p.FetchOffset, p.MaxBytes)
			if rdrErr != nil {
				fr.PartitionResponses[j] = &protocol.FetchPartitionResponse{
					Partition: p.Partition,
					ErrorCode: protocol.ErrUnknown.Code(),
				}
				continue
			}
			b := new(bytes.Buffer)
			var n int32
			for n < r.MinBytes {
				if r.MaxWaitTime != 0 && int32(time.Since(received).Nanoseconds()/1e6) > r.MaxWaitTime {
					break
				}
				// TODO: copy these bytes to outer bytes
				nn, err := io.Copy(b, rdr)
				if err != nil && err != io.EOF {
					fr.PartitionResponses[j] = &protocol.FetchPartitionResponse{
						Partition: p.Partition,
						ErrorCode: protocol.ErrUnknown.Code(),
					}
					break
				}
				n += int32(nn)
				if err == io.EOF {
					break
				}
			}

			fr.PartitionResponses[j] = &protocol.FetchPartitionResponse{
				Partition:     p.Partition,
				ErrorCode:     protocol.ErrNone.Code(),
				HighWatermark: partition.HighWatermark(),
				RecordSet:     b.Bytes(),
			}
		}

		fresp.Responses[i] = fr
	}

	resp := &protocol.Response{
		CorrelationID: header.CorrelationID,
		Body:          fresp,
	}

	return s.write(conn, header, resp)
}

// Addr returns the address on which the Server is listening
func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
