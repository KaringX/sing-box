package wireguard

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"

	"github.com/sagernet/sing-box/option"
	tun "github.com/sagernet/sing-tun"

	"github.com/sagernet/sing-box/common/dialer"

	"github.com/sagernet/sing-box/service/powerreport"

	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
	"github.com/sagernet/wireguard-go/conn"
	"github.com/sagernet/wireguard-go/device"

	"go4.org/netipx"
)

type Endpoint struct {
	options        EndpointOptions
	peers          []peerConfig
	ipcConf        string
	allowedAddress []netip.Prefix
	tunDevice      Device
	returnDevice   *returnDeviceWrapper
	device         *device.Device
	allowedIPs     *device.AllowedIPs
	egressPool     *tun.UDPEgressPool
	pause          pause.Manager
	pauseCallback  *list.Element[pause.Callback]

	fakePackets         []int  //hiddify
	fakePacketsSize     []int  //hiddify
	fakePacketsDelay    []int  //hiddify
	fakePacketsHeader   []byte //hiddify
	fakePacketsNoModify bool   //hiddify
	parseErr            error  //karing
}

func NewEndpoint(options EndpointOptions) (*Endpoint, error) {
	//hiddify begin
	fakePackets := []int{0, 0}
	fakePacketsSize := []int{0, 0}
	fakePacketsDelay := []int{0, 0}
	fakePacketsHeader := []byte{}
	fakePacketsNoModify := false
	if options.FakePackets != "" {
		var err error
		fakePackets, err = option.ParseIntRange(options.FakePackets)
		if err != nil {
			return nil, err
		}
		fakePacketsSize = []int{40, 100}
		fakePacketsDelay = []int{10, 50}

		if options.FakePacketsSize != "" {
			var err error
			fakePacketsSize, err = option.ParseIntRange(options.FakePacketsSize)
			if err != nil {
				return nil, err
			}
		}

		if options.FakePacketsDelay != "" {
			var err error
			fakePacketsDelay, err = option.ParseIntRange(options.FakePacketsDelay)
			if err != nil {
				return nil, err
			}
		}
	}

	mode := strings.ToLower(options.FakePacketsMode)
	if mode == "" || mode == "m1" {
		fakePacketsHeader = []byte{}
		fakePacketsNoModify = false
	} else if mode == "m2" {
		fakePacketsHeader = []byte{}
		fakePacketsNoModify = true
	} else if mode == "m3" {
		// clist := []byte{0xC0, 0xC2, 0xC3, 0xC4, 0xC9, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF}
		fakePacketsHeader = []byte{0xDC, 0xDE, 0xD3, 0xD9, 0xD0, 0xEC, 0xEE, 0xE3}
		fakePacketsNoModify = false
	} else if mode == "m4" {
		fakePacketsHeader = []byte{0xDC, 0xDE, 0xD3, 0xD9, 0xD0, 0xEC, 0xEE, 0xE3}
		fakePacketsNoModify = true
	} else if mode == "m5" {
		fakePacketsHeader = []byte{0xC0, 0xC2, 0xC3, 0xC4, 0xC9, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF}
		fakePacketsNoModify = false
	} else if mode == "m6" {
		fakePacketsHeader = []byte{0x40, 0x42, 0x43, 0x44, 0x49, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F}
		fakePacketsNoModify = true
	} else if strings.HasPrefix(mode, "h") || strings.HasPrefix(mode, "g") {
		clist, err := hex.DecodeString(strings.ReplaceAll(mode[1:], "_", ""))
		if err != nil {
			return nil, E.Cause(err, "decode FakePacketsMode")
		}
		fakePacketsHeader = clist
		fakePacketsNoModify = strings.HasPrefix(mode, "h")
	} else {
		return nil, fmt.Errorf("incorrect packet mode: %s", mode)
	}
	//hiddify end
	if options.PrivateKey == "" {
		return nil, E.New("missing private key")
	}
	privateKeyBytes, err := base64.StdEncoding.DecodeString(options.PrivateKey)
	if err != nil {
		return nil, E.Cause(err, "decode private key")
	}
	privateKey := hex.EncodeToString(privateKeyBytes)
	ipcConf := "private_key=" + privateKey
	if options.ListenPort != 0 {
		ipcConf += "\nlisten_port=" + F.ToString(options.ListenPort)
	}
	var peers []peerConfig
	for peerIndex, rawPeer := range options.Peers {
		peer := peerConfig{
			allowedIPs: rawPeer.AllowedIPs,
			keepalive:  rawPeer.PersistentKeepaliveInterval,
		}
		if rawPeer.Endpoint.Addr.IsValid() {
			peer.endpoint = rawPeer.Endpoint.AddrPort()
		} else if rawPeer.Endpoint.IsDomain() {
			peer.destination = rawPeer.Endpoint
		}
		publicKeyBytes, err := base64.StdEncoding.DecodeString(rawPeer.PublicKey)
		if err != nil {
			return nil, E.Cause(err, "decode public key for peer ", peerIndex)
		}
		peer.publicKeyHex = hex.EncodeToString(publicKeyBytes)
		if rawPeer.PreSharedKey != "" {
			preSharedKeyBytes, err := base64.StdEncoding.DecodeString(rawPeer.PreSharedKey)
			if err != nil {
				return nil, E.Cause(err, "decode pre shared key for peer ", peerIndex)
			}
			peer.preSharedKeyHex = hex.EncodeToString(preSharedKeyBytes)
		}
		if len(rawPeer.AllowedIPs) == 0 {
			return nil, E.New("missing allowed ips for peer ", peerIndex)
		}
		if len(rawPeer.Reserved) > 0 {
			if len(rawPeer.Reserved) != 3 {
				return nil, E.New("invalid reserved value for peer ", peerIndex, ", required 3 bytes, got ", len(peer.reserved))
			}
			copy(peer.reserved[:], rawPeer.Reserved[:])
		}
		peers = append(peers, peer)
	}
	var allowedPrefixBuilder netipx.IPSetBuilder
	for _, peer := range options.Peers {
		for _, prefix := range peer.AllowedIPs {
			allowedPrefixBuilder.AddPrefix(prefix)
		}
	}
	allowedIPSet, err := allowedPrefixBuilder.IPSet()
	if err != nil {
		return nil, err
	}
	allowedAddresses := allowedIPSet.Prefixes()
	if options.MTU == 0 {
		options.MTU = 1408
	}
	deviceOptions := DeviceOptions{
		Context:         options.Context,
		Logger:          options.Logger,
		System:          options.System,
		Handler:         options.Handler,
		UDPTimeout:      options.UDPTimeout,
		ICMPTimeout:     options.ICMPTimeout,
		UDPMapping:      options.UDPMapping,
		UDPFiltering:    options.UDPFiltering,
		UDPNATMax:       options.UDPNATMax,
		InterfaceFinder: options.InterfaceFinder,
		CreateDialer:    options.CreateDialer,
		Name:            options.Name,
		MTU:             options.MTU,
		Address:         options.Address,
		AllowedAddress:  allowedAddresses,
	}
	tunDevice, err := NewDevice(deviceOptions)
	if err != nil {
		return nil, E.Cause(err, "create WireGuard device")
	}
	return &Endpoint{
		options:             options,
		peers:               peers,
		ipcConf:             ipcConf,
		allowedAddress:      allowedAddresses,
		tunDevice:           tunDevice,
		returnDevice:        &returnDeviceWrapper{Device: tunDevice},
		fakePackets:         fakePackets,         //hiddify
		fakePacketsSize:     fakePacketsSize,     //hiddify
		fakePacketsDelay:    fakePacketsDelay,    //hiddify
		fakePacketsHeader:   fakePacketsHeader,   //hiddify
		fakePacketsNoModify: fakePacketsNoModify, //hiddify
	}, nil
}

