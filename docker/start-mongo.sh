#!/usr/bin/env bash
set -euo pipefail

# This key is local to the single member and is regenerated when it starts.
keyfile=/tmp/aep-rs.key
head -c 512 /dev/urandom | base64 > "$keyfile"
chown mongodb:mongodb "$keyfile"
chmod 400 "$keyfile"

exec /usr/local/bin/docker-entrypoint.sh mongod --replSet rs0 --keyFile "$keyfile" --bind_ip_all
