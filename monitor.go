package twitterapi

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// MonitorService wraps /oapi/x_user_stream/* — the user-monitoring stream
// (separate from filter rules). Requires an active monitoring subscription.
type MonitorService struct{ c *Client }

type MonitoredListResponse struct {
	Data []MonitoredUser `json:"data"`
	APIStatus
}

// AddUser starts monitoring a screen name. The leading "@" is stripped.
func (s *MonitorService) AddUser(ctx context.Context, xUserName string) (SimpleStatusResponse, error) {
	xUserName = strings.TrimPrefix(strings.TrimSpace(xUserName), "@")
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/oapi/x_user_stream/add_user_to_monitor_tweet", map[string]any{"x_user_name": xUserName}, &r)
	return r, err
}

// QueryType filters MonitoredList results.
type MonitorQueryType int

const (
	MonitorAll     MonitorQueryType = 0
	MonitorTweets  MonitorQueryType = 1 // server default
	MonitorProfile MonitorQueryType = 2
)

// List returns the current monitored users. queryType may be 0/1/2 (see
// MonitorQueryType constants); pass -1 to omit.
func (s *MonitorService) List(ctx context.Context, queryType MonitorQueryType) ([]MonitoredUser, error) {
	q := url.Values{}
	if queryType >= 0 {
		q.Set("query_type", strconv.Itoa(int(queryType)))
	}
	var r MonitoredListResponse
	if err := s.c.getJSON(ctx, "/oapi/x_user_stream/get_user_to_monitor_tweet", q, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

// RemoveUser stops monitoring by id_for_user (NOT the X user_id — you get
// id_for_user from List).
func (s *MonitorService) RemoveUser(ctx context.Context, idForUser string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/oapi/x_user_stream/remove_user_to_monitor_tweet", map[string]any{"id_for_user": idForUser}, &r)
	return r, err
}

// MonitorAccount returns the current monitor account configuration.
func (s *MonitorService) MonitorAccount(ctx context.Context, queryType MonitorQueryType) ([]byte, error) {
	q := url.Values{}
	if queryType >= 0 {
		q.Set("query_type", strconv.Itoa(int(queryType)))
	}
	return s.c.rawGet(ctx, "/oapi/x_user_stream/get_user_monitor_account", q)
}
