package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/exp/inotify"
)

func main() {

	configFile := flag.String("c", "config.toml", "Config file path")
	flag.Parse()

	findSocatBinary()

	services := parseConfig(*configFile)

	for _, s := range services {
		if !s.Enabled {
			log.Println(fmt.Sprintf("Service %s is disabled", s.Name))
		} else {
			s.Start()
		}
	}

	var watch *inotify.Watcher
	var err error
	if watch, err = inotify.NewWatcher(); err != nil {
		log.Println(fmt.Sprintf("Could not create a watcher: %s. Changes will need a restart to become effective", err))
	} else {
		setWatch(watch, filepath.Dir(*configFile))
	}

	for {
		select {
		case ev := <-watch.Event:
			if filepath.Base(ev.Name) == *configFile {
				log.Println("Change on the config file detected. Reloading services...")

				unsetWatch(watch, filepath.Dir(*configFile))

				for _, s := range services {
					s.Stop()
				}

				<-time.NewTimer(time.Millisecond * 500).C
				services = parseConfig(*configFile)

				for _, s := range services {
					if !s.Enabled {
						log.Println(fmt.Sprintf("Service %s is disabled", s.Name))
					} else {
						s.Start()
					}
				}
				setWatch(watch, filepath.Dir(*configFile))
			}
		}
	}
}

func parseConfig(configFile string) map[string]*Service {
	var config Services
	var md toml.MetaData
	var err error
	if md, err = toml.DecodeFile(configFile, &config); err != nil {
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
	return services
}

func setWatch(watch *inotify.Watcher, path string) {
	if err := watch.AddWatch(path, inotify.IN_MODIFY); err != nil {
		log.Println(fmt.Sprintf("Error %s setting a watcher on the config file. Changes will need a restart to become effective", err))
	}
}

func unsetWatch(watch *inotify.Watcher, path string) {
	watch.RemoveWatch(path)
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
