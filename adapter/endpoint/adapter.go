package endpoint

import "github.com/sagernet/sing-box/option"

type Adapter struct {
	endpointType string
	endpointTag  string
	network      []string
	dependencies []string
	parseErr     error //karing
}

func NewAdapter(endpointType string, endpointTag string, network []string, dependencies []string) Adapter {
	return Adapter{
		endpointType: endpointType,
		endpointTag:  endpointTag,
		network:      network,
		dependencies: dependencies,
	}
}

func NewAdapterWithDialerOptions(endpointType string, endpointTag string, network []string, dialOptions option.DialerOptions) Adapter {
	var dependencies []string
	if dialOptions.Detour != "" {
		dependencies = []string{dialOptions.Detour}
	}
	return NewAdapter(endpointType, endpointTag, network, dependencies)
}

func (a *Adapter) Type() string {
	return a.endpointType
}

func (a *Adapter) Tag() string {
	return a.endpointTag
}

func (a *Adapter) Network() []string {
	return a.network
}

func (a *Adapter) Dependencies() []string {
	return a.dependencies
}

func (h *Adapter) SetParseErr(err error) { //karing
	h.parseErr = err
}

func (h *Adapter) GetParseErr() error { //karing
	return h.parseErr
}
