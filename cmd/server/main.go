package main

import (
	"fmt"
	"os"

	"github.com/dustywilson/remailer"
	guerrilla "github.com/flashmob/go-guerrilla"
	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/log"
)

func main() {
	hostname, _ := os.Hostname()
	d := guerrilla.Daemon{
		Config: &guerrilla.AppConfig{
			LogFile: log.OutputStdout.String(),
			BackendConfig: backends.BackendConfig{
				"validate_process":        "Remailer",
				"save_process":            "HeadersParser|Debugger|Hasher|Header|Remailer",
				"remailer_heloname":       hostname,
				"remailer_dir":            "./config",
				"remailer_forwarder_addr": "smtp:25",
			},
			AllowedHosts: []string{"."}, // everyone and everything
			Servers: []guerrilla.ServerConfig{{
				Hostname:        "dnscow",
				ListenInterface: "0.0.0.0:5555",
				IsEnabled:       true,
			}},
		},
	}
	d.AddProcessor("Remailer", remailer.Processor)

	err := d.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	select {} // hang out!
}
