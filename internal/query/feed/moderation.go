package feed

import (
	"context"

	"github.com/simbachu/twisky/internal/moderation"
)

func ApplyModeration(ctx context.Context, provider moderation.PrefsProvider, feed FeedView, uiContext moderation.UIContext) FeedView {
	posts := make([]PostView, 0, len(feed.Posts))
	for _, post := range feed.Posts {
		moderated := applyModerationToPostView(ctx, provider, post, uiContext)
		if moderated.Moderation.Filtered {
			continue
		}
		posts = append(posts, moderated)
	}
	feed.Posts = posts
	return feed
}

func ApplyModerationToPostPage(ctx context.Context, provider moderation.PrefsProvider, page PostPageView) PostPageView {
	page.Ancestors = applyModerationToPostViews(ctx, provider, page.Ancestors, moderation.UIContextContentView)
	page.Post = applyModerationToPostView(ctx, provider, page.Post, moderation.UIContextContentView)
	page.Replies = applyModerationToThreadNodes(ctx, provider, page.Replies)
	return page
}

func applyModerationToPostViews(ctx context.Context, provider moderation.PrefsProvider, posts []PostView, uiContext moderation.UIContext) []PostView {
	out := make([]PostView, 0, len(posts))
	for _, post := range posts {
		out = append(out, applyModerationToPostView(ctx, provider, post, uiContext))
	}
	return out
}

func applyModerationToThreadNodes(ctx context.Context, provider moderation.PrefsProvider, nodes []ThreadNodeView) []ThreadNodeView {
	out := make([]ThreadNodeView, 0, len(nodes))
	for _, node := range nodes {
		if node.Unavailable {
			out = append(out, node)
			continue
		}
		node.Post = applyModerationToPostView(ctx, provider, node.Post, moderation.UIContextContentView)
		node.Replies = applyModerationToThreadNodes(ctx, provider, node.Replies)
		out = append(out, node)
	}
	return out
}

func applyModerationToPostView(ctx context.Context, provider moderation.PrefsProvider, view PostView, uiContext moderation.UIContext) PostView {
	if view.ReplyParentMaybe != nil {
		parent := applyModerationToPostView(ctx, provider, *view.ReplyParentMaybe, uiContext)
		if parent.Moderation.Filtered {
			view.ReplyParentMaybe = nil
		} else {
			view.ReplyParentMaybe = &parent
		}
	}
	if view.QuotedPostMaybe != nil {
		quoted := applyModerationToPostView(ctx, provider, *view.QuotedPostMaybe, uiContext)
		if quoted.Moderation.Filtered {
			view.QuotedPostMaybe = nil
		} else {
			view.QuotedPostMaybe = &quoted
		}
	}

	listUI := moderation.EvaluatePostView(ctx, provider, view.authorDID, view.labels, view.authorLabels, quotedSubject(view.QuotedPostMaybe), uiContext)
	mediaUI := moderation.EvaluatePostView(ctx, provider, view.authorDID, view.labels, view.authorLabels, quotedSubject(view.QuotedPostMaybe), moderation.UIContextContentMedia)

	view.Moderation = ModerationView{
		Filtered:   listUI.Filter,
		Blurred:    listUI.Blur,
		NoOverride: listUI.NoOverride || mediaUI.NoOverride,
		AlertText:  coalesceMessage(listUI, mediaUI),
		BlurMedia:  mediaUI.BlurMedia,
	}
	if view.Moderation.AlertText == "" && (listUI.Alert || listUI.Inform) {
		view.Moderation.AlertText = listUI.PrimaryMessage()
	}
	return view
}

func quotedSubject(view *PostView) *moderation.PostSubject {
	if view == nil {
		return nil
	}
	subject := postSubjectFromView(*view)
	return &subject
}

func postSubjectFromView(view PostView) moderation.PostSubject {
	subject := moderation.PostSubject{
		AuthorDID:    view.authorDID,
		Labels:       view.labels,
		AuthorLabels: view.authorLabels,
	}
	if view.QuotedPostMaybe != nil {
		quoted := postSubjectFromView(*view.QuotedPostMaybe)
		subject.Quoted = &quoted
	}
	return subject
}

func coalesceMessage(listUI, mediaUI moderation.UIResult) string {
	if message := listUI.PrimaryMessage(); message != "" {
		return message
	}
	return mediaUI.PrimaryMessage()
}
