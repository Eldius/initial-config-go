package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotAuthorized = errors.New("unauthorized")
)

// User represents a user.
type User interface {
	UserID() string
	UserData() map[string]any
}

type ctxUserKey string

const (
	userKey                  ctxUserKey = "user"
	DefaultXApiKeyHeaderName string     = "X-Api-Key"
)

// UserAuthenticationFunc is a function that authenticates a user based on the provided request.
type UserAuthenticationFunc func(r *http.Request) (User, error)

const defaultLookupTokenSize = 12

type apiKeyEntry struct {
	hash string
	user User
}

// ApiKeyMap stores hashed API keys grouped by a deterministic lookup token.
//
// It uses a bucket index to reduce the amount of bcrypt comparisons performed on
// each request while still using bcrypt for the final key verification.
type ApiKeyMap struct {
	buckets      map[string][]apiKeyEntry
	fallback     []apiKeyEntry
	lookupSecret []byte
}

func lookupToken(secret []byte, key string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(key))
	token := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(token[:defaultLookupTokenSize])
}

func (m ApiKeyMap) get(key string) (User, bool) {
	if m.lookupSecret == nil {
		return nil, false
	}

	token := lookupToken(m.lookupSecret, key)

	for _, entry := range m.buckets[token] {
		if bcrypt.CompareHashAndPassword([]byte(entry.hash), []byte(key)) == nil {
			return entry.user, true
		}
	}

	for _, entry := range m.fallback {
		if bcrypt.CompareHashAndPassword([]byte(entry.hash), []byte(key)) == nil {
			return entry.user, true
		}
	}

	return nil, false
}

// NewApiKeyMapFromPlainMap creates a new ApiKeyMap from a plain map of API keys to users.
func NewApiKeyMapFromPlainMap(m map[string]User) (ApiKeyMap, error) {
	lookupSecret := []byte(DefaultXApiKeyHeaderName)
	apiKeyMap := ApiKeyMap{
		buckets:      make(map[string][]apiKeyEntry),
		lookupSecret: lookupSecret,
	}

	for k, v := range m {
		key, err := bcrypt.GenerateFromPassword([]byte(k), bcrypt.DefaultCost)
		if err != nil {
			return ApiKeyMap{}, fmt.Errorf("%w:failed to generate bcrypt key: %w", ErrNotAuthorized, err)
		}

		token := lookupToken(lookupSecret, k)
		apiKeyMap.buckets[token] = append(apiKeyMap.buckets[token], apiKeyEntry{
			hash: string(key),
			user: v,
		})
	}

	return apiKeyMap, nil
}

// NewApiKeyMapFromHashedMap creates a new ApiKeyMap from a map of hashed API keys to users.
func NewApiKeyMapFromHashedMap(m map[string]User) (ApiKeyMap, error) {
	apiKeyMap := ApiKeyMap{
		buckets:      make(map[string][]apiKeyEntry),
		lookupSecret: []byte(DefaultXApiKeyHeaderName),
	}

	for k, v := range m {
		apiKeyMap.fallback = append(apiKeyMap.fallback, apiKeyEntry{
			hash: k,
			user: v,
		})
	}

	return apiKeyMap, nil
}

func defineHeaderName(headerName string) string {
	if headerName == "" {
		return DefaultXApiKeyHeaderName
	}
	return headerName
}

// SingleUserApiKeyAuthenticationFunc authenticates the user based on the provided API key and user.
func SingleUserApiKeyAuthenticationFunc(apiKey, headerName string, user User) UserAuthenticationFunc {
	headerName = defineHeaderName(headerName)
	return func(r *http.Request) (User, error) {
		if apiKey != "" && r.Header.Get(headerName) != apiKey {
			return nil, ErrNotAuthorized
		}
		return user, nil
	}
}

// MultipleUserApiKeyAuthenticationFunc authenticates the user based on the provided map of API keys to users.
func MultipleUserApiKeyAuthenticationFunc(authData ApiKeyMap, headerName string) UserAuthenticationFunc {
	headerName = defineHeaderName(headerName)
	return func(r *http.Request) (User, error) {
		apiKeyHeaderValue := r.Header.Get(headerName)

		if apiKeyHeaderValue == "" {
			return nil, ErrNotAuthorized
		}

		if u, ok := authData.get(apiKeyHeaderValue); ok {
			return u, nil
		}
		return nil, ErrNotAuthorized
	}
}

// AuthenticationMiddleware is a middleware that authenticates the user through the given UserAuthenticationFunc
// and sets the user in the context.
func AuthenticationMiddleware(authFunc UserAuthenticationFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authFunc(r)
			if err != nil {
				if errors.Is(err, ErrNotAuthorized) {
					http.Error(w, ErrNotAuthorized.Error(), http.StatusUnauthorized)
				}
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if user == nil {
				http.Error(w, ErrNotAuthorized.Error(), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
		})
	}
}

// AuthenticatedUserFromContext returns the authenticated user from the context.
func AuthenticatedUserFromContext(ctx context.Context) User {
	value := ctx.Value(userKey)
	if value == nil {
		return nil
	}
	return value.(User)
}
