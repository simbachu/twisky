package moderation

import (
	"context"

	"github.com/simbachu/twisky/internal/bluesky"
)

func PostSubjectFromBluesky(post bluesky.Post) PostSubject {
	subject := PostSubject{
		AuthorDID: post.Author.DID,
		Labels:    labelsFromBluesky(post.AllLabels()),
		AuthorLabels: labelsFromBluesky(post.Author.Labels),
	}
	if post.Embed != nil {
		if quoted := post.Embed.QuotedPost(); quoted != nil {
			quotedSubject := PostSubjectFromBluesky(*quoted)
			subject.Quoted = &quotedSubject
		}
	}
	return subject
}

func labelsFromBluesky(labels []bluesky.Label) []Label {
	if len(labels) == 0 {
		return nil
	}
	out := make([]Label, 0, len(labels))
	for _, label := range labels {
		out = append(out, Label{Val: label.Val, Src: label.Src})
	}
	return out
}

func EvaluatePost(ctx context.Context, provider PrefsProvider, post bluesky.Post, uiContext UIContext) UIResult {
	opts := Options{
		Prefs: provider.Prefs(ctx),
	}
	decision := ModeratePost(PostSubjectFromBluesky(post), opts)
	return decision.UI(uiContext)
}

func EvaluatePostView(ctx context.Context, provider PrefsProvider, authorDID string, labels, authorLabels []Label, quoted *PostSubject, uiContext UIContext) UIResult {
	subject := PostSubject{
		AuthorDID:    authorDID,
		Labels:       labels,
		AuthorLabels: authorLabels,
		Quoted:       quoted,
	}
	opts := Options{
		Prefs: provider.Prefs(ctx),
	}
	return ModeratePost(subject, opts).UI(uiContext)
}
