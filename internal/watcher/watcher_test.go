package watcher

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"cronwatch/internal/alert"
	"cronwatch/internal/config"
	"cronwatch/internal/store"
)

func makeWatcher(t *testing.T, jobs []config.Job) (*Watcher, *store.Updater, *bytes.Buffer) {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "store-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	s, err := store.New(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	updater := store.NewUpdater(s)

	var buf bytes.Buffer
	notifier := alert.NewLogNotifier(&buf)
	disp := alert.NewDispatcher(notifier)

	cfg := &config.Config{Jobs: jobs}
	w := New(cfg, updater, disp, time.Hour)
	return w, updater, &buf
}

func TestCheck_NoLastRun_Warns(t *testing.T) {
	jobs := []config.Job{{Name: "backup", Schedule: "@hourly", DriftThreshold: 5 * time.Minute}}
	w, _, buf := makeWatcher(t, jobs)

	w.check(time.Now())

	if buf.Len() == 0 {
		t.Fatal("expected warning output, got none")
	}
}

func TestCheck_NoDrift_Silent(t *testing.T) {
	jobs := []config.Job{{Name: "sync", Schedule: "@hourly", DriftThreshold: 10 * time.Minute}}
	w, updater, buf := makeWatcher(t, jobs)

	now := time.Now()
	// Record a run that aligns perfectly with the schedule.
	if err := updater.Record("sync", now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	w.check(now)

	if buf.Len() != 0 {
		t.Fatalf("expected no output, got: %s", buf.String())
	}
}

func TestRun_CancelStops(t *testing.T) {
	w, _, _ := makeWatcher(t, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		w.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop after context cancellation")
	}
}
