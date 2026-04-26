package twitterapi

import "context"

// WebhookService wraps /oapi/tweet_filter/*. Filter rules feed both the
// dashboard-registered HTTP webhook and the experimental WebSocket stream
// (see WSClient).
type WebhookService struct{ c *Client }

// AddRuleResponse is the payload of /oapi/tweet_filter/add_rule.
type AddRuleResponse struct {
	RuleID string `json:"rule_id"`
	APIStatus
}

// GetRulesResponse is the payload of /oapi/tweet_filter/get_rules.
type GetRulesResponse struct {
	Rules []FilterRule `json:"rules"`
	APIStatus
}

// AddRule registers a new filter rule. New rules are inactive by default —
// activate via UpdateRule(... isEffect=1).
//
//	tag:             ≤ 255 chars, your label
//	value:           ≤ 255 chars, X advanced-search query (e.g. "from:elonmusk")
//	intervalSeconds: 0.05 – 86400. Default 60.
func (s *WebhookService) AddRule(ctx context.Context, tag, value string, intervalSeconds float64) (AddRuleResponse, error) {
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}
	payload := map[string]any{
		"tag":              tag,
		"value":            value,
		"interval_seconds": intervalSeconds,
	}
	var r AddRuleResponse
	err := s.c.postJSON(ctx, "/oapi/tweet_filter/add_rule", payload, &r)
	return r, err
}

// UpdateRule mutates an existing rule. Pass active=true to activate (is_effect=1).
func (s *WebhookService) UpdateRule(ctx context.Context, ruleID, tag, value string, intervalSeconds float64, active bool) (SimpleStatusResponse, error) {
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}
	isEffect := 0
	if active {
		isEffect = 1
	}
	payload := map[string]any{
		"rule_id":          ruleID,
		"tag":              tag,
		"value":            value,
		"interval_seconds": intervalSeconds,
		"is_effect":        isEffect,
	}
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/oapi/tweet_filter/update_rule", payload, &r)
	return r, err
}

// DeleteRule removes a rule. The API expects the rule_id in a JSON body even
// though the verb is DELETE.
func (s *WebhookService) DeleteRule(ctx context.Context, ruleID string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.c.deleteJSON(ctx, "/oapi/tweet_filter/delete_rule", map[string]any{"rule_id": ruleID}, &r)
	return r, err
}

// ListRules returns every rule registered for the account.
func (s *WebhookService) ListRules(ctx context.Context) ([]FilterRule, error) {
	var r GetRulesResponse
	if err := s.c.getJSON(ctx, "/oapi/tweet_filter/get_rules", nil, &r); err != nil {
		return nil, err
	}
	return r.Rules, nil
}
