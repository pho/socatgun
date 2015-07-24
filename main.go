package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func main() {

	configFile := flag.String("c", "config.toml", "Config file path")
	flag.Parse()

	findSocatBinary()

	var config Services
	var md toml.MetaData
	var err error
	if md, err = toml.DecodeFile(*configFile, &config); err != nil {
		log.Fatal(err)
	}

	services := make(map[string]*Service)

	for name, serv := range config {
		s := NewService()
		s.Name = name

		services[name] = s

		for _, k := range md.Keys() {
			ks := strings.Split(k.String(), ".")
			if len(ks) > 1 && ks[0] == name {
				switch ks[1] {
				case "srcport":
					s.Srcport = serv.Srcport
				case "srcflags":
					s.Srcflags = serv.Srcflags
				case "srcproto":
					s.Srcproto = serv.Srcproto
				case "dsthost":
					s.Dsthost = serv.Dsthost
				case "dstproto":
					s.Dstproto = serv.Dstproto
				case "dstport":
					s.Dstport = serv.Dstport
				case "enabled":
					s.Enabled = serv.Enabled
				}
			}
		}
	}

	log.Println("Found", len(services), "services")

	for _, s := range services {
		if !s.Enabled {
			log.Println(fmt.Sprintf("Service %s is disabled", s.Name))
		} else {
			s.Start()
		}
	}

	for {
		<-time.NewTimer(time.Hour).C
	}
}

func findSocatBinary() {
	c := exec.Command("which", "socat")
	o, _ := c.Output()
	err := c.Run()
	if err != nil {
		if _, err := os.Stat(*bin); os.IsNotExist(err) {
			log.Fatalln("Couldnt find a socat binary")
		}
	} else {
		*bin = string(o)
	}
}
