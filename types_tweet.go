package twitterapi

import "encoding/json"

// HashtagEntity is a single hashtag in a tweet.
type HashtagEntity struct {
	Text    string `json:"text"`
	Indices []int  `json:"indices,omitempty"`
}

// UserMentionEntity is a single @-mention.
type UserMentionEntity struct {
	IDStr      string `json:"id_str"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

// TweetEntities is the parsed entities block of a tweet.
type TweetEntities struct {
	Hashtags     []HashtagEntity     `json:"hashtags,omitempty"`
	URLs         []URLEntity         `json:"urls,omitempty"`
	UserMentions []UserMentionEntity `json:"user_mentions,omitempty"`
}

// Tweet is the canonical tweet payload.
//
// QuotedTweet / RetweetedTweet are populated by UnmarshalJSON when present.
// The raw bytes are kept in QuotedRaw / RetweetedRaw — useful when the API
// returns the sentinel "<unknown>" instead of an object.
type Tweet struct {
	Type   string `json:"type,omitempty"`
	ID     string `json:"id,omitempty"`
	URL    string `json:"url,omitempty"`
	Text   string `json:"text,omitempty"`
	Source string `json:"source,omitempty"`

	RetweetCount int64 `json:"retweetCount,omitempty"`
	ReplyCount   int64 `json:"replyCount,omitempty"`
	LikeCount    int64 `json:"likeCount,omitempty"`
	QuoteCount   int64 `json:"quoteCount,omitempty"`
	ViewCount    int64 `json:"viewCount,omitempty"`

	CreatedAt string `json:"createdAt,omitempty"`
	Lang      string `json:"lang,omitempty"`

	BookmarkCount int64 `json:"bookmarkCount,omitempty"`

	IsReply        bool   `json:"isReply,omitempty"`
	InReplyToId    string `json:"inReplyToId,omitempty"`
	ConversationId string `json:"conversationId,omitempty"`

	DisplayTextRange []int `json:"displayTextRange,omitempty"`

	InReplyToUserId   string `json:"inReplyToUserId,omitempty"`
	InReplyToUsername string `json:"inReplyToUsername,omitempty"`

	Author   User          `json:"author"`
	Entities TweetEntities `json:"entities,omitempty"`

	QuotedTweet    *Tweet          `json:"-"`
	RetweetedTweet *Tweet          `json:"-"`
	QuotedRaw      json.RawMessage `json:"-"`
	RetweetedRaw   json.RawMessage `json:"-"`

	IsLimitedReply bool `json:"isLimitedReply,omitempty"`
}

// UnmarshalJSON deals with the API's habit of returning either an object,
// null, or the literal string "<unknown>" for quoted_tweet / retweeted_tweet.
func (t *Tweet) UnmarshalJSON(b []byte) error {
	type alias Tweet
	var aux struct {
		alias
		Quoted    json.RawMessage `json:"quoted_tweet"`
		Retweeted json.RawMessage `json:"retweeted_tweet"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	*t = Tweet(aux.alias)
	t.QuotedRaw = aux.Quoted
	t.RetweetedRaw = aux.Retweeted

	if isObjectJSON(aux.Quoted) {
		var qt Tweet
		if err := json.Unmarshal(aux.Quoted, &qt); err == nil {
			t.QuotedTweet = &qt
		}
	}
	if isObjectJSON(aux.Retweeted) {
		var rt Tweet
		if err := json.Unmarshal(aux.Retweeted, &rt); err == nil {
			t.RetweetedTweet = &rt
		}
	}
	return nil
}

func isObjectJSON(b json.RawMessage) bool {
	if len(b) == 0 {
		return false
	}
	s := string(b)
	if s == "null" || s == `"<unknown>"` {
		return false
	}
	return s[0] == '{'
}

// ArticleContent is one segment of a long-form X article.
type ArticleContent struct {
	Type       string `json:"type,omitempty"`
	Text       string `json:"text,omitempty"`
	URL        string `json:"url,omitempty"`
	PreviewURL string `json:"previewUrl,omitempty"`
}

// Article is the payload of /twitter/article.
type Article struct {
	Author          User             `json:"author"`
	ReplyCount      int64            `json:"replyCount,omitempty"`
	LikeCount       int64            `json:"likeCount,omitempty"`
	QuoteCount      int64            `json:"quoteCount,omitempty"`
	ViewCount       int64            `json:"viewCount,omitempty"`
	CreatedAt       string           `json:"createdAt,omitempty"`
	Title           string           `json:"title,omitempty"`
	PreviewText     string           `json:"preview_text,omitempty"`
	CoverMediaImage string           `json:"cover_media_img_url,omitempty"`
	Contents        []ArticleContent `json:"contents,omitempty"`
}
