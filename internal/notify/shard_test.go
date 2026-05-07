package notify

import (
	"context"
	"errors"
	"testing"
)

func TestShardNotifier_RequiresAtLeastOneShard(t *testing.T) {
	_, err := NewShardNotifier()
	if err == nil {
		t.Fatal("expected error for empty shards")
	}
}

func TestShardNotifier_ConsistentRouting(t *testing.T) {
	var got [3][]string
	shards := make([]Notifier, 3)
	for i := range shards {
		i := i
		shards[i] = NotifierFunc(func(_ context.Context, msg Message) error {
			got[i] = append(got[i], msg.Subject)
			return nil
		})
	}

	sn, err := NewShardNotifier(shards...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subjects := []string{"job-a", "job-b", "job-c", "job-a", "job-b"}
	for _, subj := range subjects {
		_ = sn.Notify(context.Background(), Message{Subject: subj})
	}

	// job-a and job-b should land on the same shard each time.
	for i := 0; i < 3; i++ {
		for _, subj := range got[i] {
			if subj != got[i][0] {
				t.Errorf("shard %d received inconsistent subjects: %v", i, got[i])
			}
		}
	}
}

func TestShardNotifier_NilShardSkipped(t *testing.T) {
	received := false
	real := NotifierFunc(func(_ context.Context, _ Message) error {
		received = true
		return nil
	})

	// Force nil at index 0 by providing nil first; real notifier at index 1.
	sn := &ShardNotifier{shards: []Notifier{nil, real}}
	// Use a subject that hashes to shard 0.
	for subj := "a"; ; subj += "a" {
		if sn.ShardIndex(subj) == 0 {
			_ = sn.Notify(context.Background(), Message{Subject: subj})
			break
		}
		if len(subj) > 64 {
			t.Skip("could not find subject hashing to shard 0")
		}
	}

	if !received {
		t.Error("expected fallback to non-nil shard")
	}
}

func TestShardNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("shard error")
	sn, _ := NewShardNotifier(
		NotifierFunc(func(_ context.Context, _ Message) error { return sentinel }),
	)
	err := sn.Notify(context.Background(), Message{Subject: "x"})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestShardNotifier_ShardIndex_Deterministic(t *testing.T) {
	sn, _ := NewShardNotifier(
		NotifierFunc(func(_ context.Context, _ Message) error { return nil }),
		NotifierFunc(func(_ context.Context, _ Message) error { return nil }),
		NotifierFunc(func(_ context.Context, _ Message) error { return nil }),
	)
	subject := "deterministic-job"
	idx1 := sn.ShardIndex(subject)
	idx2 := sn.ShardIndex(subject)
	if idx1 != idx2 {
		t.Errorf("ShardIndex not deterministic: %d != %d", idx1, idx2)
	}
}

func TestShardNotifier_ShardIndex_EmptyShards(t *testing.T) {
	sn := &ShardNotifier{}
	if got := sn.ShardIndex("x"); got != -1 {
		t.Errorf("expected -1 for empty shards, got %d", got)
	}
}
