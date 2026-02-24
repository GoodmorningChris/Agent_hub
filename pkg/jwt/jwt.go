package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷，包含 user_id 和 agent_id
type Claims struct {
	UserID  int64 `json:"user_id"`
	AgentID int64 `json:"agent_id"`
	jwt.RegisteredClaims
}

// Generate 生成 JWT token
func Generate(secret []byte, userID, agentID int64, expireHours int) (string, error) {
	exp := time.Now().Add(time.Duration(expireHours) * time.Hour)
	claims := &Claims{
		UserID:  userID,
		AgentID: agentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// Parse 解析并校验 JWT，返回 Claims
func Parse(secret []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
