package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Sub         string `json:"sub"`
	WorkspaceID string `json:"workspace_id"`
	TenantID    string `json:"tenant_id"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

func IssueJWT(secret string, sub, tenantID, workspaceID, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Sub:         sub,
		WorkspaceID: workspaceID,
		TenantID:    tenantID,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}


