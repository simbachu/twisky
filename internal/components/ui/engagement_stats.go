package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// EngagementStatRow is one labeled exact count in an EngagementStats list.
type EngagementStatRow struct {
	Label string
	ID    string
	Count int
}

// EngagementStatsConfig describes the focused-post engagement stats disclosure.
type EngagementStatsConfig struct {
	ID        string
	StatsHref string
	Open      bool
	Summary   g.Node
	Rows      []EngagementStatRow
}

// EngagementStats renders exact reply/repost/like counts in a Disclosure.
func EngagementStats(cfg EngagementStatsConfig) g.Node {
	rows := make([]g.Node, len(cfg.Rows))
	for i, row := range cfg.Rows {
		rows[i] = Div(
			Dt(g.Text(row.Label)),
			GroupedStatCount(row.ID, row.Count, false),
		)
	}
	attrs := []g.Node{}
	if cfg.StatsHref != "" {
		attrs = append(attrs, g.Attr("data-stats-href", cfg.StatsHref))
	}
	return DisclosureWith(DisclosureConfig{
		ExtraClass: "post-engagement-stats",
		Summary:    cfg.Summary,
		Content:    []g.Node{Dl(g.Group(rows))},
		ID:         cfg.ID,
		Open:       cfg.Open,
		Attrs:      attrs,
	})
}
