package twitterapi

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
)

// MediaService groups multipart upload and the three PATCH /update_*_v2
// endpoints (profile/avatar/banner). Like ActionsService, every call carries
// login_cookies + proxy.
type MediaService struct{ c *Client }

// UploadMediaResponse is the payload of /twitter/upload_media_v2.
type UploadMediaResponse struct {
	MediaID string `json:"media_id"`
	APIStatus
}

// UploadOpts configures Upload.
type UploadOpts struct {
	// IsLongVideo opts into long-form video upload. Required for clips >2:20.
	IsLongVideo *bool

	// Proxy overrides the per-call proxy.
	Proxy string
}

// Upload sends raw bytes to /twitter/upload_media_v2 as multipart/form-data.
// Filename should include the correct extension — the server uses it to infer
// the media category.
func (s *MediaService) Upload(ctx context.Context, filename string, body io.Reader, opts *UploadOpts) (UploadMediaResponse, error) {
	if strings.TrimSpace(filename) == "" {
		return UploadMediaResponse{}, errors.New("twitterapi: filename is required")
	}
	if body == nil {
		return UploadMediaResponse{}, errors.New("twitterapi: body is required")
	}
	cookie, err := s.c.Account.EnsureLogin(ctx)
	if err != nil {
		return UploadMediaResponse{}, err
	}
	proxy := ""
	if opts != nil {
		proxy = opts.Proxy
	}
	proxy = s.c.pickProxy(proxy)
	if proxy == "" {
		return UploadMediaResponse{}, errors.New("twitterapi: Upload requires a proxy")
	}
	var r UploadMediaResponse
	err = s.c.postMultipart(ctx, "POST", "/twitter/upload_media_v2", func(w *multipart.Writer) error {
		fw, err := w.CreateFormFile("file", filename)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, body); err != nil {
			return err
		}
		_ = w.WriteField("login_cookies", cookie)
		_ = w.WriteField("proxy", proxy)
		if opts != nil && opts.IsLongVideo != nil {
			_ = w.WriteField("is_long_video", strconv.FormatBool(*opts.IsLongVideo))
		}
		return nil
	}, &r)
	return r, err
}

// UpdateProfileParams configures UpdateProfile. Set the fields you want to
// change; nil/empty means "leave alone". At least one must be provided.
type UpdateProfileParams struct {
	Name        *string // ≤ 50 chars
	Description *string // ≤ 160 chars
	Location    *string // ≤ 30 chars
	URL         *string
	Proxy       string
}

// UpdateProfile updates display fields on the authed user's profile.
func (s *MediaService) UpdateProfile(ctx context.Context, p UpdateProfileParams) (SimpleStatusResponse, error) {
	if p.Name == nil && p.Description == nil && p.Location == nil && p.URL == nil {
		return SimpleStatusResponse{}, errors.New("twitterapi: UpdateProfile requires at least one field")
	}
	payload := map[string]any{}
	if p.Name != nil {
		payload["name"] = *p.Name
	}
	if p.Description != nil {
		payload["description"] = *p.Description
	}
	if p.Location != nil {
		payload["location"] = *p.Location
	}
	if p.URL != nil {
		payload["url"] = *p.URL
	}
	cookie, err := s.c.Account.EnsureLogin(ctx)
	if err != nil {
		return SimpleStatusResponse{}, err
	}
	proxy := s.c.pickProxy(p.Proxy)
	if proxy == "" {
		return SimpleStatusResponse{}, errors.New("twitterapi: UpdateProfile requires a proxy")
	}
	payload["login_cookies"] = cookie
	payload["proxy"] = proxy

	var r SimpleStatusResponse
	err = s.c.patchJSON(ctx, "/twitter/update_profile_v2", payload, &r)
	if err != nil && errors.Is(err, ErrCookieExpired) {
		s.c.clearToken()
		cookie2, e2 := s.c.Account.EnsureLogin(ctx)
		if e2 != nil {
			return r, err
		}
		payload["login_cookies"] = cookie2
		err = s.c.patchJSON(ctx, "/twitter/update_profile_v2", payload, &r)
	}
	return r, err
}

// UpdateAvatar uploads a new profile picture.
//
//	JPG/PNG, ≤ 700 KB, recommended 400×400.
func (s *MediaService) UpdateAvatar(ctx context.Context, filename string, body io.Reader, proxy string) (SimpleStatusResponse, error) {
	return s.patchImage(ctx, "/twitter/update_avatar_v2", filename, body, proxy)
}

// UpdateBanner uploads a new banner image.
//
//	JPG/PNG, ≤ 2 MB, recommended 1500×500.
func (s *MediaService) UpdateBanner(ctx context.Context, filename string, body io.Reader, proxy string) (SimpleStatusResponse, error) {
	return s.patchImage(ctx, "/twitter/update_banner_v2", filename, body, proxy)
}

func (s *MediaService) patchImage(ctx context.Context, path, filename string, body io.Reader, proxyOverride string) (SimpleStatusResponse, error) {
	if strings.TrimSpace(filename) == "" || body == nil {
		return SimpleStatusResponse{}, errors.New("twitterapi: filename and body are required")
	}
	cookie, err := s.c.Account.EnsureLogin(ctx)
	if err != nil {
		return SimpleStatusResponse{}, err
	}
	proxy := s.c.pickProxy(proxyOverride)
	if proxy == "" {
		return SimpleStatusResponse{}, errors.New("twitterapi: requires a proxy")
	}
	// Body must be readable possibly twice — but multipart-PATCH is rare; we
	// don't retry on cookie expiry to keep the body stream simple. Callers
	// should refresh tokens out-of-band for these endpoints.
	var r SimpleStatusResponse
	err = s.c.postMultipart(ctx, "PATCH", path, func(w *multipart.Writer) error {
		fw, err := w.CreateFormFile("file", filename)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, body); err != nil {
			return err
		}
		_ = w.WriteField("login_cookies", cookie)
		_ = w.WriteField("proxy", proxy)
		return nil
	}, &r)
	return r, err
}
