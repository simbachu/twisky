package errorpage_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/errorpage"
	"github.com/simbachu/twisky/internal/components/ui"
)

func TestPage_RendersChromeAndContextualAlert(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := errorpage.Page(
		`Failed to resolve handle alice.bsky.social`,
		http.StatusBadGateway,
		"/alice.bsky.social",
		nil,
		ui.AccountMenuView{},
		"https://twisky.test",
	).Render(&buf)
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()

	for _, want := range []string{
		`id="page-alert"`,
		`role="alert"`,
		"Temporarily unavailable",
		"Failed to resolve handle alice.bsky.social",
		`aria-label="Site"`,
		`rel="canonical" href="https://twisky.test/alice.bsky.social"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}
}

func TestPage_EscapesUserControlledMessage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := errorpage.Page(
		`<script>alert(1)</script>`,
		http.StatusBadRequest,
		"/oops",
		nil,
		ui.AccountMenuView{},
		"",
	).Render(&buf)
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if strings.Contains(html, `<script>alert(1)</script>`) {
		t.Fatalf("html left script unescaped; got:\n%s", html)
	}
	if !strings.Contains(html, `&lt;script&gt;alert(1)&lt;/script&gt;`) {
		t.Fatalf("html missing escaped message; got:\n%s", html)
	}
}

func TestAlertFragment_IsOutOfBandPageAlert(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := errorpage.AlertFragment("Could not find handle missing.example", http.StatusNotFound).Render(&buf)
	if err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`id="page-alert"`,
		`hx-swap-oob="true"`,
		`role="alert"`,
		"Not found",
		"Could not find handle missing.example",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}
	if strings.Contains(html, `<html`) {
		t.Fatalf("fragment unexpectedly wrapped in full document; got:\n%s", html)
	}
}
