package constant

import "codedln/util/types"

const MaxPayloadSize int64 = 10_485_760 // 10MB
const UserCollection = "users"

const GoogleSignIn types.OAuthSignIn = "google"
const GitHubSignIn types.OAuthSignIn = "github"

const (
	PayloadKey  types.ContextKey = iota // iota increments for each constant, ensuring uniqueness
	AuthUserKey types.ContextKey = iota
)

const AccessTokenTTL = 24 //hours
const JwtCookieName = "_access_token"

const AliasLength = 8
const AliasRetry = 50

const Newest types.DateSort = -1
const Oldest types.DateSort = 1
