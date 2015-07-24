package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"
)

var bin = flag.String("bin", "/usr/bin/socat", "Socat binary path")

const (
	STOPPED = iota
	RUNNING
)

type Services map[string]Service

type Service struct {
	Name    string
	Enabled bool
	State   int

	Srcport  int      `toml:"srcport"`
	Srcflags []string `toml:"srcflags"`
	Srcproto string   `toml:"srcproto"`

	Dsthost  string `toml:"dsthost"`
	Dstport  int    `toml:"dstport"`
	Dstproto string `toml:"dstproto"`

	cmd     *exec.Cmd
	cmdArgs []string
	restart bool
}

func NewService() (s *Service) {
	s = new(Service)

	s.cmdArgs = make([]string, 2)
	s.State = STOPPED
	s.Enabled = true
	s.restart = false

	s.Srcport = 80
	s.Srcflags = []string{"fork", "reuseaddr"}
	s.Srcproto = "TCP4-LISTEN"

	s.Dsthost = "localhost"
	s.Dstport = 8080
	s.Dstproto = "TCP4"

	return s
}

func (s *Service) Start() {
	if s.State == RUNNING {
		return
	}

	flags := ""
	for _, f := range s.Srcflags {
		flags = fmt.Sprintf("%s,%s", flags, f)
	}

	s.cmdArgs[0] = fmt.Sprintf("%s:%d%s",
		s.Srcproto,
		s.Srcport,
		flags)

	s.cmdArgs[1] = fmt.Sprintf("%s:%s:%d",
		s.Dstproto,
		s.Dsthost,
		s.Dstport)

	s.restart = true

	go s.run()
}

func (s *Service) run() {
	for s.restart {

		s.cmd = exec.Command(*bin, s.cmdArgs...)

		s.State = RUNNING
		err := s.cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(s.Name, s.cmd.Args)
		err = s.cmd.Wait()
		log.Println(s.Name, err)
		<-time.NewTimer(time.Second * 3).C
		s.State = STOPPED
	}
}

func (s *Service) Stop() {
	s.restart = false
	if s.State == RUNNING {
		s.cmd.Process.Kill()
	}
	s.State = STOPPED
}
