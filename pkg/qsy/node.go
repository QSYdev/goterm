package qsy

import (
	"context"
	"log"
	"time"
)

// Conn has the methods necessary for the node
// to operate
type Conn interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
	SetReadDeadline(t time.Time) error
}

// node represents a single node, it holds the information
// relevant to that node.
type node struct {
	Conn
	id   uint16
	addr string
}

// Listen listens over the TCPConn for incoming packets.
func (n *node) listen(ctx context.Context, packets chan<- Packet, lost chan<- uint16, kadelay int64) {
	if err := n.SetReadDeadline(time.Now().Add(time.Duration(kadelay) * time.Second)); err != nil {
		log.Printf("failed to set read deadline: %s", err)
		if err := n.Close(); err != nil {
			log.Printf("failed to close node conn: %s", err)
		}
		lost <- n.id
		return
	}
	for {
		select {
		case <-ctx.Done():
			if err := n.Close(); err != nil {
				log.Printf("failed to close node conn: %s", err)
			}
			return
		default:
			b := make([]byte, PacketSize)
			if _, err := n.Read(b); err != nil {
				lost <- n.id
				return
			}
			pkt := Packet{}
			if err := Decode(b, &pkt); err != nil {
				log.Printf("failed to decode packet, id: %v", n.id)
				break
			}
			if pkt.T == KeepAliveT {
				if err := n.SetReadDeadline(time.Now().Add(time.Duration(kadelay) * time.Second)); err != nil {
					log.Printf("failed to set read deadline: %s", err)
					if err := n.Close(); err != nil {
						log.Printf("failed to close node conn: %s", err)
					}
					lost <- n.id
					return
				}
				break
			}
			packets <- pkt
		}
	}
}
