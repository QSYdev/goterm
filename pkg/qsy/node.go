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
	id       uint16
	addr     string
	requests chan []byte
}

// newNode returns a node with the specified config.
func newNode(conn Conn, id uint16, addr string) *node {
	return &node{
		Conn: conn,
		id:   id,
		addr: addr,
	}
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
			// TODO: read is blocking so we can't really do this.
			// export this to a new method
		case b := <-n.requests:
			if _, err := n.Write(b); err != nil {
				log.Printf("failed to write to node: %s", err)
			}
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
			}
			packets <- pkt
		}
	}
}

// send sends the encoded packet to the requests channel.
// Listen will pickup that requests and write to the conn.
// This is so that we don't expose the channel.
func (n *node) send(b []byte) {
	n.requests <- b
}
