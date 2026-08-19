// Package claims propagates the authenticated caller's identity across
// every transport this service speaks: an HTTP header on the way in, a
// context.Context value once inside the process, and a message-broker
// header when an event is published — so a consumer processing that event
// later can still answer "who triggered this".
package claims

import "context"

// Claims is the identity extracted from a validated JWT.
type Claims struct {
	UserID   string
	Username string
	Email    string
	Role     string
}

const (
	headerUserID   = "x-user-id"
	headerUsername = "x-username"
	headerEmail    = "x-email"
	headerRole     = "x-role"
)

type contextKey struct{}

var claimsKey = contextKey{}

// SetClaims returns a new context carrying claims.
func SetClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// GetClaims returns the claims stored in ctx, or a zero Claims if none were
// set (e.g. an unauthenticated request to a public endpoint).
func GetClaims(ctx context.Context) Claims {
	c, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return Claims{}
	}
	return c
}

// ToMetadata flattens Claims into a string map suitable for gRPC metadata or
// message-broker headers.
func (c Claims) ToMetadata() map[string]string {
	md := map[string]string{}
	if c.UserID != "" {
		md[headerUserID] = c.UserID
	}
	if c.Username != "" {
		md[headerUsername] = c.Username
	}
	if c.Email != "" {
		md[headerEmail] = c.Email
	}
	if c.Role != "" {
		md[headerRole] = c.Role
	}
	return md
}

// FromMetadata reconstructs Claims from a string map previously produced by
// ToMetadata (gRPC metadata, HTTP headers, or message-broker headers).
func FromMetadata(md map[string]string) Claims {
	return Claims{
		UserID:   md[headerUserID],
		Username: md[headerUsername],
		Email:    md[headerEmail],
		Role:     md[headerRole],
	}
}

// IsAuthenticated reports whether c was populated from a validated token.
func (c Claims) IsAuthenticated() bool {
	return c.UserID != ""
}
