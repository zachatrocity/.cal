# .cal - Self Hosted Scheduler
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

This is in progress, I'll publish to a docker image when alpha is ready, you can see example output here:

https://github.com/zachatrocity/cal/


## TODO

- [ ] Create tests and coverage 
- [ ] Create month overview template
- [ ] Publish anon .ics
- [ ] Add booking guidlines
- [ ] Convert markdown.go to use a templating engine
- [ ] Publish docker image

## Goals

- aggregate multiple ICS calendar feeds (URLs)
- remove event titles so we can make public
- leverage github repo as public facing UI
- sync w/ cronjob

## Setup
this is designed to be self hosted and depends on a preconfigured ssh key for authenticiation.

1. Create a new git repo to be used as your public facing calendar (i.e. `zachatrocity/cal`)

2. Copy the `docker-compose.yml` and update the envs

3. `docker-compose up -d`

## Schedule Format

The schedule is published to your GitHub repository with the following structure:

- `README.md` - current week's schedule
- `future/YYYY-WXX.md` - upcoming weekly schedules
- `past/YYYY-WXX.md` - past weekly schedule archives

The schedule shows:
- 🟢 Available - Open slots
- 🔴 Busy - Scheduled meetings (aggregated & anonymized)
- 🟡 Tentative - Tentatively scheduled

## Development

To build and run locally:

```bash
cd backend
go mod download
go build ./cmd/server
./server
```
