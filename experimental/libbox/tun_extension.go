//go:build with_karing

package libbox

func (o *tunOptions) GetAllowBypass() bool {
	return o.TunPlatformOptions.AllowBypass
}