func (e *Endpoint) Start(postStart bool) error {
	return nil //karing
}

func (e *Endpoint) start() error {
	if e.parseErr != nil { //karing
		return e.parseErr
	}
	if e.device != nil { //karing
		return nil
	}
	/*//karing
	hasDomainPeer := common.Any(e.peers, func(peer peerConfig) bool {
		return peer.destination.IsDomain()
	})
	if postStart != hasDomainPeer {//karing
		return nil
	}
	*/
	var bind conn.Bind
	udpListener, isUDPListener := common.Cast[dialer.UDPListener](e.options.Dialer)
	if isUDPListener {
		listenerControl, egressEnabled := udpListener.UDPListenerControl()
		standardBind := conn.NewStdNetBind(listenerControl).(*conn.StdNetBind)
		if e.options.ListenPort == 0 && len(e.peers) == 1 && e.peers[0].endpoint.IsValid() {
			standardBind.SetSinglePeerMode()
		}
		if egressEnabled {
			egressPoolOptions := e.options.EgressPoolOptions
			egressPoolOptions.Control = listenerControl
			e.egressPool = tun.NewUDPEgressPool(egressPoolOptions)
			standardBind.SetEgressProvider(e.egressPool)
		}
		powerManager := service.FromContext[*powerreport.Manager](e.options.Context)
		if powerManager != nil {
			recorder := powerManager.Recorder()
			if recorder != nil {
				attribution := &powerreport.Attribution{Endpoint: e.options.Tag}
				standardBind.SetIOActivityFuncs(func(size int) {
					recorder.Touch(powerreport.DirectionInbound, size, attribution)
				}, func(size int) {
					recorder.Touch(powerreport.DirectionOutbound, size, attribution)
				})
			}
		}
		bind = standardBind
	} else {
		var (
			isConnect   bool
			connectAddr netip.AddrPort
			reserved    [3]uint8
		)
		if len(e.peers) == 1 {
			reserved = e.peers[0].reserved
			if e.peers[0].endpoint.IsValid() {
				isConnect = true
				connectAddr = e.peers[0].endpoint
			}
		}
		bind = NewClientBind(e.options.Context, e.options.Logger, e.options.Dialer, isConnect, connectAddr, reserved)
	}
	if isUDPListener || len(e.peers) > 1 {
		for _, peer := range e.peers {
			if peer.endpoint.IsValid() && peer.reserved != [3]uint8{} {
				bind.SetReservedForEndpoint(peer.endpoint, peer.reserved)
			}
		}
	}
	err := e.tunDevice.Start()
	if err != nil {
		e.parseErr = err //karing
		return err
	}
	logger := &device.Logger{
		Verbosef: func(format string, args ...any) {
			e.options.Logger.Debug(fmt.Sprintf(strings.ToLower(format), args...))
		},
		Errorf: func(format string, args ...any) {
			e.options.Logger.Error(fmt.Sprintf(strings.ToLower(format), args...))
		},
	}
	wgDevice := device.NewDevice(e.options.Context, e.returnDevice, bind, logger, e.options.Workers)
	wgDevice.FakePackets = e.fakePackets                 //hiddify
	wgDevice.FakePacketsSize = e.fakePacketsSize         //hiddify
	wgDevice.FakePacketsDelays = e.fakePacketsDelay      //hiddify
	wgDevice.FakePacketsHeader = e.fakePacketsHeader     //hiddify
	wgDevice.FakePacketsNoModify = e.fakePacketsNoModify //hiddify

	e.tunDevice.SetDevice(wgDevice)
	var ipcConf strings.Builder
	ipcConf.WriteString(e.ipcConf)
	for _, peer := range e.peers {
		ipcConf.WriteString(peer.GenerateIpcLines())
	}
	err = wgDevice.IpcSet(ipcConf.String())
	if err != nil {
		wgDevice.Close()
		e.parseErr = E.Cause(err, "setup wireguard: \n", ipcConf) //karing
		return E.Cause(err, "setup wireguard: \n", ipcConf.String())
	}
	for _, peer := range e.peers {
		if !peer.destination.IsDomain() {
			continue
		}
		var publicKey device.NoisePublicKey
		common.Must(publicKey.FromHex(peer.publicKeyHex))
		wgPeer, found := wgDevice.LookupActivePeer(publicKey)
		if !found {
			wgDevice.Close()
			return E.New("missing configured peer: ", peer.destination)
		}
		wgPeer.SetEndpointResolver(func() ([]conn.Endpoint, error) {
			addresses, lookupErr := e.options.ResolvePeer(peer.destination.Fqdn)
			if lookupErr != nil {
				return nil, lookupErr
			}
			endpoints := make([]conn.Endpoint, 0, len(addresses))
			for _, address := range addresses {
				destination := netip.AddrPortFrom(address, peer.destination.Port)
				if peer.reserved != ([3]uint8{}) {
					bind.SetReservedForEndpoint(destination, peer.reserved)
				}
				endpoint, parseErr := bind.ParseEndpoint(destination.String())
				if parseErr != nil {
					e.parseErr = E.Cause(parseErr, "setup wireguard endpoint: \n", destination.String()) //karing
					return nil, parseErr
				}
				endpoints = append(endpoints, endpoint)
			}
			return endpoints, nil
		})
	}
	e.device = wgDevice
	e.pause = service.FromContext[pause.Manager](e.options.Context)
	if e.pause != nil {
		e.pauseCallback = e.pause.RegisterCallback(e.onPauseUpdated)
	}
	e.allowedIPs = wgDevice.AllowedIPs()
	return nil
}

