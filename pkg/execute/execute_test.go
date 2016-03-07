package execute

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestHoge(t *testing.T) {
	opt := &ProcessOption{
		Dir:  "./log",
		Name: "test",
	}

	p, err := ExecutePTY(opt, "ping", "-c 5", "127.0.0.1")
	//p, err := ExecutePTY(opt, "ls")
	if err != nil {
		log.Fatalf("failed to setup process: %s", err)
	}
	if err := p.Start(); err != nil {
		log.Fatalf("failed to start process: %s", err)
	}

	r, err := NewProcessLogReader(opt)
	if err != nil {
		log.Fatalf("failed to create log reader: %s", err)
	}
	go r.Start()
	if err := p.Wait(); err != nil {
		log.Fatalf("failed to wait process: %s", err)
	}
	b, err := ioutil.ReadAll(r.Stdout)
	if err != nil {
		log.Fatalf("failed to read stdout:%s", err)
	}
	println(string(b))
}
