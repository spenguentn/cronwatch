# cronwatch

Lightweight daemon that monitors cron job execution times and alerts on drift or missed runs.

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && go build ./...
```

## Usage

Define your monitored jobs in a config file:

```yaml
# cronwatch.yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    tolerance: 5m
    alert: slack

  - name: hourly-sync
    schedule: "0 * * * *"
    tolerance: 2m
    alert: email
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap an existing cron job to report execution:

```bash
# In your crontab
0 2 * * * cronwatch run --job daily-backup -- /usr/local/bin/backup.sh
```

cronwatch will alert you if a job runs significantly late, finishes too quickly, or stops running altogether.

## Alerts

Supported alert targets: `slack`, `email`, `pagerduty`, `webhook`

Set credentials via environment variables or the `alerts` block in your config file.

## License

MIT © yourname