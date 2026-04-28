package server

import (
	"net/http"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type testUser struct {
	id string
}

func (u testUser) UserID() string {
	return u.id
}

func (u testUser) UserData() map[string]any {
	return map[string]any{"id": u.id}
}

func TestMultipleUserApiKeyAuthenticationFunc_FromPlainMap(t *testing.T) {
	t.Parallel()

	authData, err := NewApiKeyMapFromPlainMap(map[string]User{
		"key-1": testUser{id: "u1"},
		"key-2": testUser{id: "u2"},
	})
	if err != nil {
		t.Fatalf("unexpected error creating auth data: %v", err)
	}

	authFn := MultipleUserApiKeyAuthenticationFunc(authData, "")

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	req.Header.Set(DefaultXApiKeyHeaderName, "key-2")

	u, err := authFn(req)
	if err != nil {
		t.Fatalf("expected successful authentication, got error: %v", err)
	}
	if u == nil || u.UserID() != "u2" {
		t.Fatalf("unexpected user returned: %#v", u)
	}
}

func TestMultipleUserApiKeyAuthenticationFunc_FromHashedMapFallback(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("plain-key"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("unexpected error generating hash: %v", err)
	}

	authData, err := NewApiKeyMapFromHashedMap(map[string]User{
		string(hash): testUser{id: "hashed-user"},
	})
	if err != nil {
		t.Fatalf("unexpected error creating auth data: %v", err)
	}

	authFn := MultipleUserApiKeyAuthenticationFunc(authData, "")

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	req.Header.Set(DefaultXApiKeyHeaderName, "plain-key")

	u, err := authFn(req)
	if err != nil {
		t.Fatalf("expected successful authentication, got error: %v", err)
	}
	if u == nil || u.UserID() != "hashed-user" {
		t.Fatalf("unexpected user returned: %#v", u)
	}
}

func TestMultipleUserApiKeyAuthenticationFunc_InvalidKey(t *testing.T) {
	t.Parallel()

	authData, err := NewApiKeyMapFromPlainMap(map[string]User{
		"valid-key": testUser{id: "u1"},
	})
	if err != nil {
		t.Fatalf("unexpected error creating auth data: %v", err)
	}

	authFn := MultipleUserApiKeyAuthenticationFunc(authData, "")

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	req.Header.Set(DefaultXApiKeyHeaderName, "invalid-key")

	u, err := authFn(req)
	if err == nil {
		t.Fatalf("expected authentication error for invalid key, got user: %#v", u)
	}
	if u != nil {
		t.Fatalf("expected nil user for invalid key, got: %#v", u)
	}
}
