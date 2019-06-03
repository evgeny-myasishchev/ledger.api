#!/bin/sh

if [ $# -eq 0 ]; then
  exec ${SERVICE_NAME}
else
  exec "$@"
fi
