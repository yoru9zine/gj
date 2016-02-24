package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"

	"github.com/kr/pty"
	"github.com/yoru9zine/gj"
)

func main() {
	srv := gj.NewAPIServer()
	srv.Run(":8181")
	cmd := exec.Command("bash", "-l")
	f, err := pty.Start(cmd)
	if err != nil {
		log.Fatalf("failed to start command: %s", err)
	}
	go startServer(1111, func(c net.Conn) {
		go io.Copy(c, f)
		io.Copy(f, c)
	})
	if err := cmd.Wait(); err != nil {
		log.Fatalf("failed to wait command: %s", err)
	}
}

func startServer(port int, handler func(net.Conn)) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handler(conn)
	}
}
