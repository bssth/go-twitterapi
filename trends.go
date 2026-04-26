package twitterapi

import (
	"context"
	"net/url"
	"strconv"
)

// TrendsService wraps /twitter/trends.
type TrendsService struct{ c *Client }

type TrendsResponse struct {
	Trends []Trend `json:"trends"`
	APIStatus
}

// Get returns trending topics for a WOEID. Common values: 1 = worldwide,
// 23424977 = US. Server enforces a minimum of 30 results.
func (s *TrendsService) Get(ctx context.Context, woeid int64, count int) ([]Trend, error) {
	q := url.Values{}
	q.Set("woeid", strconv.FormatInt(woeid, 10))
	if count > 0 {
		q.Set("count", strconv.Itoa(count))
	}
	var r TrendsResponse
	if err := s.c.getJSON(ctx, "/twitter/trends", q, &r); err != nil {
		return nil, err
	}
	return r.Trends, nil
}
