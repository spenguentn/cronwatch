// Package notify provides notification backends and middleware for
// cronwatch alerts.
//
// Backends
//
// The following backends are available:
//   - WebhookNotifier  – HTTP POST to an arbitrary URL.
//   - SlackNotifier    – Slack incoming-webhook messages.
//   - EmailNotifier    – SMTP email delivery.
//   - PagerDutyNotifier – PagerDuty Events API v2.
//
// Middleware
//
// Backends can be wrapped with middleware that modifies delivery behaviour:
//   - RetryNotifier   – retries failed deliveries with a fixed back-off.
//   - MultiNotifier   – fans a single notification out to many backends.
//   - FilterNotifier  – conditionally drops notifications based on a
//                       caller-supplied FilterFunc predicate.
//
// FilterFunc helpers
//
//   SubjectContainsFilter – passes notifications whose subject contains
//                           one of a set of substrings.
//   SeverityFilter        – passes notifications whose subject starts with
//                           one of a set of severity prefixes.
package notify
