package qsy

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
}

type r struct{}

func (r r) Receive(p Packet) {
	log.Printf("received from: %v", p.ID)
}

func (r r) LostNode(id uint16) {
	log.Printf("lost node: %v", id)
}

func (r r) NewNode(id uint16) {
	log.Printf("new node: %v", id)
}

const (
	interfaceName = "en0"
	route         = "192.168.0.6"
	duration      = 120
)

func BenchmarkServer(b *testing.B) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(duration*time.Second))
	defer cancel()
	s, err := NewServer(ctx, interfaceName, route, r{})
	if err != nil {
		b.Fatal(err)
	}
	if err := s.ListenAndAccept(); err != nil {
		b.Fatal(err)
	}
	<-ctx.Done()
}
