package notify

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
)

// TemplateNotifier renders a subject/body template before forwarding to an
// inner Notifier. Template data is the Message itself, so fields like
// .Subject, .Body, and .Meta are all accessible.
type TemplateNotifier struct {
	inner    Notifier
	subjectT *template.Template
	bodyT    *template.Template
}

// NewTemplateNotifier creates a TemplateNotifier. subjectTmpl and bodyTmpl are
// Go text/template strings. Pass an empty string to leave that field unchanged.
func NewTemplateNotifier(inner Notifier, subjectTmpl, bodyTmpl string) (*TemplateNotifier, error) {
	if inner == nil {
		return nil, fmt.Errorf("notify: template: inner notifier must not be nil")
	}

	var st, bt *template.Template
	var err error

	if subjectTmpl != "" {
		st, err = template.New("subject").Parse(subjectTmpl)
		if err != nil {
			return nil, fmt.Errorf("notify: template: invalid subject template: %w", err)
		}
	}

	if bodyTmpl != "" {
		bt, err = template.New("body").Parse(bodyTmpl)
		if err != nil {
			return nil, fmt.Errorf("notify: template: invalid body template: %w", err)
		}
	}

	return &TemplateNotifier{inner: inner, subjectT: st, bodyT: bt}, nil
}

// Notify renders the templates against msg and forwards the result.
func (t *TemplateNotifier) Notify(ctx context.Context, msg Message) error {
	if t.subjectT != nil {
		var buf bytes.Buffer
		if err := t.subjectT.Execute(&buf, msg); err != nil {
			return fmt.Errorf("notify: template: render subject: %w", err)
		}
		msg.Subject = buf.String()
	}

	if t.bodyT != nil {
		var buf bytes.Buffer
		if err := t.bodyT.Execute(&buf, msg); err != nil {
			return fmt.Errorf("notify: template: render body: %w", err)
		}
		msg.Body = buf.String()
	}

	return t.inner.Notify(ctx, msg)
}
