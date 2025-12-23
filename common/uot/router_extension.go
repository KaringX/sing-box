// karing
package uot

import (
	"github.com/sagernet/sing-box/adapter"
)

func (r *Router) GetRouter() adapter.ConnectionRouter {
	return r.router
}
