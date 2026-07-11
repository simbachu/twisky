package moderation

type PostSubject struct {
	AuthorDID    string
	Labels       []Label
	AuthorLabels []Label
	Quoted       *PostSubject
}

func ModeratePost(subject PostSubject, opts Options) Decision {
	decision := &Decision{authorDID: subject.AuthorDID}

	for _, label := range subject.Labels {
		decision.addLabel(LabelTargetContent, label, opts)
	}
	for _, label := range subject.AuthorLabels {
		decision.addLabel(LabelTargetAccount, label, opts)
	}

	if subject.Quoted != nil {
		quoted := ModeratePost(*subject.Quoted, opts)
		decision.merge(quoted)
	}

	return *decision
}

func (d *Decision) merge(other Decision) {
	d.causes = append(d.causes, other.causes...)
}
