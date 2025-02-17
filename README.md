# 📅 dotcal - An ultralight calendar aggregator and scheduler

![Coverage](https://img.shields.io/badge/Coverage-83.1%25-brightgreen)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

A Self-hosted `.ics` feed aggregator that publishes an anonymized public schedule to a GitHub repository. A light weight, ultra simplified, copyleft alternative to [cal.com](https://cal.com).

You can see an exampled output at [zachatrocity/cal](https://github.com/zachatrocity/cal).

<p align="center">
  <img src="https://gist.githubusercontent.com/zachatrocity/e0246929ef65bb738bcf7a74c42b1bbf/raw/03eacfef248a275d915c314c295b673c6b1c4f7d/IMG_0291.jpeg" alt="dotcal screenshot"/>
</p>

> ⚠️ **Development Status** ⚠️: This project is in active development. While dotcal never modifies your calendar data (it only reads from ICS feeds), it's recommended to maintain backups of any important GitHub repositories you publish to, as this project's behavior may change during development.

## ✨ Features

- 🔄 Aggregates multiple ICS calendar feeds into pure markdown
- 🔒 Anonymizes event details
- 📝 Publishes formatted schedule to public repo
- ⚡ Automatic syncing via cron

## 🚀 Setup

1. Create a GitHub repository for your public calendar
2. Setup SSH key on host with repository access
3. Create `docker-compose.yml` (see [docker-compose.yaml](/docker-compose.yml) for all options):

```yaml
version: '3.8'
services:
  dotcal:
    image: ghcr.io/zachatrocity/dotcal:latest
    environment:
      - GITHUB_REPO=git@github.com:username/calendar.git
      - ICS_FEEDS=https://calendar.google.com/calendar/ical/example/basic.ics
      - TIMEZONE=America/Boise
    volumes:
      - ~/.ssh:/root/.ssh:ro
    restart: unless-stopped
```

4. Start the service:
```bash
docker-compose up -d
```

## 📋 Schedule Format

Schedules are published to your GitHub repository:
- `README.md` - Current week
- `future/YYYY-WXX.md` - Upcoming weeks
- `past/YYYY-WXX.md` - Past weeks

Status indicators:
- 🟢 Available
- 🔴 Busy
- 🟡 Tentative

## Motivation
I built this because managing multiple calendars across work, personal, family, startup, etc.. is way harder than it needs to be. I needed to enable public sharing of my availability without exposing sensitive calendar data, after playing with it for a bit a public github repo felt natural.

## 🗺️ Roadmap
- add custom template support
- add a month view (if there's interest)
- publish .ics to the public repo
- implement some way to book time slots (need ideas; mailto?)
- static site output option (htmx)

## 📦 Versioning

This project follows [Semantic Versioning](https://semver.org/). See [CHANGELOG.md](CHANGELOG.md) for version history.
