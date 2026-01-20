// karing
package rule

func (s *RemoteRuleSet) Url() string {
	return s.options.RemoteOptions.URL
}

func (s *RemoteRuleSet) RulesCount() int {
	return len(s.rules)
}
