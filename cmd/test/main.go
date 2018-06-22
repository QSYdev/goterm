package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/matipan/terminal/pkg/qsy"
)

type r struct {
	t time.Time
}

func (r r) Receive(p qsy.Packet) {
	if p.T == qsy.KeepAliveT {
		log.Printf("Got packet at %v", time.Since(r.t).Nanoseconds()/1000)
	}
}

func (r r) LostNode(id uint16) {
	log.Printf("lost node at %v: %v", time.Since(r.t).Nanoseconds()/1000, id)
}

func (r r) NewNode(id uint16) {
	log.Printf("new node: %v", id)
}

func main() {
	// TODO: make flags for the multicast address and all that.
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(120*time.Second))
	t := time.Now()
	s, err := qsy.NewServer(ctx, t, os.Stdout, "wlan0", net.IP{224, 0, 0, 12}, "", "10.0.0.1", r{t})
	if err != nil {
		log.Printf("failed to create server: %s", err)
		os.Exit(1)
	}
	if err := s.ListenAndAccept(); err != nil {
		log.Printf("failed to start server: %s", err)
		os.Exit(1)
	}
	<-ctx.Done()
}
