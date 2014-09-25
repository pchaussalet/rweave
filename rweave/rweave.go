package main

import (
	"flag"
	"os"

	client "./client"
	server "./server"
)

func main() {
	flags := flag.NewFlagSet("remote", flag.ExitOnError)
	daemon := flags.Bool("d", false, "Enable daemon mode")
	port := flags.Int("p", 14243, "Port to bind to (Daemon mode)")
	host := flags.String("h", "127.0.0.1:14243", "Server address in the form ip:port (Client mode)")
	templateFile := flags.String("t", "", "Template file location (Client mode)")
	varsFile := flags.String("v", "", "Variables file location (Client mode)")
	flags.Parse(os.Args[1:])
	if *daemon {
		server.Start(*port)
	} else {
		command := flags.Args()
		if len(command) == 0 {
			flags.PrintDefaults()
			return
		}
		switch command[0] {
		case "list":
			client.List(*host)
		case "deploy":
			if len(command) < 3 || *templateFile == "" || *varsFile == "" {
				flags.PrintDefaults()
				return
			}
			client.Deploy(command, *templateFile, *varsFile, *host)
		}
	}
}

