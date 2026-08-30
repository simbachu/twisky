package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestAccountMenu_OmitsWhenDisabled(t *testing.T) {
	t.Parallel()

	if ui.AccountMenu(ui.AccountMenuView{}) != nil {
		t.Fatal("AccountMenu(disabled) = node, want nil")
	}
}

func TestAccountMenu_RendersLoginWhenLoggedOut(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AccountMenu(ui.AccountMenuView{Enabled: true}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Account"`) {
		t.Fatalf("html = %q, want Account aria-label", html)
	}
	if !strings.Contains(html, `href="/oauth/login"`) {
		t.Fatalf("html = %q, want login href", html)
	}
	if !strings.Contains(html, "Log in") {
		t.Fatalf("html = %q, want Log in text", html)
	}
	if strings.Contains(html, "<details") {
		t.Fatalf("html = %q, want no details when logged out", html)
	}
}

func TestAccountMenu_RendersDetailsWithLogout(t *testing.T) {
	t.Parallel()

	current := ui.AuthorInfo{
		Handle:      "alice.test",
		DID:         "did:plc:alice",
		DisplayName: "Alice",
	}
	var buf bytes.Buffer
	if err := ui.AccountMenu(ui.AccountMenuView{
		Enabled: true,
		Current: &current,
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<details") {
		t.Fatalf("html = %q, want details", html)
	}
	if !strings.Contains(html, "<summary") {
		t.Fatalf("html = %q, want summary", html)
	}
	if !strings.Contains(html, `class="mini-profile"`) {
		t.Fatalf("html = %q, want mini-profile in summary", html)
	}
	if !strings.Contains(html, `class="iface-disclosure account-menu"`) {
		t.Fatalf("html = %q, want disclosure composed with account-menu", html)
	}
	if !strings.Contains(html, "Alice") {
		t.Fatalf("html = %q, want display name", html)
	}
	if !strings.Contains(html, `method="post"`) {
		t.Fatalf("html = %q, want logout POST form", html)
	}
	if !strings.Contains(html, `action="/oauth/logout"`) {
		t.Fatalf("html = %q, want logout action", html)
	}
	if !strings.Contains(html, "Log out") {
		t.Fatalf("html = %q, want Log out text", html)
	}
	if strings.Contains(html, "Switch account") {
		t.Fatalf("html = %q, want no switcher when Additional empty", html)
	}
	if !strings.Contains(html, `href="/oauth/login"`) {
		t.Fatalf("html = %q, want Add account href", html)
	}
	if !strings.Contains(html, "Add account") {
		t.Fatalf("html = %q, want Add account text", html)
	}
	settingsIdx := strings.Index(html, `href="/settings"`)
	logoutIdx := strings.Index(html, `action="/oauth/logout"`)
	if settingsIdx < 0 {
		t.Fatalf("html = %q, want Settings href", html)
	}
	if settingsIdx > logoutIdx {
		t.Fatalf("html = %q, want Settings before Log out", html)
	}
	if !strings.Contains(html, `class="account-menu-panel"`) {
		t.Fatalf("html = %q, want account-menu-panel", html)
	}
	if !strings.Contains(html, `class="account-menu-current"`) {
		t.Fatalf("html = %q, want current account in panel", html)
	}
	if strings.Count(html, "Alice") < 2 {
		t.Fatalf("html = %q, want current name in summary and panel", html)
	}
	if strings.Count(html, "@alice.test") < 2 {
		t.Fatalf("html = %q, want current handle in summary and panel", html)
	}
	if strings.Contains(html, `href="/alice.test"`) {
		t.Fatalf("html = %q, want no profile href on account chrome", html)
	}
}

func TestAccountMenu_RendersAdditionalSwitcher(t *testing.T) {
	t.Parallel()

	current := ui.AuthorInfo{Handle: "alice.test", DID: "did:plc:alice", DisplayName: "Alice"}
	var buf bytes.Buffer
	if err := ui.AccountMenu(ui.AccountMenuView{
		Enabled: true,
		Current: &current,
		Additional: []ui.AuthorInfo{
			{Handle: "bob.test", DID: "did:plc:bob", DisplayName: "Bob"},
		},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Switch account") {
		t.Fatalf("html = %q, want Switch account label", html)
	}
	if strings.Contains(html, `href="/bob.test"`) {
		t.Fatalf("html = %q, want no profile href on switcher chrome", html)
	}
	if strings.Contains(html, `href="/alice.test"`) {
		t.Fatalf("html = %q, want no profile href on summary chrome", html)
	}
	if !strings.Contains(html, "Bob") {
		t.Fatalf("html = %q, want additional account name", html)
	}
	if !strings.Contains(html, "@bob.test") {
		t.Fatalf("html = %q, want additional handle", html)
	}
	if !strings.Contains(html, `action="/oauth/switch"`) {
		t.Fatalf("html = %q, want switch action", html)
	}
	if !strings.Contains(html, `name="did"`) {
		t.Fatalf("html = %q, want did field", html)
	}
	if !strings.Contains(html, `value="did:plc:bob"`) {
		t.Fatalf("html = %q, want bob DID", html)
	}
}
