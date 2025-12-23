// karing
package endpoint

func (h *Adapter) SetParseErr(err error) {
	h.parseErr = err
}

func (h *Adapter) GetParseErr() error {
	return h.parseErr
}
