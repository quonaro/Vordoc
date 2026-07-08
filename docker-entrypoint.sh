#!/bin/sh
set -eu

if [ "${VORDOC_INIT:-false}" = "true" ]; then
  echo "==> Initializing Vordoc content..."
  vordoc init
fi

exec vordoc run
