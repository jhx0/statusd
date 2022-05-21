package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const prgName = "statusd"
const prgVersion = "0.2"

const logfile = "/var/log/statusd.log"
const commandsFile = "/etc/statusd/commands"

const helpMsg = `
USAGE: statusd [ -v | -h ]

	-v		show version information
	-h		show this help message

Copyright (2018-2022) jhx (jhx0x00@gmail.com)
If there are any suggestions or in general feedback, send
me a mail to the given address above. enjoy!
`

type address struct {
	address string
	port    string
}

type Statusd struct {
	commands []string
}

const (
	OK = iota
	ERR
)

func (a address) getFullAddress() string { 
	return (a.address + ":" + a.port) 
}

func (a address) getPort() string { 
	return a.port 
}

func (a address) getAddress() string {
	return a.address 
}

func (s Statusd) server(c address) {
	serverListener, err := net.Listen("tcp", c.getFullAddress())
	if err != nil {
		s.logEntry("Listen", err.Error(), ERR)
		os.Exit(1)
	}

	s.logEntry("server", fmt.Sprintf("Server is listening on %s:%s", c.getAddress(), c.getPort()), OK)

	for {
		client, err := serverListener.Accept()
		if err != nil {
			continue
		}

		s.logEntry("server", fmt.Sprintf("Client connected from %s", client.RemoteAddr()), OK)

		go s.sendStatus(client)
	}

}

func (s *Statusd) sendStatus(client net.Conn) {
	var statusLine string

	defer client.Close()

	statusLine += "Current time: " + time.Now().String() + "\n\n"

	s.logEntry("sendStatus", "Executing commands on the server", OK)

	for i := range s.commands {
		statusLine += "### " + s.commands[i] + " ###" + "\n"
		statusLine += s.getCommandOutput(s.commands[i])
		statusLine += "\n\n"
	}

	s.logEntry("sendStatus", "Sending information to the client", OK)

	_, err := io.WriteString(client, statusLine)
	if err != nil {
		return
	}

}

func (s Statusd) getCommandOutput(command string) string {
	cmd := strings.Split(command, " ")

	output, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		s.logEntry("getCommandOutput", err.Error(), ERR)
		os.Exit(1)
	}

	return strings.TrimSpace(string(output))
}

func (s *Statusd) parseCommands() {
	f, err := os.Open(commandsFile)
	if err != nil {
		s.logEntry("parseCommands", err.Error(), ERR)
		os.Exit(1)
	}

	reader := bufio.NewReader(f)

	defer f.Close()

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		if len(line) < 1 {
			continue
		}

		s.commands = append(s.commands, string(line))
	}
}

func (s Statusd) logEntry(function string, message string, loglevel int) {
	if loglevel == OK {
		log.Printf("[INFO] (%s) - %s\n", function, message)
	} else {
		log.Printf("[ERROR] (%s) - %s\n", function, message)
	}
}

func (s Statusd) createLogfile(logfile string) *os.File {
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Cannot create logfile, aborting!")
		os.Exit(1)
	}

	log.SetOutput(f)

	return f
}

func (s Statusd) hasCommandsFile() {
	_, err := os.Stat(commandsFile) 
	if err != nil {
		s.logEntry("hasCommandsFile", "No commands file found, exiting!", ERR)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println(helpMsg)
	os.Exit(0)
}

func showVersion() {
	fmt.Printf("%s v%s\n", prgName, prgVersion)
	os.Exit(0)
}

func main() {
	if os.Getuid() != 0 {
		fmt.Println("statusd needs to be run with root rights, aborting!")
		os.Exit(1)
	}

	ip := flag.String("i", "localhost", "IP Address of the machine")
	port := flag.String("p", "7000", "Port number")

	for _, arg := range os.Args[1:] {
		if strings.Compare(arg, "-v") == 0 {
			showVersion()
		}

		if strings.Compare(arg, "-h") == 0 {
			showHelp()
		}
	}

	flag.Parse()

	s := Statusd{}

	logfile := s.createLogfile(logfile)

	defer logfile.Close()

	s.hasCommandsFile()
	s.parseCommands()

	s.server(address{*ip, *port})

	return
}