func (e *Endpoint) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if !destination.Addr.IsValid() {
		return nil, E.Cause(os.ErrInvalid, "invalid non-IP destination")
	}
	err := e.start() //karing
	if err != nil {  //karing
		return nil, err
	}
	return e.tunDevice.DialContext(ctx, network, destination)
}

func (e *Endpoint) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if !destination.Addr.IsValid() {
		return nil, E.Cause(os.ErrInvalid, "invalid non-IP destination")
	}
	err := e.start() //karing
	if err != nil {  //karing
		return nil, err
	}
	return e.tunDevice.ListenPacket(ctx, destination)
}

func (e *Endpoint) Close() error {
	if e.pauseCallback != nil {
		e.pause.UnregisterCallback(e.pauseCallback)
		e.pauseCallback = nil
	}
	if e.egressPool != nil {
		e.egressPool.Close()
		e.egressPool = nil
	}
	if e.device != nil {
		e.device.Down()
		e.device.Close()
		e.device = nil
		return nil
	}
	return e.tunDevice.Close()
}

func (e *Endpoint) Lookup(address netip.Addr) *device.Peer {
	if e.allowedIPs == nil {
		return nil
	}
	return e.allowedIPs.LookupFromPacket(netip.Addr{}, address, nil)
}

func (e *Endpoint) BindUpdate() error {
	if e.device == nil {
		return nil
	}
	return e.device.BindUpdate()
}

