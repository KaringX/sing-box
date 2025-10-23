// karing
package rule

func (r *RuleActionRoute) Target() string {
	return r.Outbound
}

func (r *RuleActionRouteOptions) Target() string {
	return ""
}

func (r *RuleActionDNSRoute) Target() string {
	return r.Server
}

func (r *RuleActionDNSRouteOptions) Target() string {
	return ""
}

func (r *RuleActionDirect) Target() string {
	return ""
}

func (r *RuleActionReject) Target() string {
	return ""
}

func (r *RuleActionHijackDNS) Target() string {
	return ""
}

func (r *RuleActionSniff) Target() string {
	return ""
}

func (r *RuleActionResolve) Target() string {
	return ""
}

func (r *RuleActionPredefined) Target() string {
	return ""
}
