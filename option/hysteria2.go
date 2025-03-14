package option

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	"github.com/sagernet/sing/common/json/badoption"
)

type Hysteria2InboundOptions struct {
	ListenOptions
	UpMbps                int             `json:"up_mbps,omitempty"`
	DownMbps              int             `json:"down_mbps,omitempty"`
	Obfs                  *Hysteria2Obfs  `json:"obfs,omitempty"`
	Users                 []Hysteria2User `json:"users,omitempty"`
	IgnoreClientBandwidth bool            `json:"ignore_client_bandwidth,omitempty"`
	InboundTLSOptionsContainer
	Masquerade  *Hysteria2Masquerade `json:"masquerade,omitempty"`
	BrutalDebug bool                 `json:"brutal_debug,omitempty"`
}

type Hysteria2Obfs struct {
	Type     string `json:"type,omitempty"`
	Password string `json:"password,omitempty"`
}

type Hysteria2User struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

type _Hysteria2Masquerade struct {
	Type          string                    `json:"type,omitempty"`
	FileOptions   Hysteria2MasqueradeFile   `json:"-"`
	ProxyOptions  Hysteria2MasqueradeProxy  `json:"-"`
	StringOptions Hysteria2MasqueradeString `json:"-"`
}

type Hysteria2Masquerade _Hysteria2Masquerade

func (m Hysteria2Masquerade) MarshalJSON() ([]byte, error) {
	var v any
	switch m.Type {
	case C.Hysterai2MasqueradeTypeFile:
		v = m.FileOptions
	case C.Hysterai2MasqueradeTypeProxy:
		v = m.ProxyOptions
	case C.Hysterai2MasqueradeTypeString:
		v = m.StringOptions
	default:
		return nil, E.New("unknown masquerade type: ", m.Type)
	}
	return badjson.MarshallObjects((_Hysteria2Masquerade)(m), v)
}

func (m *Hysteria2Masquerade) UnmarshalJSON(bytes []byte) error {
	var urlString string
	err := json.Unmarshal(bytes, &urlString)
	if err == nil {
		masqueradeURL, err := url.Parse(urlString)
		if err != nil {
			return E.Cause(err, "invalid masquerade URL")
		}
		switch masqueradeURL.Scheme {
		case "file":
			m.Type = C.Hysterai2MasqueradeTypeFile
			m.FileOptions.Directory = masqueradeURL.Path
		case "http", "https":
			m.Type = C.Hysterai2MasqueradeTypeProxy
			m.ProxyOptions.URL = urlString
		default:
			return E.New("unknown masquerade URL scheme: ", masqueradeURL.Scheme)
		}
		return nil
	}
	err = json.Unmarshal(bytes, (*_Hysteria2Masquerade)(m))
	if err != nil {
		return err
	}
	var v any
	switch m.Type {
	case C.Hysterai2MasqueradeTypeFile:
		v = &m.FileOptions
	case C.Hysterai2MasqueradeTypeProxy:
		v = &m.ProxyOptions
	case C.Hysterai2MasqueradeTypeString:
		v = &m.StringOptions
	default:
		return E.New("unknown masquerade type: ", m.Type)
	}
	return badjson.UnmarshallExcluded(bytes, (*_Hysteria2Masquerade)(m), v)
}

type Hysteria2MasqueradeFile struct {
	Directory string `json:"directory"`
}

type Hysteria2MasqueradeProxy struct {
	URL         string `json:"url"`
	RewriteHost bool   `json:"rewrite_host,omitempty"`
}

type Hysteria2MasqueradeString struct {
	StatusCode int                  `json:"status_code,omitempty"`
	Headers    badoption.HTTPHeader `json:"headers,omitempty"`
	Content    string               `json:"content"`
}

type Hysteria2OutboundOptions struct {
	DialerOptions
	ServerOptions
	ServerPorts badoption.Listable[string] `json:"server_ports,omitempty"`
	HopInterval HopIntervalValue           `json:"hop_interval,omitempty"` //karing
	UpMbps      int                        `json:"up_mbps,omitempty"`
	DownMbps    int                        `json:"down_mbps,omitempty"`
	Obfs        *Hysteria2Obfs             `json:"obfs,omitempty"`
	Password    string                     `json:"password,omitempty"`
	Network     NetworkList                `json:"network,omitempty"`
	OutboundTLSOptionsContainer
	BrutalDebug bool `json:"brutal_debug,omitempty"`
	TurnRelay   *TurnRelayOptions `json:"turn_relay,omitempty"` //hiddify
	HopPorts    HopPortsValue     `json:"hop_ports,omitempty"`  //https://github.com/morgenanno/sing-box //"114,514,810-1919"
	//HopInterval int   //`json:"hop_interval,omitempty"`  //https://github.com/morgenanno/sing-box
}

type HopPortsValue string //karing
type HopIntervalValue badoption.Duration //karing

type _Hysteria2OutboundOptions Hysteria2OutboundOptions //karing
func (m *Hysteria2OutboundOptions) UnmarshalJSON(bytes []byte) error { //karing
	err := json.Unmarshal(bytes, (*_Hysteria2OutboundOptions)(m))
	if err != nil {
		return err
	}
    
	if len(m.ServerPorts) == 0 && len(m.HopPorts) > 0 {
		ports := strings.Split(string(m.HopPorts), ",")
		for i := 0; i < len(ports); i++ {
			ports[i] = strings.Replace(ports[i], "-", ":", -1)
			parts := strings.Split(ports[i], ":")
			if len(parts) == 1 {
				m.ServerPorts = append(m.ServerPorts, fmt.Sprintf("%s:%s", parts[0], parts[0]))
			} else if len (parts) == 2 {
				m.ServerPorts = append(m.ServerPorts, fmt.Sprintf("%s:%s", parts[0], parts[1]))
			} else {
				return E.New("invalid hop_ports format: ", string(m.HopPorts))
			}
		}
	}
	m.HopPorts = ""
    return nil
}
func (m *HopIntervalValue) UnmarshalJSON(bytes []byte) error { //karing
	var stringValue string
    if err := json.Unmarshal(bytes, &stringValue); err == nil {
        duration, err := time.ParseDuration(stringValue)
        if err != nil {
            return fmt.Errorf("invalid duration string: %w", err)
        }
        *m = HopIntervalValue(duration)
        return nil
    }
	var intValue int64
	if err := json.Unmarshal(bytes, &intValue); err == nil {
		if intValue < 0 {
            return fmt.Errorf("negative duration not allowed")
        }
        if intValue > math.MaxInt64/int64(time.Second) {
            return fmt.Errorf("integer overflow for duration")
        }

        *m = HopIntervalValue(time.Duration(intValue) * time.Second)
		return nil
	}
	 
    return nil
}
