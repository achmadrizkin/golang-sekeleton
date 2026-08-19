package claims

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig configures HMAC JWT validation.
type JWTConfig struct {
	Secret string
	Issuer string
}

// customClaims is the JWT payload shape this service expects. Adjust to
// match whatever your identity provider issues.
type customClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ParseAndValidate verifies tokenString's signature (HS256) and expiry, then
// returns the Claims carried inside it.
func ParseAndValidate(tokenString string, cfg JWTConfig) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("claims: parse token: %w", err)
	}

	cc, ok := parsed.Claims.(*customClaims)
	if !ok || !parsed.Valid {
		return Claims{}, fmt.Errorf("claims: invalid token")
	}

	return Claims{
		UserID:   cc.UserID,
		Username: cc.Username,
		Email:    cc.Email,
		Role:     cc.Role,
	}, nil
}
