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
	"github.com/travisjeffery/jocko/api"
	"github.com/travisjeffery/jocko/api/scheme"
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
		if err := header.Decode(d); err != nil {
			// TODO handle this
			s.logger.Info("request header decoding failed: %s", err)
			break
		}
		s.logger.Debug("request: correlation id [%d], client id [%s], request size [%d], key [%d]", header.CorrelationID, header.ClientID, size, header.APIKey)

		af, err := api.Scheme.Lookup(header.APIKey, header.APIVersion)
		if err != nil {
			// TODO handle this
			break
		}

		respBody, err := af.Handle(header.KeyVersion(), d, s.broker)
		if err != nil {
			// TODO handle this
			break
		}

		// TODO put in method .writeResponse
		r := &protocol.Response{
			CorrelationID: header.CorrelationID,
			Body: respBody,
		}

		if err := s.write(conn, header, r); err != nil {
			// TODO handle this
		}

		//api := protocol.Register.Lookup(header.APIKey)
		//
		//if api.Supported == false {
		//	// TODO handle this
		//	s.logger.Info("Requested api key is not supported: api key [%d]", header.APIKey)
		//	break
		//}
		//
		//if !api.Version.Inherits(header.APIVersion) {
		//	// TODO handle this
		//	s.logger.Info("Requested api version is not supported: api key [%d], version [req: %d]", header.APIKey, header.APIVersion)
		//	break
		//}
		//
		//req := api.NewRequest()
		//if err := s.decode(header, req, d); err != nil {
		//	// TODO handle this
		//	s.logger.Info("Failed to decode incoming request: %s", err)
		//	break
		//}

		// Handler
		// -> Handle([header, version?], decoder, broker) ([Encoder, ResponseBody?], error) // everything is a decoder?

	}
}

func (s *Server) decode(header *protocol.RequestHeader, req protocol.Decoder, d protocol.PacketDecoder) error {
	err := req.Decode(d, header.APIVersion)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	b := new(jocko.ClusterMember)
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: change join to take a broker
	if err := s.broker.Join(b.IP); err != protocol.ErrNone {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

// Addr returns the address on which the Server is listening
func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