func (e *Endpoint) onPauseUpdated(event int) {
	if e.device == nil { //karing
		return
	}
	switch event {
	case pause.EventDevicePaused, pause.EventNetworkPause:
		e.device.Down()
	case pause.EventDeviceWake, pause.EventNetworkWake:
		e.device.Up()
	}
}

type peerConfig struct {
	destination     M.Socksaddr
	endpoint        netip.AddrPort
	publicKeyHex    string
	preSharedKeyHex string
	allowedIPs      []netip.Prefix
	keepalive       uint16
	reserved        [3]uint8
}

func (c peerConfig) GenerateIpcLines() string {
	var ipcLines strings.Builder
	ipcLines.WriteString("\npublic_key=" + c.publicKeyHex)
	if c.endpoint.IsValid() {
		ipcLines.WriteString("\nendpoint=" + c.endpoint.String())
	}
	if c.preSharedKeyHex != "" {
		ipcLines.WriteString("\npreshared_key=" + c.preSharedKeyHex)
	}
	for _, allowedIP := range c.allowedIPs {
		ipcLines.WriteString("\nallowed_ip=" + allowedIP.String())
	}
	if c.keepalive > 0 {
		ipcLines.WriteString("\npersistent_keepalive_interval=" + F.ToString(c.keepalive))
	}
	return ipcLines.String()
}
