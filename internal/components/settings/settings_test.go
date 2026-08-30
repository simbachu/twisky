package settings_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	settingspage "github.com/simbachu/twisky/internal/components/settings"
	"github.com/simbachu/twisky/internal/components/ui"
	settingsquery "github.com/simbachu/twisky/internal/query/settings"
)

func sampleView() settingsquery.SettingsView {
	return settingsquery.SettingsView{
		DID:                 "did:plc:alice",
		Handle:              "alice.test",
		DisplayName:         "Alice",
		AdultContentEnabled: true,
		Labels: []settingsquery.LabelSetting{
			{Identifier: "porn", Name: "Pornography", Value: bluesky.LabelHide},
			{Identifier: "sexual", Name: "Suggestive content", Value: bluesky.LabelWarn},
			{Identifier: "nudity", Name: "Non-sexual nudity", Value: bluesky.LabelIgnore},
			{Identifier: "graphic-media", Name: "Graphic media", Value: bluesky.LabelHide},
		},
		Feeds: []settingsquery.FeedSetting{
			{Label: "For You", URI: "at://did:plc:feeds/app.bsky.feed.generator/for-you", Pinned: true, Type: "feed"},
		},
		ThreadSort:              bluesky.ThreadSortNewest,
		PrioritizeFollowedUsers: true,
	}
}

func TestSettings_SectionOrderAndHypermedia(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := settingspage.Settings(sampleView(), nil, ui.AccountMenuView{Enabled: true}, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()

	for _, want := range []string{
		`<a href="/settings" aria-current="page">`,
		">Settings</h1>",
		`id="settings-account"`,
		`id="settings-content-filtering"`,
		`id="settings-feeds"`,
		`id="settings-threading"`,
		`id="settings-sign-out"`,
		`action="/settings/content-filtering"`,
		`hx-post="/settings/content-filtering"`,
		`hx-target="closest section"`,
		`action="/settings/threading"`,
		`hx-post="/settings/threading"`,
		`action="/oauth/logout"`,
		"Log out",
		"For You · Pinned",
		`href="/alice.test"`,
		`class="mini-profile"`,
		"Enable adult content",
		"Prioritize people you follow",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}

	account := strings.Index(html, `id="settings-account"`)
	filtering := strings.Index(html, `id="settings-content-filtering"`)
	feeds := strings.Index(html, `id="settings-feeds"`)
	threading := strings.Index(html, `id="settings-threading"`)
	signOut := strings.Index(html, `id="settings-sign-out"`)
	logout := strings.LastIndex(html, `action="/oauth/logout"`)
	if account < 0 || filtering < 0 || feeds < 0 || threading < 0 || signOut < 0 {
		t.Fatalf("missing section ids; html:\n%s", html)
	}
	if !(account < filtering && filtering < feeds && feeds < threading && threading < signOut) {
		t.Fatalf("section order wrong: account=%d filtering=%d feeds=%d threading=%d signOut=%d", account, filtering, feeds, threading, signOut)
	}
	if logout < signOut {
		t.Fatalf("logout form should be in the last section")
	}
	if strings.Count(html, `aria-current="page"`) != 1 {
		t.Fatalf("want exactly one aria-current; got:\n%s", html)
	}
}

func TestContentFilteringFragment_SwapsSection(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(nil)
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	prefs.AdultContentEnabled = true
	var buf bytes.Buffer
	if err := settingspage.ContentFilteringFragment(prefs).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `id="settings-content-filtering"`) {
		t.Fatalf("html = %q, want content-filtering section", html)
	}
	if strings.Contains(html, "<html") {
		t.Fatalf("fragment should not be a full document")
	}
	if !strings.Contains(html, `name="adult_content"`) || !strings.Contains(html, "checked") {
		t.Fatalf("html = %q, want adult content checked", html)
	}
}
