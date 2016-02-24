package main

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/nsf/termbox-go"
)

func main() {

	c, err := net.Dial("tcp", "localhost:1111")
	if err != nil {
		log.Fatalf("failed to connect server: %s", err)
	}

	go func() { io.Copy(os.Stdout, c) }()
	if err := termbox.Init(); err != nil {
		log.Fatalf("failed to initialize terminal: %s", err)
	}

	defer termbox.Close()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			default:
				var b byte
				if ev.Ch != 0 {
					if ev.Ch == '~' {
						return
					}
					// send character
					b = byte(ev.Ch)
				} else {
					// send ctrl sequence
					b = byte(ev.Key)
				}
				c.Write([]byte{b})
			}
		}
	}
}
