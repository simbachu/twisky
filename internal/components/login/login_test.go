package login_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/login"
)

func TestPage_IsChromeLessLoginWithBluesky(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := login.Page("", "https://twisky.test", "/", false).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()

	for _, want := range []string{
		"Login with Bluesky",
		`action="/oauth/login"`,
		`name="username"`,
		`aria-label="Twisky"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}
	for _, unwanted := range []string{
		`aria-label="Site"`,
		`aria-label="Account"`,
		"You might like:",
	} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("html unexpectedly contains %q", unwanted)
		}
	}
}

func TestPage_ShowsErrorAlert(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := login.Page("Handle or DID is required.", "", "/oauth/login", false).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `role="alert"`) {
		t.Fatalf("html missing alert; got:\n%s", html)
	}
	if !strings.Contains(html, "Handle or DID is required.") {
		t.Fatalf("html missing error message; got:\n%s", html)
	}
}

func TestPage_AddAnotherAccountCopy(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := login.Page("", "https://twisky.test", "/oauth/login", true).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, "Add another account") {
		t.Fatalf("html missing add-account heading; got:\n%s", html)
	}
	if !strings.Contains(html, `action="/oauth/login"`) {
		t.Fatalf("html missing login form; got:\n%s", html)
	}
	if strings.Contains(html, `aria-label="Site"`) {
		t.Fatalf("html unexpectedly contains site nav")
	}
}
