package twitterapi

import (
	"context"
	"net/url"
)

// SpacesService groups /twitter/spaces/* endpoints.
type SpacesService struct{ c *Client }

type SpaceDetailResponse struct {
	Data Space `json:"data"`
	APIStatus
}

// Detail fetches /twitter/spaces/detail.
func (s *SpacesService) Detail(ctx context.Context, spaceID string) (Space, error) {
	q := url.Values{}
	q.Set("space_id", spaceID)
	var r SpaceDetailResponse
	if err := s.c.getJSON(ctx, "/twitter/spaces/detail", q, &r); err != nil {
		return Space{}, err
	}
	return r.Data, nil
}
