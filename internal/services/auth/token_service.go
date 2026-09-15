package services

import (
	"crypto/rand"
	"encoding/hex"
	models "fairchild_be/internal/models/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims is the JWT payload signed into every access token.
type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenService struct {
	cfg *models.JWTConfig
}

func NewTokenService(cfg *models.JWTConfig) *TokenService {
	return &TokenService{cfg: cfg}
}

func (t *TokenService) GenerateAccessToken(userID int64, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(t.cfg.AccessTTLMinutes) * time.Minute)

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(t.cfg.AccessSecret))

	return signed, expiresAt, err
}

func (t *TokenService) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
		return []byte(t.cfg.AccessSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired access token")
	}

	return claims, nil
}

func (t *TokenService) RefreshTokenTTL() time.Duration {
	return time.Duration(t.cfg.RefreshTTLDays) * 24 * time.Hour
}

// GenerateOpaqueToken produces the raw refresh token handed to the client.
// Only its SHA-256 hash (see utils.MustStringSHA256) is ever persisted.
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
