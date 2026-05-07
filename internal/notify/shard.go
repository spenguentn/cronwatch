package notify

import (
	"fmt"
	"hash/fnv"
)

// ShardNotifier routes messages to one of N notifiers based on a consistent
// hash of the message subject. This allows horizontal distribution of
// notifications across independent backends.
type ShardNotifier struct {
	shards []Notifier
}

// NewShardNotifier creates a ShardNotifier that distributes messages across
// the provided notifiers using consistent hashing on the message subject.
// At least one shard must be provided.
func NewShardNotifier(shards ...Notifier) (*ShardNotifier, error) {
	if len(shards) == 0 {
		return nil, fmt.Errorf("shard: at least one notifier required")
	}
	return &ShardNotifier{shards: shards}, nil
}

// Notify hashes the message subject to select a shard and forwards the
// message to that notifier. Nil shards at the selected index are skipped
// and the next non-nil shard is used.
func (s *ShardNotifier) Notify(ctx context.Context, msg Message) error {
	if len(s.shards) == 0 {
		return nil
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(msg.Subject))
	idx := int(h.Sum32()) % len(s.shards)

	// Walk forward to find a non-nil shard.
	for i := 0; i < len(s.shards); i++ {
		n := s.shards[(idx+i)%len(s.shards)]
		if n != nil {
			return n.Notify(ctx, msg)
		}
	}
	return nil
}

// ShardIndex returns the shard index that would be selected for the given
// subject without sending a notification. Useful for debugging.
func (s *ShardNotifier) ShardIndex(subject string) int {
	if len(s.shards) == 0 {
		return -1
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(subject))
	return int(h.Sum32()) % len(s.shards)
}
