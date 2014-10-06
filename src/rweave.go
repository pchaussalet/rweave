package main

import (
	"flag"
	"os"

	client "./client"
	server "./server"
	"fmt"
	"strings"
)

var actions = []string{"list", "deploy", "help"}

func main() {
	flags := flag.NewFlagSet("remote", flag.ExitOnError)
	daemon := flags.Bool("d", false, "Enable daemon mode")
	host := flags.String("h", "127.0.0.1:14243", "Server address in the form ip:port (Client and Daemons mode)")
	templateFile := flags.String("t", "", "Template file location (Client mode)")
	varsFile := flags.String("v", "", "Variables file location (Client mode)")
	verbose := flags.Bool("verbose", false, "Run in verbose mode (Client mode)")
	flags.Parse(os.Args[1:])
	if *daemon {
		server.Start(*host)
	} else {
		if *verbose {
			fmt.Println("Verbose mode = ON")
		}
		command := flags.Args()
		if len(command) == 0 {
			usage(flags)
			return
		}
		switch command[0] {
		case "list":
			client.List(*host, *verbose)
		case "deploy":
			if len(command) < 3 || *templateFile == "" || *varsFile == "" {
				usage(flags)
				return
			}
			client.Deploy(command, *templateFile, *varsFile, *host, *verbose)
		case "help":
			if len(command) == 1 {
				doc("help")
				return
			}
			doc(command[1])
		}
	}
}

func usage(flags *flag.FlagSet) {
	flags.SetOutput(os.Stdout)
	fmt.Printf("USAGE:\n%v [OPTIONS] (%v) [ARGS]\n", os.Args[0], strings.Join(actions, "|"))
	flags.PrintDefaults()
}

func doc(action string) {
	var message = ""
	switch action {
	case "help":
		message = "help:\n\tPrint this message\nhelp <action>:\n\tPrint help about <action>"
	case "list":
		message = "list:\n\tPrint the list of containers running on the remote host"
	case "deploy":
		message = "deploy <component> <environment> <version>:\n\tDeploy version <version> of <component> using <environment>'s configuration"
	}
	fmt.Println(message)
}
