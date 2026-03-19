#!/usr/bin/env sh
set -e

PORT=8080

if command -v python3 >/dev/null 2>&1; then
  echo "Starting server on http://localhost:$PORT"
  python3 -m http.server $PORT
elif command -v python >/dev/null 2>&1; then
  echo "Starting server on http://localhost:$PORT"
  python -m SimpleHTTPServer $PORT
else
  echo "Error: python3 or python is required to run the local server."
  echo "Install Python from https://www.python.org/ and try again."
  exit 1
fi
