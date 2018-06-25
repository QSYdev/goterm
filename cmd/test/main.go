package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/qsydev/goterm/pkg/qsy"
)

type r struct{}

func (r r) Receive(p qsy.Packet) {
	if p.T == qsy.KeepAliveT {
		log.Printf("keep alive node: %v", p.ID)
	}
}

func (r r) LostNode(id uint16) {
	log.Printf("lost node: %v", id)
}

func (r r) NewNode(id uint16) {
	log.Printf("new node: %v", id)
}

func main() {
	ctx := context.Background()
	s, err := qsy.NewServer(ctx, os.Stdout, "wlan0", net.IP{224, 0, 0, 12}, "", "10.0.0.1", r{})
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
