package server

import (
	"net/http"
	"os"
	"table-tennis/pkg/jwttoken"
	"time"

	"github.com/google/uuid"
)

func NewAccessToken(at_id string, val uint, exp time.Time) (*http.Cookie, error) {
	alg := jwttoken.HmacSha256(os.Getenv(jwtKey))
	claims := jwttoken.NewClaims()
	claims.Set("user_id", val)
	claims.Set("at_id", at_id)
	claims.Set("exp", exp.Unix())
	token, err := alg.Encode(claims)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     "at",
		Value:    token,
		Expires:  exp,
		Path:     "/",
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func NewRefreshToken(at_id string, exp time.Time) (*http.Cookie, error) {
	alg := jwttoken.HmacSha256(os.Getenv(jwtKey))
	claims := jwttoken.NewClaims()
	claims.Set("id", uuid.New().String())
	claims.Set("at_id", at_id)
	claims.Set("exp", exp.Unix())
	token, err := alg.Encode(claims)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     "rt",
		Value:    token,
		Expires:  exp,
		Path:     "/",
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func DeleteAccessToken() *http.Cookie {
	deletedCookie := http.Cookie{
		Name:     "at",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	return &deletedCookie
}

func DeleteRefreshToken() *http.Cookie {
	deletedCookie := http.Cookie{
		Name:     "rt",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	return &deletedCookie
}
