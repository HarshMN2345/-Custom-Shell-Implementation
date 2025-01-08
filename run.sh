#!/bin/sh

set -e # Exit early if any commands fail

# Compile the Go program
(
  cd "$(dirname "$0")" # Ensure compile steps are run within the repository directory
  go build -o /tmp/shell-target cmd/myshell/*.go
)

# Run the compiled program
exec /tmp/shell-target