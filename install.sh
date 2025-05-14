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

PORTSCAN_DIR="src/attacks/portScan"

if [ ! -d "$PORTSCAN_DIR" ]; then
    mkdir -p "$PORTSCAN_DIR"
fi

cd "$PORTSCAN_DIR" || exit 1

if [ ! -f "$PORTSCAN_DIR/go.mod" ]; then
    echo "Initializing go module"
    go mod init group69/port-scan
    go mod tidy
else
    echo "Updating go module"
    go mod download
    go mod tidy
fi