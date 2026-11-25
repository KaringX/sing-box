//karing

package libbox

func (o *tunOptions) GetAllowBypass() bool {
	return o.TunPlatformOptions.AllowBypass
}

func (o *tunOptions) GetExcludeApns() bool {
	return o.TunPlatformOptions.ExcludeApns
}

func (o *tunOptions) GetAutoRouteUseSubRangesByDefault() bool {
	return o.TunPlatformOptions.AutoRouteUseSubRangesByDefault
}
