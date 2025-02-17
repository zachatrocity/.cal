#!/bin/sh

# Configure Git
git config --global user.email "dotcal@example.com"
git config --global user.name "dotcal"

# Configure safe directory
git config --global --add safe.directory /app/repo

if [ "$DEV_MODE" != "true" ]; then
    # Create cron job from environment variable with output redirection
    echo "$SYNC_SCHEDULE cd /app && /usr/local/bin/dotcal >> /proc/1/fd/1 2>&1" > /etc/crontabs/root

    # Start crond in the background
    crond -f &
fi

# Run initial sync
/usr/local/bin/dotcal

# Keep container running
tail -f /dev/null
