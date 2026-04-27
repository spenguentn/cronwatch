package notify

import (
	"context"
	"errors"
	"testing"
)

// recordNotifier captures the last call made to Notify.
type recordNotifier struct {
	called  bool
	subject string
	body    string
	err     error
}

func (r *recordNotifier) Notify(_ context.Context, subject, body string) error {
	r.called = true
	r.subject = subject
	r.body = body
	return r.err
}

func TestFilterNotifier_PassesWhenFilterTrue(t *testing.T) {
	rec := &recordNotifier{}
	fn := NewFilterNotifier(rec, func(_, _ string) bool { return true })

	if err := fn.Notify(context.Background(), "subject", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rec.called {
		t.Fatal("expected inner notifier to be called")
	}
}

func TestFilterNotifier_DropsWhenFilterFalse(t *testing.T) {
	rec := &recordNotifier{}
	fn := NewFilterNotifier(rec, func(_, _ string) bool { return false })

	if err := fn.Notify(context.Background(), "subject", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.called {
		t.Fatal("expected inner notifier NOT to be called")
	}
}

func TestFilterNotifier_NilFilterPassesAll(t *testing.T) {
	rec := &recordNotifier{}
	fn := NewFilterNotifier(rec, nil)

	_ = fn.Notify(context.Background(), "anything", "body")
	if !rec.called {
		t.Fatal("nil filter should allow all notifications")
	}
}

func TestFilterNotifier_PropagatesError(t *testing.T) {
	want := errors.New("send failed")
	rec := &recordNotifier{err: want}
	fn := NewFilterNotifier(rec, func(_, _ string) bool { return true })

	if got := fn.Notify(context.Background(), "s", "b"); !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestSubjectContainsFilter(t *testing.T) {
	f := SubjectContainsFilter("ERROR", "MISS")

	if !f("[ERROR] job failed", "") {
		t.Error("expected ERROR to pass")
	}
	if !f("job MISS detected", "") {
		t.Error("expected MISS to pass")
	}
	if f("[WARN] slight drift", "") {
		t.Error("expected WARN to be filtered")
	}
}

func TestSeverityFilter(t *testing.T) {
	f := SeverityFilter("[ERROR]", "[WARN]")

	if !f("[ERROR] something", "") {
		t.Error("expected [ERROR] prefix to pass")
	}
	if !f("[WARN] something", "") {
		t.Error("expected [WARN] prefix to pass")
	}
	if f("[INFO] something", "") {
		t.Error("expected [INFO] prefix to be filtered")
	}
	if f("no prefix", "") {
		t.Error("expected message without prefix to be filtered")
	}
}
