#!/bin/sh
redis-server &
# Generate a new encryption key every time the container starts
/app/gohole -gkey
# Run GoHole
/app/gohole -s -c /root/gohole_config.json
