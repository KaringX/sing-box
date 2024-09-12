package option

import (
	"github.com/sagernet/sing/common/json"
)

type _Options struct {
	RawMessage   json.RawMessage         `json:"-"`
	Schema       string                  `json:"$schema,omitempty"`
	Log          *LogOptions             `json:"log,omitempty"`
	DNS          *DNSOptions             `json:"dns,omitempty"`
	NTP          *NTPOptions             `json:"ntp,omitempty"`
	Inbounds     []Inbound               `json:"inbounds,omitempty"`
	Outbounds    []Outbound              `json:"outbounds,omitempty"`
	Route        *RouteOptions           `json:"route,omitempty"`
	Experimental *ExperimentalOptions    `json:"experimental,omitempty"`
	Custom       *map[string]interface{} `json:"custom,omitempty"`
}

type Options _Options

func (o *Options) UnmarshalJSON(content []byte) error {
	return o.UnmarshalFastJSON(content) //karing
	/*decoder := json.NewDecoder(json.NewCommentFilter(bytes.NewReader(content)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode((*_Options)(o))
	if err != nil {
		return err
	}
	var options Options
	options.UnmarshalFastJSON(content)
	if !reflect.DeepEqual(&options, o) {//test equal
		panic("Options not equal.")
	}
	o.RawMessage = content

	return nil*/
}

type LogOptions struct {
	Disabled     bool   `json:"disabled,omitempty"`
	Level        string `json:"level,omitempty"`
	Output       string `json:"output,omitempty"`
	Timestamp    bool   `json:"timestamp,omitempty"`
	DisableColor bool   `json:"-"`
}
