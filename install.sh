#!/usr/bin/env bash

SSHBF_DIR="src/attacks/ssh-bf"

if [ ! -d "$SSHBF_DIR" ]; then
    mkdir -p "$SSHBF_DIR"
fi

cd "$SSHBF_DIR" || exit 1

 if [ ! -f "$SSHBF_DIR/go.mod" ]; then
    go mod init group69/ssh-bf
    go mod tidy
else
    go mod tidy
fi