package twitterapi

// URLEntity is a single URL annotation inside a tweet or profile description.
type URLEntity struct {
	DisplayURL  string `json:"display_url,omitempty"`
	ExpandedURL string `json:"expanded_url,omitempty"`
	URL         string `json:"url,omitempty"`
	Indices     []int  `json:"indices,omitempty"`
}

// URLEntityBlock wraps a list of URL entities in profile bios.
type URLEntityBlock struct {
	URLs []URLEntity `json:"urls,omitempty"`
}

// ProfileBioEntities holds the parsed entities of a user's bio.
type ProfileBioEntities struct {
	Description *URLEntityBlock `json:"description,omitempty"`
	URL         *URLEntityBlock `json:"url,omitempty"`
}

// ProfileBio is the description block returned alongside a User.
type ProfileBio struct {
	Description string             `json:"description,omitempty"`
	Entities    ProfileBioEntities `json:"entities,omitempty"`
}

// User describes a Twitter/X account.
type User struct {
	Type     string `json:"type,omitempty"`
	ID       string `json:"id,omitempty"`
	UserName string `json:"userName,omitempty"`
	Name     string `json:"name,omitempty"`
	URL      string `json:"url,omitempty"`

	IsBlueVerified bool   `json:"isBlueVerified,omitempty"`
	VerifiedType   string `json:"verifiedType,omitempty"`

	ProfilePicture string `json:"profilePicture,omitempty"`
	CoverPicture   string `json:"coverPicture,omitempty"`

	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`

	Followers int64 `json:"followers,omitempty"`
	Following int64 `json:"following,omitempty"`

	CanDm     bool   `json:"canDm,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`

	FavouritesCount    int64 `json:"favouritesCount,omitempty"`
	HasCustomTimelines bool  `json:"hasCustomTimelines,omitempty"`
	IsTranslator       bool  `json:"isTranslator,omitempty"`
	MediaCount         int64 `json:"mediaCount,omitempty"`
	StatusesCount      int64 `json:"statusesCount,omitempty"`

	WithheldInCountries []string `json:"withheldInCountries,omitempty"`

	AffiliatesHighlightedLabel map[string]any `json:"affiliatesHighlightedLabel,omitempty"`

	PossiblySensitive bool     `json:"possiblySensitive,omitempty"`
	PinnedTweetIds    []string `json:"pinnedTweetIds,omitempty"`

	IsAutomated bool   `json:"isAutomated,omitempty"`
	AutomatedBy string `json:"automatedBy,omitempty"`

	Unavailable       bool   `json:"unavailable,omitempty"`
	Message           string `json:"message,omitempty"`
	UnavailableReason string `json:"unavailableReason,omitempty"`

	Protected bool `json:"protected,omitempty"`

	ProfileBio ProfileBio `json:"profile_bio,omitempty"`
}

// FollowRelationship is the response payload for check_follow_relationship.
type FollowRelationship struct {
	Following  bool `json:"following"`
	FollowedBy bool `json:"followed_by"`
}
