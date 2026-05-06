// Package notify provides composable notification primitives for cronwatch.
//
// # TagNotifier
//
// TagNotifier enriches every outgoing [Message] with a fixed set of metadata
// key-value pairs before forwarding to an inner [Notifier].
//
// Tags are merged into a shallow copy of the message's Meta map so the caller's
// original map is never mutated. If a tag key already exists in the message
// metadata it is silently overwritten by the configured tag value, allowing
// TagNotifier to act as a metadata override stage in a pipeline.
//
// Typical use-cases:
//
//	// Stamp every alert with deployment context.
//	tagger := notify.NewTagNotifier(downstream, map[string]string{
//		"env":     "production",
//		"service": "cronwatch",
//		"region":  os.Getenv("AWS_REGION"),
//	})
//
// TagNotifier is safe for concurrent use; the internal tag map is copied at
// construction time and never modified afterwards.
package notify
