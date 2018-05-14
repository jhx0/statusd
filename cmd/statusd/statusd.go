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
const prgVersion = "0.1"

const logfile = "/var/log/statusd.log"
const commandsFile = "/etc/statusd/commands"

const helpMsg = `
USAGE: statusd [ -v | -h ]

	-v		show version information
	-h		show this help message

Copyright (2018) Julian "jhx" Weber (jhx0x00@gmail.com)
If there are any suggestions or in general feedback, send
me a mail to the given address above. enjoy! :^)
`

type address struct {
	address string
	port    string
}

func (a address) getFullAddress() string { return (a.address + ":" + a.port) }
func (a address) getPort() string        { return a.port }
func (a address) getAddress() string     { return a.address }

var commands []string

func checkError(function, message string, err error) {
	if err != nil {
		log.Printf("%s: %s - %s\n", function, message, err.Error())
		os.Exit(1)
	}
}

func server(c address) {
	serverListener, err := net.Listen("tcp", c.getFullAddress())
	if err != nil {
		checkError("server", "Cannot listen", err)
	}

	log.Printf("Server is listening on %s:%s\n", c.getAddress(), c.getPort())

	for {
		client, err := serverListener.Accept()
		if err != nil {
			continue
		}

		log.Printf("Client connected from %s\n", client.RemoteAddr())

		go sendStatus(client)
	}

}

func sendStatus(client net.Conn) {
	var statusLine string

	defer client.Close()

	statusLine += "Current time: " + time.Now().String() + "\n\n"

	log.Printf("Executing commands on the server\n")

	for i := range commands {
		statusLine += "### " + commands[i] + " ###" + "\n"
		statusLine += getCommandOutput(commands[i])
		statusLine += "\n\n"
	}

	log.Printf("Sending information to client\n")

	_, err := io.WriteString(client, statusLine)
	if err != nil {
		return
	}

}

func getCommandOutput(command string) string {
	cmd := strings.Split(command, " ")

	output, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		checkError("getCommandOutput", "exec.Command", err)
	}

	return strings.TrimSpace(string(output))
}

func showVersion() {
	fmt.Printf("%s v%s\n", prgName, prgVersion)
	os.Exit(0)
}

func createLogfile(logfile string) *os.File {
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		checkError("createLogfile", "os.OpenFile", err)
	}

	log.SetOutput(f)

	return f
}

func parseCommands() {
	f, err := os.Open(commandsFile)
	if err != nil {
		checkError("parseCommands", "os.Open", err)
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

		commands = append(commands, string(line))
	}
}

func showHelp() {
	fmt.Println(helpMsg)
	os.Exit(0)
}

func main() {
	if os.Getuid() != 0 {
		fmt.Printf("statusd needs to be run with root rights, aborting.")
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

	logfile := createLogfile(logfile)

	defer logfile.Close()

	parseCommands()

	server(address{*ip, *port})

	return
}
