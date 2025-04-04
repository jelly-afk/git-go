#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

(cd "$SCRIPT_DIR" && go build -o mygit cmd/mygit/main.go)

"$SCRIPT_DIR/mygit" "$@"

rm "$SCRIPT_DIR/mygit" 