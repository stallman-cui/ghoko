// Copyright 2013 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/golib/signal"
	"net"
	"os"
	"syscall"
	"time"
)

const (
	GITLAB = "gitlab"
	GITHUB = "github"
)

var (
	addr       = flag.String("addr", ":8080", "Address of http service")
	scriptPath = flag.String("script", "./", "Path of lua files")
	secret     = flag.String("secret", "", "Secret token")
	defHosting = flag.String("defualt", GITLAB, "Default code hosting site")
	cert       = flag.String("tls-cert", "", "TLS cert file")
	key        = flag.String("tls-key", "", "TLS key file")
)

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
	log.Flag()
}

func main() {
	log.Messagef("Starting: addr=%q script=%q default=%q",
		*addr, *scriptPath, *defHosting)
	defer func() {
		log.Message("Exited!")
		time.Sleep(time.Millisecond * 100)
	}()
	p := *scriptPath
	if p[len(p)-1] == 47 {
		p = p[:len(p)-1]
	}
	hook := NewHook(*addr, p, *secret, *defHosting)
	if *cert != "" && *key != "" {
		if err := hook.SetTLS(*cert, *key); err != nil {
			log.Error(err)
			return
		}
	}
	go func() {
		if e := hook.Serve(); e != nil {
			if _, ok := e.(*net.OpError); !ok {
				log.Error(e)
			}
		}
	}()
	defer hook.Close()
	sh := signal.NewHandler()
	sh.Bind(os.Interrupt, func() bool { return true })
	sh.Bind(syscall.SIGUSR1, func() bool {
		if e := hook.Close(); e != nil {
			log.Error(e)
		}
		var attr os.ProcAttr
		attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
		attr.Sys = &syscall.SysProcAttr{}
		_, e := os.StartProcess(os.Args[0], os.Args, &attr)
		if e != nil {
			log.Error(e)
		}
		return true
	})
	sh.Loop()
}