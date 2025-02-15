#!/bin/sh

# Configure Git
git config --global user.email "dotcal@example.com"
git config --global user.name "dotcal"

# Configure safe directory
git config --global --add safe.directory /app/repo

if [ "$DEV_MODE" != "true" ]; then
    # Create cron job from environment variable
    echo "$SYNC_SCHEDULE /usr/local/bin/dotcal" > /etc/crontabs/root

    # Start crond in the background
    crond -f -d 8 &
fi

# Run initial sync
/usr/local/bin/dotcal

# Keep container running
tail -f /dev/null
