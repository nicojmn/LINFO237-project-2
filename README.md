# LINFO237-project-2 : network attacks and protection

Authors:

- **Da Silva Mathos, Pedro** - *NOMA*
- **Jeanmenne, Nicolas** - *48741900*

## Introduction

## Protections

## Attacks

### SSH brute-force

#### Usage

The SSH brute-force attack tool is a simple command-line script that allows you to perform a brute-force attack on an SSH server. It has no dependencies and is written in Go. Below is the usage information, arguments surrounded by brackets are optional.

If you wanna use the threaded mode, beware that if you set the number of threads too high, packets will be dropped. We recommend using a number of threads between

**TODO: find boundaries**

```bash
TODO : Paste go run ssh.go -h when done
```

### SYN flood

The script is also a CLI tool written in Go. It's send *n (or infinite)* SYN packets to a given IP address and port. Below is the usage information, arguments surrounded by  squared brackets are optional.

#### Usage

```no-highlight
usage: syn-flood [-h|--help] --src-ip "<value>" --dst-ip "<value>"
                 [-s|--src-port <integer>] [-d|--dst-port <integer>]
                 -i|--interface "<value>" [-n|--number <integer>] [-D|--debug]
                 [-t|--threaded]

                 SYN flood attack tool

Arguments:

  -h  --help       Print help information
      --src-ip     Source IP address
      --dst-ip     Destination IP address
  -s  --src-port   Source port, must be between 1025 and 65535, default is
                   random. Default: 61963
  -d  --dst-port   Destination port, must be between 1025 and 65535, default is
                   random. Default: 9582
  -i  --interface  Network interface to use for sending packets
  -n  --number     Number of packets to send, -1 for infinite. Default: 100
  -D  --debug      Enable debug mode. You should avoid using this in
                   production, or with threaded mode, as it slows down the
                   program. Default: false
  -t  --threaded   Enable packet creation and socket threading, default is
                   false. Default: false
```

## Conclusion