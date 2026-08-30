package settings

import (
	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	settingsquery "github.com/simbachu/twisky/internal/query/settings"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Settings(view settingsquery.SettingsView, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, publicBaseURL string) g.Node {
	return page.Page(
		page.PageMeta{
			Title:        "Settings · Twisky",
			Description:  "Account settings on Twisky.",
			CanonicalURL: page.AbsoluteURL(publicBaseURL, "/settings"),
			Path:         "/settings",
			OGType:       "website",
		},
		suggested,
		accounts,
		Article(
			g.Attr("class", "settings"),
			H1(g.Text("Settings")),
			accountSection(view),
			ContentFilteringSection(view),
			feedsSection(view),
			ThreadingSection(view),
			signOutSection(),
		),
	)
}

func accountSection(view settingsquery.SettingsView) g.Node {
	author := ui.AuthorInfo{
		Handle:      view.Handle,
		DisplayName: view.DisplayName,
		DID:         view.DID,
		Avatar:      view.Avatar,
	}
	profilePath := actor.ProfilePath(view.Handle, view.DID)
	return Section(
		g.Attr("id", "settings-account"),
		H2(g.Text("Account")),
		ui.MiniProfile(author, ui.AuthorLinked),
		P(A(g.Attr("href", profilePath), g.Text("View profile"))),
	)
}

func ContentFilteringSection(view settingsquery.SettingsView) g.Node {
	return Section(
		g.Attr("id", "settings-content-filtering"),
		H2(g.Text("Content filtering")),
		Form(
			g.Attr("method", "post"),
			g.Attr("action", "/settings/content-filtering"),
			g.Attr("hx-post", "/settings/content-filtering"),
			g.Attr("hx-target", "closest section"),
			g.Attr("hx-swap", "outerHTML"),
			g.Attr("hx-disabled-elt", "this"),
			Label(
				Input(
					g.Attr("type", "checkbox"),
					g.Attr("name", "adult_content"),
					g.Attr("value", "1"),
					g.If(view.AdultContentEnabled, g.Attr("checked", "")),
				),
				g.Text(" Enable adult content"),
			),
			FieldSet(
				Legend(g.Text("Content labels")),
				g.Group(g.Map(view.Labels, labelSelect)),
			),
			Button(g.Attr("type", "submit"), g.Text("Save")),
		),
	)
}

func labelSelect(setting settingsquery.LabelSetting) g.Node {
	id := "label-" + setting.Identifier
	return Div(
		g.Attr("class", "settings-field"),
		Label(g.Attr("for", id), g.Text(setting.Name)),
		Select(
			g.Attr("id", id),
			g.Attr("name", setting.Identifier),
			visibilityOption("hide", "Hide", setting.Value),
			visibilityOption("warn", "Warn", setting.Value),
			visibilityOption("ignore", "Show", setting.Value),
		),
	)
}

func visibilityOption(value, label string, current bluesky.LabelVisibility) g.Node {
	return Option(
		g.Attr("value", value),
		g.If(string(current) == value, g.Attr("selected", "")),
		g.Text(label),
	)
}

func feedsSection(view settingsquery.SettingsView) g.Node {
	var items g.Node
	if len(view.Feeds) == 0 {
		items = P(g.Text("No saved feeds."))
	} else {
		items = Ul(g.Group(g.Map(view.Feeds, feedItem)))
	}
	return Section(
		g.Attr("id", "settings-feeds"),
		H2(g.Text("Feeds")),
		items,
	)
}

func feedItem(feed settingsquery.FeedSetting) g.Node {
	status := "Saved"
	if feed.Pinned {
		status = "Pinned"
	}
	return Li(
		g.Text(feed.Label + " · " + status),
	)
}

func ThreadingSection(view settingsquery.SettingsView) g.Node {
	return Section(
		g.Attr("id", "settings-threading"),
		H2(g.Text("Threading")),
		Form(
			g.Attr("method", "post"),
			g.Attr("action", "/settings/threading"),
			g.Attr("hx-post", "/settings/threading"),
			g.Attr("hx-target", "closest section"),
			g.Attr("hx-swap", "outerHTML"),
			g.Attr("hx-disabled-elt", "this"),
			Div(
				g.Attr("class", "settings-field"),
				Label(
					g.Attr("for", "thread-sort"),
					g.Text("Sort order"),
				),
				Select(
					g.Attr("id", "thread-sort"),
					g.Attr("name", "sort"),
					sortOption(bluesky.ThreadSortHotness, "Hot", view.ThreadSort),
					sortOption(bluesky.ThreadSortNewest, "Newest", view.ThreadSort),
					sortOption(bluesky.ThreadSortOldest, "Oldest", view.ThreadSort),
					sortOption(bluesky.ThreadSortMostLikes, "Most liked", view.ThreadSort),
				),
			),
			Label(
				Input(
					g.Attr("type", "checkbox"),
					g.Attr("name", "prioritize_followed"),
					g.Attr("value", "1"),
					g.If(view.PrioritizeFollowedUsers, g.Attr("checked", "")),
				),
				g.Text(" Prioritize people you follow"),
			),
			Button(g.Attr("type", "submit"), g.Text("Save")),
		),
	)
}

func sortOption(value, label, current string) g.Node {
	return Option(
		g.Attr("value", value),
		g.If(current == value, g.Attr("selected", "")),
		g.Text(label),
	)
}

func signOutSection() g.Node {
	return Section(
		g.Attr("id", "settings-sign-out"),
		g.Attr("class", "destructive"),
		H2(g.Text("Sign out")),
		P(g.Text("Sign out of Twisky on this device. Your Bluesky account stays intact.")),
		Form(
			g.Attr("method", "post"),
			g.Attr("action", "/oauth/logout"),
			Button(g.Attr("type", "submit"), g.Text("Log out")),
		),
	)
}

func ContentFilteringFragment(prefs bluesky.Preferences) g.Node {
	return ContentFilteringSection(settingsquery.ViewFromPreferences(prefs))
}

// ThreadingFragment renders the threading section after a save.
func ThreadingFragment(prefs bluesky.Preferences) g.Node {
	return ThreadingSection(settingsquery.ViewFromPreferences(prefs))
}
