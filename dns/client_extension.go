// karing
package dns

func (c *Client) Close() {
	c.ClearCache()
	c.cache = nil
	c.transportCache = nil
	c.rdrc = nil
}
