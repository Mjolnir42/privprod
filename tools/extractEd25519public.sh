#!/bin/sh

PUBKEY="${1:?}"

cat "${PUBKEY}" | cut -d' ' -f2 | b64decode -r | hexdump -s 19 -n 32 -e '1/1 "%.2x"' | tr "[:upper:]" "[:lower:]"
