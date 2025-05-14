#!/usr/bin/env bash

INIT_CD=$(pwd)

SSHBF_DIR="src/attacks/ssh-bf"

if [ ! -d "$SSHBF_DIR" ]; then
    mkdir -p "$SSHBF_DIR"
fi

cd "$SSHBF_DIR" || exit 1

if [ ! -f "$SSHBF_DIR/go.mod" ]; then
    echo "Initializing go module"
    go mod init group69/ssh-bf
    go mod tidy
else
    echo "Updating go module"
    go mod download
    go mod tidy
fi

cd "$INIT_CD" || exit 1

RFDOS_DIR="src/attacks/rf-dos"

if [ ! -d "$RFDOS_DIR" ]; then
    mkdir -p "$RFDOS_DIR"
fi

cd "$RFDOS_DIR" || exit 1

if [ ! -f "$RFDOS_DIR/go.mod" ]; then
    echo "Initializing go module"
    go mod init group69/rf-dos
    go mod tidy
else
    echo "Updating go module"
    go mod download
    go mod tidy
fi

cd "$INIT_CD" || exit 1

SYNSCAN_DIR="src/attacks/syn-scan"

if [ ! -d "$SYNSCAN_DIR" ]; then
    mkdir -p "$SYNSCAN_DIR"
fi

cd "$SYNSCAN_DIR" || exit 1

if [ ! -f "$SYNSCAN_DIR/go.mod" ]; then
    echo "Initializing go module"
    go mod init group69/syn-scan
    go mod tidy
else
    echo "Updating go module"
    go mod download
    go mod tidy
fi