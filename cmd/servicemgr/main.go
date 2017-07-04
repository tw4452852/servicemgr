package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/tw4452852/servicemgr/client"
)

const VER = "0.1"

func main() {
	help := flag.Bool("h", false, "help message")
	version := flag.Bool("v", false, "version")
	serverAddr := flag.String("s", ":22222", "server listen address")
	clientAddr := flag.String("c", ":22223", "client listen address")
	debugAddr := flag.String("d", ":22224", "debug listen address")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *version {
		fmt.Println(VER)
		os.Exit(0)
	}

	log.Printf("serverAddr[%q], clientAddr[%q], debugAddr[%q]\n", *serverAddr, *clientAddr, *debugAddr)

	server, err := NewServer(*serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	ln, err := net.Listen("tcp", *clientAddr)
	if err != nil {
		log.Fatal(err)
	}

	// for debug
	go func() {
		log.Println(http.ListenAndServe(*debugAddr, nil))
	}()

	// init debug log
	LogInit()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[client]: accept error: %s\n", err)
			continue
		}
		err = server.AddClient(client.NewClient(conn))
		if err != nil {
			log.Printf("[client]: AddClient error: %s\n", err)
			continue
		}
	}
}
