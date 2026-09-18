package trusttunnel

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/sagernet/sing/common/auth"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type Handler interface {
	N.TCPConnectionHandler
	N.UDPConnectionHandler
}

type ServiceOptions struct {
	Ctx     context.Context
	Logger  logger.ContextLogger
	Handler Handler
}

type Service struct {
	ctx     context.Context
	logger  logger.ContextLogger
	users   map[string]string
	handler Handler
	conns   map[string][]io.Closer

	mu sync.RWMutex
}

func NewService(options ServiceOptions) *Service {
	return &Service{
		ctx:     options.Ctx,
		logger:  options.Logger,
		handler: options.Handler,
		conns:   make(map[string][]io.Closer),
	}
}

func (s *Service) UpdateUsers(users map[string]string) {
	s.mu.Lock()
	s.users = users
	var closedConns []io.Closer
	for user, conns := range s.conns {
		if _, exists := users[user]; !exists {
			closedConns = append(closedConns, conns...)
			delete(s.conns, user)
		}
	}
	s.mu.Unlock()
	for _, conn := range closedConns {
		conn.Close()
	}
}

func (s *Service) trackConn(username string, conn io.Closer) {
	s.mu.Lock()
	s.conns[username] = append(s.conns[username], conn)
	s.mu.Unlock()
}

func (s *Service) untrackConn(username string, conn io.Closer) {
	s.mu.Lock()
	conns := s.conns[username]
	for i, c := range conns {
		if c == conn {
			s.conns[username] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	s.mu.Unlock()
}

func (s *Service) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	authorization := request.Header.Get("Proxy-Authorization")
	username, loaded := s.verify(authorization)
	if !loaded {
		writer.WriteHeader(http.StatusProxyAuthRequired)
		s.badRequest(request.Context(), request, E.New("authorization failed"))
		return
	}
	if request.Method != http.MethodConnect {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		s.badRequest(request.Context(), request, E.New("unexpected HTTP method ", request.Method))
		return
	}
	ctx := request.Context()
	ctx = auth.ContextWithUser(ctx, username)
	switch request.Host {
	case UDPMagicAddress:
		writer.WriteHeader(http.StatusOK)
		flusher, isFlusher := writer.(http.Flusher)
		if isFlusher {
			flusher.Flush()
		}
		done := make(chan struct{})
		conn := &serverPacketConn{
			httpConn: httpConn{
				writer:     writer,
				flusher:    flusher,
				created:    make(chan struct{}),
				done:       done,
				remoteAddr: parseRemoteAddr(request.RemoteAddr),
			},
		}
		conn.setup(request.Body, nil)
		firstPacket := buf.NewPacket()
		destination, err := conn.ReadPacket(firstPacket)
		if err != nil {
			firstPacket.Release()
			_ = conn.Close()
			s.logger.ErrorContext(ctx, E.Cause(err, "read first packet from ", request.RemoteAddr))
			return
		}
		destination = destination.Unwrap()
		cachedConn := bufio.NewCachedPacketConn(conn, firstPacket, destination)
		s.trackConn(username, conn)
		_ = s.handler.NewPacketConnection(ctx, cachedConn, M.Metadata{
			Protocol:    "trusttunnel",
			Source:      M.ParseSocksaddr(request.RemoteAddr),
			Destination: destination,
		})
		<-done
		s.untrackConn(username, conn)
	case HealthCheckMagicAddress:
		writer.WriteHeader(http.StatusOK)
		if flusher, isFlusher := writer.(http.Flusher); isFlusher {
			flusher.Flush()
		}
		_ = request.Body.Close()
	default:
		writer.WriteHeader(http.StatusOK)
		flusher, isFlusher := writer.(http.Flusher)
		if isFlusher {
			flusher.Flush()
		}
		done := make(chan struct{})
		conn := &tcpConn{
			httpConn{
				writer:     writer,
				flusher:    flusher,
				created:    make(chan struct{}),
				done:       done,
				remoteAddr: parseRemoteAddr(request.RemoteAddr),
			},
		}
		conn.setup(request.Body, nil)
		wrapper := &h2ConnWrapper{Conn: conn}
		s.trackConn(username, wrapper)
		_ = s.handler.NewConnection(ctx, wrapper, M.Metadata{
			Protocol:    "trusttunnel",
			Source:      M.ParseSocksaddr(request.RemoteAddr),
			Destination: M.ParseSocksaddr(request.Host).Unwrap(),
		})
		<-done
		s.untrackConn(username, wrapper)
		wrapper.CloseWrapper()
	}
}

func (s *Service) verify(authorization string) (username string, loaded bool) {
	username, password, loaded := parseBasicAuth(authorization)
	if !loaded {
		return "", false
	}
	s.mu.RLock()
	recordedPassword, loaded := s.users[username]
	s.mu.RUnlock()
	if !loaded {
		return "", false
	}
	if password != recordedPassword {
		return "", false
	}
	return username, true
}

func (s *Service) badRequest(ctx context.Context, request *http.Request, err error) {
	s.logger.ErrorContext(ctx, E.Cause(err, "process connection from ", request.RemoteAddr))
}

func parseRemoteAddr(addr string) net.Addr {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil
	}
	return tcpAddr
}

type h2ConnWrapper struct {
	net.Conn
	access sync.Mutex
	closed bool
}

func (w *h2ConnWrapper) Write(p []byte) (n int, err error) {
	w.access.Lock()
	defer w.access.Unlock()
	if w.closed {
		return 0, net.ErrClosed
	}
	return w.Conn.Write(p)
}

func (w *h2ConnWrapper) CloseWrapper() {
	w.access.Lock()
	defer w.access.Unlock()
	w.closed = true
}
