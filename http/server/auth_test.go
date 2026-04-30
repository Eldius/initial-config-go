package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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

func TestSingleUserApiKeyAuthenticationFunc(t *testing.T) {
	t.Run("given an empty header, should return unauthorized", func(t *testing.T) {
		mw := AuthenticationMiddleware(SingleUserApiKeyAuthenticationFunc("my-key", "x-api-key", testUser{id: "u1"}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("inner handler reached"))
		}))
		req := httptest.NewRequest("GET", "http://example.com", nil)
		w := httptest.NewRecorder()

		mw.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("given an auth header with the right api key, should return ok", func(t *testing.T) {
		mw := AuthenticationMiddleware(SingleUserApiKeyAuthenticationFunc("my-key", "x-api-key", testUser{id: "u1"}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("inner handler reached"))
		}))
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("x-api-key", "my-key")
		w := httptest.NewRecorder()

		mw.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("given an auth header with the wrong api key, should return unauthorized", func(t *testing.T) {
		mw := AuthenticationMiddleware(SingleUserApiKeyAuthenticationFunc("my-key", "x-api-key", testUser{id: "u1"}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("inner handler reached"))
		}))
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("x-api-key", "wrong-key")
		w := httptest.NewRecorder()

		mw.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("given a valid user, should return a success", func(t *testing.T) {
		mw := AuthenticationMiddleware(func(r *http.Request) (User, error) {
			return testUser{id: "user-1"}, nil
		})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("inner handler reached"))
		}))

		req := httptest.NewRequest("GET", "http://example.com", nil)
		w := httptest.NewRecorder()

		mw.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("given an invalid user, should return an error", func(t *testing.T) {
		mw := AuthenticationMiddleware(func(r *http.Request) (User, error) {
			return nil, nil
		})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("inner handler reached"))
		}))

		req := httptest.NewRequest("GET", "http://example.com", nil)
		w := httptest.NewRecorder()

		mw.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
