package notify

import (
	"strings"
	"testing"
)

func TestCombineBatch_SingleMessage(t *testing.T) {
	msgs := []Message{
		{Subject: "only", Body: "the body", Severity: SeverityWarn},
	}
	out := combineBatch(msgs)
	if out.Subject != "only" {
		t.Errorf("single message subject should pass through, got %q", out.Subject)
	}
	if out.Body != "the body" {
		t.Errorf("single message body should pass through, got %q", out.Body)
	}
}

func TestCombineBatch_MultipleMessages(t *testing.T) {
	msgs := []Message{
		{Subject: "alpha", Body: "body-a", Severity: SeverityWarn},
		{Subject: "beta", Body: "body-b", Severity: SeverityError},
		{Subject: "gamma", Body: "body-c", Severity: SeverityWarn},
	}
	out := combineBatch(msgs)
	if !strings.Contains(out.Subject, "3") {
		t.Errorf("combined subject should mention count, got %q", out.Subject)
	}
	for _, s := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(out.Subject, s) {
			t.Errorf("combined subject should contain %q, got %q", s, out.Subject)
		}
	}
	for _, s := range []string{"body-a", "body-b", "body-c"} {
		if !strings.Contains(out.Body, s) {
			t.Errorf("combined body should contain %q, got %q", s, out.Body)
		}
	}
}

func TestCombineBatch_PreservesFirstSeverity(t *testing.T) {
	msgs := []Message{
		{Subject: "a", Severity: SeverityError},
		{Subject: "b", Severity: SeverityWarn},
	}
	out := combineBatch(msgs)
	if out.Severity != SeverityError {
		t.Errorf("expected severity %q, got %q", SeverityError, out.Severity)
	}
}
