#!/bin/sh

# Create cron job from environment variable
echo "$SYNC_SCHEDULE /usr/local/bin/dotcal" > /etc/crontabs/root

# Start crond in the background
crond -f -d 8 &

# Run initial sync
/usr/local/bin/dotcal

# Keep container running
tail -f /dev/null
