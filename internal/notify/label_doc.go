// Package notify provides composable notification primitives for cronwatch.
//
// # LabelNotifier
//
// LabelNotifier enriches outgoing messages with a fixed set of key/value
// label pairs written into the message's Meta map. It is useful for
// stamping every alert with environment-level metadata (e.g. env, region,
// service) without requiring callers to populate Meta manually.
//
// Labels are applied non-destructively: if a key already exists in
// msg.Meta it will not be overwritten, so per-message context always
// takes precedence over static labels.
//
// # Usage
//
//	 labels := map[string]string{
//	     "env":    "production",
//	     "region": "us-east-1",
//	 }
//	 n := notify.NewLabelNotifier(webhookNotifier, labels)
//	 n.Notify(ctx, msg)
package notify
