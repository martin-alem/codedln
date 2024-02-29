package constant

import "codedln/util/types"

const MaxPayloadSize int64 = 10_485_760 // 10MB
const UserCollection = "users"
const UrlCollection = "urls"

const GoogleSignIn types.OAuthSignIn = "google"
const GitHubSignIn types.OAuthSignIn = "github"

const (
	PayloadKey  types.ContextKey = iota // iota increments for each constant, ensuring uniqueness
	AuthUserKey types.ContextKey = iota
)

const AccessTokenTTL = 24 //hours
const JwtCookieName = "_codedln_access_token"

const AliasMaxLength = 8
const AliasMinLength = 3
const AliasRetry = 50

const NewestDate types.DateSort = -1
const OldestDate types.DateSort = 1
const MaxLimit int64 = 25
