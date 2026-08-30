#!/usr/bin/env sh

set -eu

profile=${1:-coverage.out}
minimum=${2:-35.0}

if [ ! -s "$profile" ]; then
    echo "coverage profile is missing or empty: $profile" >&2
    exit 1
fi

actual=$(go tool cover -func="$profile" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')
if [ -z "$actual" ]; then
    echo "unable to determine total coverage from $profile" >&2
    exit 1
fi

if ! awk -v actual="$actual" -v minimum="$minimum" 'BEGIN { exit !(actual + 0 >= minimum + 0) }'; then
    echo "total coverage ${actual}% is below required ${minimum}%" >&2
    exit 1
fi

echo "total coverage ${actual}% meets required ${minimum}%"
