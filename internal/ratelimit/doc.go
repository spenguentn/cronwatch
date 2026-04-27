// Package ratelimit implements a per-job token-bucket rate limiter for
// cronwatch alert notifications.
//
// When many jobs drift or are missed simultaneously, the rate limiter prevents
// an alert storm by capping how frequently notifications are dispatched for
// each individual job key.
//
// # Usage
//
//	limiter := ratelimit.New(5*time.Minute, 3)
//
//	// Wrap any notify.Notifier:
//	guarded := ratelimit.NewNotifier(slackNotifier, limiter, job.Name)
//
// The Limiter itself is safe for concurrent use. Each job key maintains an
// independent token bucket, so a noisy job does not suppress alerts for
// healthy jobs.
package ratelimit
