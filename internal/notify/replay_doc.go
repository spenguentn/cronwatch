// Package notify provides composable notifier primitives for cronwatch.
//
// # ReplayNotifier
//
// ReplayNotifier is a buffering notifier that records every inbound message
// without forwarding it immediately. Callers drive delivery explicitly by
// calling Replay, which attempts to send all buffered messages to a target
// Notifier in the order they were received.
//
// Typical use-cases:
//
//   - Recovery after a transient downstream outage: buffer alerts while the
//     destination is unavailable, then replay once it recovers.
//   - Testing: capture messages in-process and assert on their content before
//     deciding whether to forward them.
//
// Buffer management:
//
// The buffer is bounded by the capacity passed to NewReplayNotifier (default
// 100). When the buffer is full the oldest message is silently dropped to make
// room for the newest, preserving recency.
//
// After a Replay call, only messages whose delivery failed are kept in the
// buffer; successfully delivered messages are removed. This makes it safe to
// call Replay repeatedly until the buffer drains.
//
// Concurrency:
//
// ReplayNotifier is safe for concurrent use by multiple goroutines.
package notify
