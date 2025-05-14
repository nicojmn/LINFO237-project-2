# LINFO2347-project-2 : network attacks and protection

Authors:

- **Da Silva Matos, Pedro** - *02092000*
- **Jeanmenne, Nicolas** - *48741900*

## Introduction

This project focuses on both attacking and protecting a entreprise network simulated through mininet. We implemented attacks
script in go , which we find more suited for network and multithreading programming. We also used the nftables firewall to protect the network from attacks.
We used the topology from the homework with some modifications. The attacks we implemented are :

- Reflected DoS
- SSH brute-force
- SYN scan
- SYN flood

### Archictecture

```no-highlight
LINFO237-project-2
├── bin
│   ├── rf-dos
│   ├── ssh-bf
│   ├── syn-flood
│   └── syn-scan
├── group69.zip
├── install.sh
├── Makefile
├── README.md
├── slides
│   ├── pics
│   ├── slides.md
│   └── slides.pdf
├── src
│   ├── attacks
│   ├── protections
│   └── topo.py
└── statement.pdf
```

## Installation

### Requirements

- [Go v1.24.2](https://go.dev/dl/)
- [Python 3](https://www.python.org/downloads/)
- [nftables](https://netfilter.org/projects/nftables/index.html)

For sucessful installation, we recommend following the [installation guide for go](https://go.dev/doc/install) to make sure you have the right version of go installed and loaded in your PATH.

### Launch a script on mininet

1. Add your public key to the mininet VM (and allow pubkey authentication if not already done)
2. Replace `SSH_HOST` with the IP/hostname and `SSH_USER` with the username of the mininet VM in the script
3. On your host machine, run `make upload` to upload the script to the mininet VM
4. On the VM, run `unzip project.zip`
5. Run `sudo -E python3 project/src/topo.py` to start the mininet topology, it will also compile the go scripts
6. Connect to a host in the topology and
7. For attacks : run `<host> ./project/bin/<script> <args>`
8. For protection : run `nft -f ./project/src/protections/<script>.nft`
9. Basic dmz firewall can be enabled by running `nft -f ./project/src/protections/dmz-firewall.nft to any dmz host.

Below is an example of how to run the SSH brute-force attack on the mininet VM. The SSH server is running on the host `ws3` and the user is `user`. The password list is `pass.txt`

```bash
# Launch the mininet topology, once in mininet you can either attack or protect
# attack
internet ./project/bin/ssh-bf --host 10.1.0.3 -u user -l ./project/src/attacks/ssh-bf/pass.txt
# protection
ws3 nft -f ./project/src/protections/ssh-bf/ssh-bf.nft
```

## Reflected DoS

### Attack

The reflected DoS tool sends a large number of UDP packets to a given IP address and port. The DNS server will respond to the target IP address. 

### Protection

The nftables script blocks every IP address that sends more than 10 packets in a minute.

## SSH brute-force

### Attack

The SSH brute-force attack CLI tool allows you to perform a brute-force attack on an SSH server. You're free to choose the host, user, port and password list. The script will try every password in the list until it finds the right one or exhausts the list. It also has a threaded mode to speed up the attack,
if you wanna use it, beware that if you set the number of threads too high, packets will be dropped. We recommend using a number of threads between 2 and 8.

### Protection

The nftables script blocks every IP address that tries to connect to the SSH server more than 3 times in a minute. You may optionally change the SSH port in `/etc/ssh/sshd_config` to make it harder for attackers / naive bots to find the SSH server, you may also allow only public key authentication to limit the attack surface.

## SYN scan

### Attack

The SYN scan tool allows you to scan a given IP address and port range. It initiates a TCP connection to the target IP address and list of ports. The script will send a SYN packet to each port and wait for a SYN_ACK or RST packet. Once done, it will print the list of open ports. 


### Protection

The nftables script blocks every IP address that initiates more than 10 connections in a minute. It also blocks every SYN packet if the number of SYN packets sent per second is greater than 10/. Loopback packets are not blocked. 

## SYN flood

### Attack

The script is also a CLI tool written in Go. It's send *n (or infinite)* SYN packets to a given IP address and port. Below is the usage information, arguments surrounded by  squared brackets are optional.


### Protection

A nftables script block every IP address that sends more than 10 SYN packets in  minute. It also blocks every SYN packet if the number of SYN packets is greater sent per second than 10/second.

We also enabled the SYN cookies option in the kernel to help mitigate SYN flood attacks. Here is the command :

```bash
sysctl -w net.ipv4.tcp_syncookies=1
```
