# go-twitterapi

[![CI](https://github.com/bssth/go-twitterapi/actions/workflows/ci.yml/badge.svg)](https://github.com/bssth/go-twitterapi/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/bssth/go-twitterapi.svg)](https://pkg.go.dev/github.com/bssth/go-twitterapi)
[![Go Report Card](https://goreportcard.com/badge/github.com/bssth/go-twitterapi)](https://goreportcard.com/report/github.com/bssth/go-twitterapi)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Idiomatic Go SDK for [twitterapi.io](https://docs.twitterapi.io/introduction) — a third-party gateway to the Twitter / X API. Read profiles, walk timelines, post tweets, send DMs, manage Communities and List members, register webhook filter rules, and tap the experimental WebSocket stream — all from one typed client.

```go
c, _ := twitterapi.New(twitterapi.Options{APIKey: os.Getenv("TWITTERAPIIO_API_KEY")})

user, _   := c.Users.GetByUsername(ctx, "elonmusk")
tweets, _ := c.Tweets.ByIDs(ctx, []string{"1234567890"})
trends, _ := c.Trends.Get(ctx, 1, 30) // worldwide

it := c.Users.Followers(ctx, "openai", &twitterapi.FollowersOpts{PageSize: 200})
for it.Next() { fmt.Println(it.Item().UserName) }
```

## Features

- **Full API coverage** — every documented twitterapi.io endpoint, grouped by resource.
- **Typed responses** — `User`, `Tweet`, `Community`, `Space`, `Trend`, `Article`, ...
- **Cursor pagination as iterators** — `Next() / Item() / Err()` plus raw `*Page` methods.
- **Two auth modes done right** — read endpoints take just an API key; v2 writes auto-attach `login_cookies` + `proxy` and refresh on expiry.
- **Persistent login** — pluggable `TokenStore` (file or in-memory) caches the cookie between runs.
- **Robust HTTP** — exponential backoff with jitter, `Retry-After` honored, automatic retries on 429 / 5xx / transient network errors.
- **Multipart uploads** — media, avatar, and banner with the right `PATCH` verbs.
- **Webhook filter rules** — full CRUD on `/oapi/tweet_filter/*`.
- **Experimental WebSocket stream** — auto-reconnecting client for low-latency tweet delivery.
- **Legacy v1 endpoints kept** — `auth_session`-based methods preserved as `c.Legacy.*`.
- **Zero magic** — no global state, no hidden goroutines, `*Client` is concurrency-safe.

## Install

```bash
go get github.com/bssth/go-twitterapi
```

Requires Go 1.22+ (uses generics for the iterator).

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/bssth/go-twitterapi"
)

func main() {
    c, err := twitterapi.New(twitterapi.Options{
        // APIKey is required. Falls back to env TWITTERAPIIO_API_KEY.
    })
    if err != nil { log.Fatal(err) }

    ctx := context.Background()

    user, err := c.Users.GetByUsername(ctx, "elonmusk")
    if err != nil { log.Fatal(err) }
    fmt.Printf("%s — %d followers\n", user.Name, user.Followers)
}
```

## Authentication

The SDK is built around two scopes:

| Scope     | What it covers                                 | What you need                        |
| --------- | ---------------------------------------------- | ------------------------------------ |
| Read      | Users, tweets, communities, spaces, search ... | `x-api-key` only                     |
| v2 write  | Tweet, like, follow, DM, upload media, ...     | `x-api-key` + `login_cookies` + proxy |

### Read-only

```go
c, _ := twitterapi.New(twitterapi.Options{APIKey: "...your key..."})
```

### v2 writes

Three setups, pick the one that fits:

**1. You already have a `login_cookies` string from another process:**

```go
c, _ := twitterapi.New(twitterapi.Options{
    APIKey:       os.Getenv("TWITTERAPIIO_API_KEY"),
    LoginCookie:  os.Getenv("LOGIN_COOKIE"),
    DefaultProxy: os.Getenv("MY_PROXY"),
})
```

**2. Persist the cookie to disk and let the SDK log in lazily:**

```go
c, _ := twitterapi.New(twitterapi.Options{
    APIKey:    os.Getenv("TWITTERAPIIO_API_KEY"),
    TokenFile: "twitterapiio.token.json",
})
// Set TWITTERAPIIO_USER_NAME, _EMAIL, _PASSWORD, _PROXY (and optionally
// _TOTP_SECRET) and the first write triggers user_login_v2 automatically.
```

**3. Drive login yourself:**

```go
cookie, err := c.Account.LoginV2(ctx, twitterapi.LoginV2Params{
    UserName: "...", Email: "...", Password: "...",
    Proxy: "http://user:pass@host:port",
    TOTPSecret: "...", // strongly recommended
})
```

When the API returns "login_cookies expired/invalid", the SDK clears the cached cookie, re-runs `EnsureLogin`, and retries the original write **once**.

## Resource map

```text
c.Users        /twitter/user/info, _about, batch_info_by_ids, search,
               followers, followings, verifiedFollowers, mentions,
               last_tweets, tweet_timeline, check_follow_relationship
c.Tweets       /twitter/tweets, /tweet/replies, /replies/v2, /quotes,
               /retweeters, /thread_context, /article,
               /tweet/advanced_search, /tweet/bulk_advanced_search
c.Communities  /twitter/community/info, members, moderators, tweets,
               get_tweets_from_all_community
c.Spaces       /twitter/spaces/detail
c.Trends       /twitter/trends
c.Lists        /twitter/list/tweets, tweets_timeline, members, followers
c.Account      /oapi/my/info, user_login_v2, user_login_v3,
               get_my_x_account_detail_v3, delete_my_x_account_v3
c.Actions      /twitter/{create,delete,like,unlike,retweet,bookmark,
               unbookmark}_tweet_v2, /bookmarks_v2,
               /follow_user_v2, /unfollow_user_v2, /send_dm_to_user,
               /report_v2, /create_community_v2, /join_community_v2,
               /leave_community_v2, /delete_community_v2,
               /list/add_member_v2
c.Media        /twitter/upload_media_v2,
               PATCH update_profile_v2, update_avatar_v2, update_banner_v2
c.Monitor      /oapi/x_user_stream/* — real-time user monitoring
c.Webhook      /oapi/tweet_filter/{add_rule,update_rule,delete_rule,get_rules}
c.Legacy       deprecated v1 endpoints (auth_session)
```

For the WebSocket stream, see `WSClient` (constructed via `twitterapi.NewWSClient(c)`).

## Pagination

Every cursor-paginated endpoint exposes two flavors:

```go
// Iterator — walks every page until exhaustion or error.
it := c.Tweets.AdvancedSearch(ctx, "from:openai", &twitterapi.AdvancedSearchOpts{QueryType: "Latest"})
it.MaxItems = 500       // optional safety cap
for it.Next() {
    t := it.Item()
    fmt.Println(t.ID, t.Text)
}
if err := it.Err(); err != nil { log.Fatal(err) }

// Page — drive the cursor yourself, e.g. to checkpoint progress.
page, err := c.Tweets.AdvancedSearchPage(ctx, "from:openai", &twitterapi.AdvancedSearchOpts{
    QueryType: "Latest",
    Cursor:    savedCursor,
})
```

Iterators stop the moment `has_next_page` flips to `false` or `next_cursor` empties — even if the server keeps returning data, which prevents accidental infinite loops on misbehaving endpoints.

## Posting tweets

```go
resp, err := c.Actions.CreateTweet(ctx, twitterapi.CreateTweetParams{
    TweetText:      "hello from go-twitterapi",
    ReplyToTweetID: "...",                       // optional
    QuoteTweetID:   "...",                       // optional
    MediaIDs:       []string{mediaID},           // optional
    ScheduleFor:    "2026-04-01T12:00:00.000Z",  // optional
})
```

Long-form posts:

```go
yes := true
resp, _ := c.Actions.CreateTweet(ctx, twitterapi.CreateTweetParams{
    TweetText:   longText,
    IsNoteTweet: &yes,
})
```

`SanitizeForTwitter` runs by default — strips zero-width characters and normalizes smart punctuation. Disable with `SkipSanitize: true`.

## Uploading media

```go
f, _ := os.Open("photo.jpg")
defer f.Close()

up, err := c.Media.Upload(ctx, "photo.jpg", f, nil)
if err != nil { log.Fatal(err) }

c.Actions.CreateTweet(ctx, twitterapi.CreateTweetParams{
    TweetText: "with a picture",
    MediaIDs:  []string{up.MediaID},
})
```

Avatar / banner use `PATCH` and dedicated helpers:

```go
c.Media.UpdateAvatar(ctx, "me.png", avatarFile, "")
c.Media.UpdateBanner(ctx, "banner.png", bannerFile, "")

newName := "Mr. Burns"
c.Media.UpdateProfile(ctx, twitterapi.UpdateProfileParams{Name: &newName})
```

## Webhook filter rules

Rules feed both your dashboard-registered HTTPS webhook and the WebSocket stream. New rules are inactive by default — call `UpdateRule(... active=true)` to start delivery.

```go
rule, _ := c.Webhook.AddRule(ctx, "openai-mentions", "from:openai OR @openai", 60)
c.Webhook.UpdateRule(ctx, rule.RuleID, "openai-mentions", "from:openai OR @openai", 60, true)

rules, _ := c.Webhook.ListRules(ctx)
c.Webhook.DeleteRule(ctx, rule.RuleID)
```

## WebSocket stream (experimental)

The WS endpoint is not officially documented — treat it as best-effort. Auto-reconnects after `ReconnectDelay` (default 90s) on read errors.

```go
ws := twitterapi.NewWSClient(c)
ws.Logger = log.Printf

err := ws.ConnectAndRead(ctx, func(ev twitterapi.WSEvent) {
    tweets, _ := ev.DecodeTweets()
    for _, t := range tweets {
        fmt.Printf("[%s] @%s: %s\n", ev.RuleTag, t.Author.UserName, t.Text)
    }
})
```

## Errors

Every non-2xx response (and every 200 response with `"status":"error"`) becomes an `*APIError`:

```go
_, err := c.Users.GetByUsername(ctx, "doesnotexist_____")

var ae *twitterapi.APIError
if errors.As(err, &ae) {
    fmt.Println(ae.StatusCode, ae.Message, string(ae.Body))
}
```

Sentinel errors:

```go
errors.Is(err, twitterapi.ErrInsufficientCredits) // HTTP 402 — top up balance
errors.Is(err, twitterapi.ErrCookieExpired)       // login_cookies invalid (auto-retried once)
errors.Is(err, twitterapi.ErrNeedLogin)           // no cookie + no env to log in
errors.Is(err, twitterapi.ErrMissingAPIKey)       // New() without a key
```

## Retries and rate limits

Defaults you can override on `Options`:

| Field        | Default            | Notes                                     |
| ------------ | ------------------ | ----------------------------------------- |
| `MaxRetries` | 5                  | Applied on 429 / 5xx / transient net errors |
| `MinBackoff` | 300ms              | Doubled each attempt (with jitter)        |
| `MaxBackoff` | 8s                 | Cap                                       |
| `FreePlan`   | false              | When true, throttles to 1 request / 5s    |

The client honors `Retry-After` on 429 / 503 responses.

Cost-sensitive endpoints (paginated followers, deep searches) can run up real money — bound iterators with `MaxItems` or `MaxPages`, and check `c.Account.Info(ctx)` for your remaining credits.

## Configuration reference

```go
type Options struct {
    APIKey       string         // or env TWITTERAPIIO_API_KEY
    BaseURL      string         // default: https://api.twitterapi.io
    WSURL        string         // default: wss://ws.twitterapi.io/twitter/tweet/websocket
    HTTPClient   *http.Client   // override transport / timeouts
    UserAgent    string

    FreePlan     bool
    MaxRetries   int
    MinBackoff   time.Duration
    MaxBackoff   time.Duration

    TokenStore   TokenStore     // implement to persist login_cookies
    TokenFile    string         // shorthand: creates a FileTokenStore
    LoginCookie  string         // pre-populate the cookie

    DefaultProxy string         // or env TWITTERAPIIO_PROXY
}
```

Custom `TokenStore` implementations need only `Load() (*LoginState, error)` and `Save(*LoginState) error`.

## Examples

Runnable programs in [`examples/`](examples):

| Example                                                  | What it does                                                                                |
| -------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| [`basic`](examples/basic)                                | Read a profile and last 5 tweets.                                                           |
| [`pagination`](examples/pagination)                      | Iterate followers and search results, with bounded walks.                                   |
| [`search`](examples/search)                              | Run an advanced search with a result cap.                                                   |
| [`thread_chain`](examples/thread_chain)                  | Reconstruct a tweet's reply chain back to root.                                             |
| [`communities`](examples/communities)                    | Print a community's metadata + first 50 members.                                            |
| [`trends_dashboard`](examples/trends_dashboard)          | Side-by-side trends for multiple WOEIDs.                                                    |
| [`post_tweet`](examples/post_tweet)                      | Env-driven login + create a tweet.                                                          |
| [`upload_media`](examples/upload_media)                  | Upload a local file and tweet with it attached.                                             |
| [`dm`](examples/dm)                                      | Send a DM and read the message history.                                                     |
| [`monitor`](examples/monitor)                            | List / add / remove users in the monitor stream.                                            |
| [`webhook_stream`](examples/webhook_stream)              | Register a filter rule and consume the WebSocket.                                           |
| [`account_balance`](examples/account_balance)            | Print remaining credits — handy as a CI guard.                                              |

Run any example:

```bash
export TWITTERAPIIO_API_KEY=...
go run ./examples/basic
```

## Versioning

Pre-1.0. The package surface may shift between minor versions — pin a tag in your `go.mod` until 1.0.

## Contributing

PRs welcome. Please:

- run `go vet ./... && go build ./...` before pushing,
- keep new endpoints inside their resource service file,
- add an example to `examples/` for non-trivial features.

## License

MIT — see [LICENSE](LICENSE).

This SDK is **not** affiliated with twitterapi.io or X Corp.
