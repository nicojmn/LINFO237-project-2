package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/akamensky/argparse"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func scanPort(target string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		logger.Error().Err(err).Int("port", port).Msg("closed or filtered")
		return false
	}
	_ = conn.Close()
	logger.Info().Int("port", port).Msg("open")
	return true
}

func scanRange(target string, startPort, endPort int, timeout time.Duration) {
	openCount := 0
	for port := startPort; port <= endPort; port++ {
		if scanPort(target, port, timeout) {
			openCount++
		}
	}
	logger.Info().
		Str("target", target).
		Int("open_ports", openCount).
		Msg("Scan complete")
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	parser := argparse.NewParser("portscan", "TCP connect scanner attack tool")

	target := parser.String("", "target", &argparse.Options{
		Required: true,
		Help:     "Target IPv4 address",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("target cannot be empty")
			} else {
				ip := net.ParseIP(args[0])
				if ip == nil {
					return errors.New("target must be a valid IP address")
				}
				if ip.To4() == nil {
					return errors.New("target must be a valid IPv4 address")
				}
				if ip.IsLoopback() {
					return errors.New("target cannot be a loopback address")
				}

				if ip.IsUnspecified() {
					return errors.New("target cannot be an unspecified address")
				}
			}
			return nil
		},
	})

	startPort := parser.Int("", "start", &argparse.Options{
		Required: false,
		Help:     "Start of port range (1-65535)",
		Default:  1,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("starting port must be a number")
			}

			if val <= 0 || val >= 65535 {
				return errors.New("starting port must be between 1 and 65535")
			}
			return nil
		},
	})

	endPort := parser.Int("", "end", &argparse.Options{
		Required: false,
		Help:     "End of port range (start-65535)",
		Default:  65535,
		Validate: func(args []string) error {
			val, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("ending port must be a number")
			}

			if val <= 0 || val >= 65535 {
				return errors.New("ending port must be between 1 and 65535")
			}
			return nil
		},
	})

	timeoutSec := parser.Int("t", "timeout", &argparse.Options{
		Required: false,
		Help:     "Timeout per port in seconds",
		Default:  0.2,
		Validate: func(args []string) error {
			_, err := strconv.Atoi(args[0])
			return err
		},
	})

	if err := parser.Parse(os.Args); err != nil {
		logger.Fatal().
			Err(err).
			Msg("Error parsing arguments")
		return
	}

	if *endPort < *startPort {
		logger.Fatal().
			Int("start", *startPort).
			Int("end", *endPort).
			Msg("end port must be >= start port")
		return
	}

	logger.Info().
		Str("target", *target).
		Int("start_port", *startPort).
		Int("end_port", *endPort).
		Int("timeout_s", *timeoutSec).
		Msg("Beginning scan ...")

	scanRange(*target, *startPort, *endPort, time.Duration(*timeoutSec)*time.Second)
}
