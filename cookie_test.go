package ngebut

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCookieString(t *testing.T) {
	tests := []struct {
		name     string
		cookie   Cookie
		expected string
	}{
		{
			name: "Basic cookie",
			cookie: Cookie{
				Name:  "test",
				Value: "value",
			},
			expected: "test=value",
		},
		{
			name: "Cookie with path",
			cookie: Cookie{
				Name:  "test",
				Value: "value",
				Path:  "/path",
			},
			expected: "test=value; Path=/path",
		},
		{
			name: "Cookie with domain",
			cookie: Cookie{
				Name:   "test",
				Value:  "value",
				Domain: "example.com",
			},
			expected: "test=value; Domain=example.com",
		},
		{
			name: "Cookie with expiration",
			cookie: Cookie{
				Name:    "test",
				Value:   "value",
				Expires: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: "test=value; Expires=Sun, 01 Jan 2023 00:00:00 GMT",
		},
		{
			name: "Cookie with max age",
			cookie: Cookie{
				Name:   "test",
				Value:  "value",
				MaxAge: 3600,
			},
			expected: "test=value; Max-Age=3600",
		},
		{
			name: "Secure cookie",
			cookie: Cookie{
				Name:   "test",
				Value:  "value",
				Secure: true,
			},
			expected: "test=value; Secure",
		},
		{
			name: "HTTP only cookie",
			cookie: Cookie{
				Name:     "test",
				Value:    "value",
				HTTPOnly: true,
			},
			expected: "test=value; HttpOnly",
		},
		{
			name: "Cookie with SameSite",
			cookie: Cookie{
				Name:     "test",
				Value:    "value",
				SameSite: "Strict",
			},
			expected: "test=value; SameSite=Strict",
		},
		{
			name: "Partitioned cookie",
			cookie: Cookie{
				Name:        "test",
				Value:       "value",
				Partitioned: true,
			},
			expected: "test=value; Partitioned",
		},
		{
			name: "Session only cookie",
			cookie: Cookie{
				Name:        "test",
				Value:       "value",
				MaxAge:      3600,
				Expires:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				SessionOnly: true,
			},
			expected: "test=value",
		},
		{
			name: "Full cookie",
			cookie: Cookie{
				Name:        "test",
				Value:       "value",
				Path:        "/path",
				Domain:      "example.com",
				MaxAge:      3600,
				Expires:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Secure:      true,
				HTTPOnly:    true,
				SameSite:    "Strict",
				Partitioned: true,
			},
			expected: "test=value; Path=/path; Domain=example.com; Expires=Sun, 01 Jan 2023 00:00:00 GMT; Max-Age=3600; Secure; HttpOnly; SameSite=Strict; Partitioned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cookie.String()
			if result != tt.expected {
				t.Errorf("Cookie.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCtxCookie(t *testing.T) {
	tests := []struct {
		name   string
		cookie *Cookie
	}{
		{
			name: "Basic cookie",
			cookie: &Cookie{
				Name:  "test",
				Value: "value",
			},
		},
		{
			name: "Full cookie",
			cookie: &Cookie{
				Name:        "test",
				Value:       "value",
				Path:        "/path",
				Domain:      "example.com",
				MaxAge:      3600,
				Expires:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Secure:      true,
				HTTPOnly:    true,
				SameSite:    "Strict",
				Partitioned: true,
			},
		},
		{
			name:   "Nil cookie",
			cookie: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			ctx := GetContext(w, r)

			ctx.Cookie(tt.cookie)

			resp := w.Result()
			cookies := resp.Header.Values("Set-Cookie")

			if tt.cookie == nil {
				if len(cookies) > 0 {
					t.Errorf("Expected no cookies to be set for nil cookie, got %v", cookies)
				}
				return
			}

			if len(cookies) == 0 {
				t.Errorf("Expected cookie to be set, got none")
				return
			}

			expected := tt.cookie.String()
			if cookies[0] != expected {
				t.Errorf("Cookie header = %q, want %q", cookies[0], expected)
			}
		})
	}
}

func TestCtxGetCookie(t *testing.T) {
	tests := []struct {
		name         string
		cookieHeader string
		cookieName   string
		expected     string
	}{
		{
			name:         "Basic cookie",
			cookieHeader: "test=value",
			cookieName:   "test",
			expected:     "value",
		},
		{
			name:         "Multiple cookies",
			cookieHeader: "test1=value1; test2=value2; test3=value3",
			cookieName:   "test2",
			expected:     "value2",
		},
		{
			name:         "Cookie not found",
			cookieHeader: "test1=value1; test2=value2",
			cookieName:   "test3",
			expected:     "",
		},
		{
			name:         "Empty cookie header",
			cookieHeader: "",
			cookieName:   "test",
			expected:     "",
		},
		{
			name:         "Malformed cookie",
			cookieHeader: "test1=value1; test2; test3=value3",
			cookieName:   "test2",
			expected:     "",
		},
		{
			name:         "Cookie with empty name",
			cookieHeader: "test1=value1; =value2; test3=value3",
			cookieName:   "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			if tt.cookieHeader != "" {
				r.Header.Set("Cookie", tt.cookieHeader)
			}
			ctx := GetContext(w, r)

			value := ctx.Cookies(tt.cookieName)

			if value != tt.expected {
				t.Errorf("GetCookie(%q) = %q, want %q", tt.cookieName, value, tt.expected)
			}
		})
	}
}
