// Package notify provides notification backends for cronwatch alerts.
//
// Available notifiers:
//
//   - WebhookNotifier  — HTTP POST to an arbitrary endpoint.
//   - SlackNotifier    — Slack incoming-webhook messages.
//   - EmailNotifier    — SMTP email delivery.
//   - PagerDutyNotifier — PagerDuty Events API v2 incidents.
//
// Composable wrappers:
//
//   - MultiNotifier   — fan-out to multiple backends.
//   - RetryNotifier   — automatic retry with back-off.
//   - FilterNotifier  — conditional delivery based on subject/severity.
//   - DedupeNotifier  — suppresses identical messages within a time window.
package notify
