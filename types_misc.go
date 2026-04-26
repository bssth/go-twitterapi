package twitterapi

// Trend is a single trending topic.
type Trend struct {
	Name       string `json:"name,omitempty"`
	URL        string `json:"url,omitempty"`
	Query      string `json:"query,omitempty"`
	TweetCount int64  `json:"tweet_volume,omitempty"`
}

// Community is the info payload of a Twitter Community.
type Community struct {
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name,omitempty"`
	Description    string          `json:"description,omitempty"`
	Question       string          `json:"question,omitempty"`
	MemberCount    int64           `json:"member_count,omitempty"`
	ModeratorCount int64           `json:"moderator_count,omitempty"`
	JoinPolicy     string          `json:"join_policy,omitempty"`
	InvitesPolicy  string          `json:"invites_policy,omitempty"`
	IsNSFW         bool            `json:"is_nsfw,omitempty"`
	IsPinned       bool            `json:"is_pinned,omitempty"`
	BannerURL      string          `json:"banner_url,omitempty"`
	SearchTags     []string        `json:"search_tags,omitempty"`
	Rules          []CommunityRule `json:"rules,omitempty"`
	Creator        *User           `json:"creator,omitempty"`
	Admin          *User           `json:"admin,omitempty"`
	MembersPreview []User          `json:"members_preview,omitempty"`
	CreatedAt      string          `json:"created_at,omitempty"`
	PrimaryTopic   map[string]any  `json:"primary_topic,omitempty"`
	Role           string          `json:"role,omitempty"`
}

// CommunityRule is a single house rule of a community.
type CommunityRule struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// SpaceParticipants groups admin/speaker/listener users for a Space.
type SpaceParticipants struct {
	Admins    []User `json:"admins,omitempty"`
	Speakers  []User `json:"speakers,omitempty"`
	Listeners []User `json:"listeners,omitempty"`
}

// SpaceStats are aggregate counters for a Space.
type SpaceStats struct {
	ReplayViews       int64 `json:"replay_views,omitempty"`
	LiveListeners     int64 `json:"live_listeners,omitempty"`
	TotalParticipants int64 `json:"total_participants,omitempty"`
}

// Space describes a single X Space.
type Space struct {
	ID             string            `json:"id,omitempty"`
	Title          string            `json:"title,omitempty"`
	State          string            `json:"state,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	ScheduledStart string            `json:"scheduled_start,omitempty"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
	MediaKey       string            `json:"media_key,omitempty"`
	IsSubscribed   bool              `json:"is_subscribed,omitempty"`
	Settings       map[string]any    `json:"settings,omitempty"`
	Stats          SpaceStats        `json:"stats,omitempty"`
	Creator        User              `json:"creator,omitempty"`
	Participants   SpaceParticipants `json:"participants,omitempty"`
}

// AccountInfo is the response of /oapi/my/info.
type AccountInfo struct {
	RechargeCredits   float64 `json:"recharge_credits,omitempty"`
	TotalBonusCredits float64 `json:"total_bonus_credits,omitempty"`
}

// MonitoredUser is one entry in the user-stream monitor list.
type MonitoredUser struct {
	IDForUser                  string `json:"id_for_user,omitempty"`
	XUserID                    string `json:"x_user_id,omitempty"`
	XUserName                  string `json:"x_user_name,omitempty"`
	XUserScreenName            string `json:"x_user_screen_name,omitempty"`
	IsMonitorTweet             bool   `json:"is_monitor_tweet,omitempty"`
	IsMonitorProfile           bool   `json:"is_monitor_profile,omitempty"`
	MonitorTweetConfigStatus   string `json:"monitor_tweet_config_status,omitempty"`
	MonitorProfileConfigStatus string `json:"monitor_profile_config_status,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
}

// FilterRule is a webhook filter rule.
type FilterRule struct {
	RuleID          string  `json:"rule_id,omitempty"`
	UserID          string  `json:"user_id,omitempty"`
	Tag             string  `json:"tag,omitempty"`
	Value           string  `json:"value,omitempty"`
	IntervalSeconds float64 `json:"interval_seconds,omitempty"`
	IsEffect        int     `json:"is_effect,omitempty"`
	IsDelete        int     `json:"is_delete,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
}
