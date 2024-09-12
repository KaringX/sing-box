package outbound

import (
	"bytes"
	"context"
	"encoding/base64"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/dialer"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/uot" //hiddify

	"golang.org/x/crypto/ssh"
)

var (
	_ adapter.Outbound                = (*SSH)(nil)
	_ adapter.InterfaceUpdateListener = (*SSH)(nil)
)

type SSH struct {
	myOutboundAdapter
	ctx               context.Context
	dialer            N.Dialer
	serverAddr        M.Socksaddr
	user              string
	hostKey           []ssh.PublicKey
	hostKeyAlgorithms []string
	clientVersion     string
	authMethod        []ssh.AuthMethod
	clientAccess      sync.Mutex
	clientConn        net.Conn
	client            *ssh.Client
	uotClient         *uot.Client //hiddify
	parseErr          error                //karing
}

func NewSSH(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.SSHOutboundOptions) (*SSH, error) {
	empty := &SSH{  //karing
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeSSH,
			network:      []string{N.NetworkTCP},
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
	}
	outboundDialer, err := dialer.New(router, options.DialerOptions)
	if err != nil {
		return empty, err
	}
	outbound := &SSH{
		myOutboundAdapter: myOutboundAdapter{
			protocol:     C.TypeSSH,
			network:      []string{N.NetworkTCP},
			router:       router,
			logger:       logger,
			tag:          tag,
			dependencies: withDialerDependency(options.DialerOptions),
		},
		ctx:               ctx,
		dialer:            outboundDialer,
		serverAddr:        options.ServerOptions.Build(),
		user:              options.User,
		hostKeyAlgorithms: options.HostKeyAlgorithms,
		clientVersion:     options.ClientVersion,
	}
	if outbound.serverAddr.Port == 0 {
		outbound.serverAddr.Port = 22
	}
	if outbound.user == "" {
		outbound.user = "root"
	}
	if outbound.clientVersion == "" {
		outbound.clientVersion = randomVersion()
	}
	if options.Password != "" {
		outbound.authMethod = append(outbound.authMethod, ssh.Password(options.Password))
	}
	if len(options.PrivateKey) > 0 || options.PrivateKeyPath != "" {
		var privateKey []byte
		if len(options.PrivateKey) > 0 {
			privateKey = []byte(strings.Join(options.PrivateKey, "\n"))
		} else {
			var err error
			privateKey, err = os.ReadFile(os.ExpandEnv(options.PrivateKeyPath))
			if err != nil {
				return empty, E.Cause(err, "read private key") //karing
			}
		}
		var signer ssh.Signer
		var err error
		if options.PrivateKeyPassphrase == "" {
			signer, err = ssh.ParsePrivateKey(privateKey)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(options.PrivateKeyPassphrase))
		}
		if err != nil {
			return empty, E.Cause(err, "parse private key") //karing
		}
		outbound.authMethod = append(outbound.authMethod, ssh.PublicKeys(signer))
	}
	if len(options.HostKey) > 0 {
		for _, hostKey := range options.HostKey {
			key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
			if err != nil {
				return empty, E.New("parse host key ", key) //karing
			}
			outbound.hostKey = append(outbound.hostKey, key)
		}
	}
	uotOptions := common.PtrValueOrDefault(options.UDPOverTCP) //hiddify
	if uotOptions.Enabled {                                    //hiddify
		outbound.uotClient = &uot.Client{
			Dialer:  outbound,
			Version: uotOptions.Version,
		}
	}
	return outbound, nil
}

func randomVersion() string {
	version := "SSH-2.0-OpenSSH_"
	if rand.Intn(2) == 0 {
		version += "7." + strconv.Itoa(rand.Intn(10))
	} else {
		version += "8." + strconv.Itoa(rand.Intn(9))
	}
	return version
}

func (s *SSH) connect() (*ssh.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	s.clientAccess.Lock()
	defer s.clientAccess.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	conn, err := s.dialer.DialContext(s.ctx, N.NetworkTCP, s.serverAddr)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:              s.user,
		Auth:              s.authMethod,
		ClientVersion:     s.clientVersion,
		HostKeyAlgorithms: s.hostKeyAlgorithms,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if len(s.hostKey) == 0 {
				return nil
			}
			serverKey := key.Marshal()
			for _, hostKey := range s.hostKey {
				if bytes.Equal(serverKey, hostKey.Marshal()) {
					return nil
				}
			}
			return E.New("host key mismatch, server send ", key.Type(), " ", base64.StdEncoding.EncodeToString(serverKey))
		},
	}
	clientConn, chans, reqs, err := ssh.NewClientConn(conn, s.serverAddr.Addr.String(), config)
	if err != nil {
		conn.Close()
		return nil, E.Cause(err, "connect to ssh server")
	}

	client := ssh.NewClient(clientConn, chans, reqs)

	s.clientConn = conn
	s.client = client

	go func() {
		client.Wait()
		conn.Close()
		s.clientAccess.Lock()
		s.client = nil
		s.clientConn = nil
		s.clientAccess.Unlock()
	}()

	return client, nil
}

func (s *SSH) InterfaceUpdated() {
	common.Close(s.clientConn)
	return
}

func (s *SSH) Close() error {
	return common.Close(s.clientConn)
}

func (s *SSH) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if(s.parseErr != nil){ //karing
		return nil, s.parseErr
	}
	client, err := s.connect()
	if err != nil {
		return nil, err
	}
	switch N.NetworkName(network) { //hiddify
	case N.NetworkTCP:
		s.logger.InfoContext(ctx, "outbound connection to ", destination)
	case N.NetworkUDP:
		if s.uotClient != nil {
			s.logger.InfoContext(ctx, "outbound UoT connect packet connection to ", destination)
			return s.uotClient.DialContext(ctx, network, destination)
		}
		s.logger.InfoContext(ctx, "outbound packet connection to ", destination)
	default:
		return nil, E.Extend(N.ErrUnknownNetwork, network)
	}
	return client.Dial(network, destination.String())
}

func (s *SSH) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if(s.parseErr != nil){ //karing
		return nil, s.parseErr
	}
	return nil, os.ErrInvalid
}

func (s *SSH) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	if(s.parseErr != nil){ //karing
		return s.parseErr
	}
	return NewConnection(ctx, s, conn, metadata)
}

func (s *SSH) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	return os.ErrInvalid
}
func (s *SSH) SetParseErr(err error){ //karing
	s.parseErr = err
}