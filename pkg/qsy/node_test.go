package qsy

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

const (
	nodeAddr = "DON'T CARE"
)

type mockNode struct {
	read func(b []byte) (int, error)
}

// m.read is used to mock whatever operation we want
// to mock.
func (m mockNode) Read(b []byte) (int, error) {
	return m.read(b)
}

// only implement to fullfil interface
func (m mockNode) SetReadDeadline(t time.Time) error {
	return nil
}

func (m mockNode) Close() error {
	return nil
}

func (m mockNode) Write(b []byte) (int, error) {
	return 0, nil
}

func TestReadPacket(t *testing.T) {
	t.Parallel()

	var (
		i       = 0
		pkt     = Packet{}
		packets = make(chan Packet, 50)
		lost    = make(chan uint16, 50)
		kadelay = int64(5)
		node    = &node{
			conn: mockNode{
				read: func(b []byte) (int, error) {
					// return a packet only once
					if i != 0 {
						return 0, errors.New("ups")
					}
					i++
					copy(b, helloPacket())
					return len(b), nil
				},
			},
			id:       uint16(18),
			addr:     nodeAddr,
			requests: make(chan []byte),
		}
	)
	Decode(helloPacket(), &pkt)
	node.read(packets, lost, kadelay)
	p := <-packets
	if p != pkt {
		t.Fatalf("packet is not valid.\n\tExpected: %s\n\tGot: %s\n", pkt, p)
	}
	node.Close()
	close(packets)
	close(lost)
}

func TestReadLostNode(t *testing.T) {
	t.Parallel()
	var (
		lost    = make(chan uint16, 50)
		packets = make(chan Packet, 50)
		kadelay = int64(5)
		node    = &node{
			conn: mockNode{
				read: func(b []byte) (int, error) {
					return 0, errors.New("uh-oh")
				},
			},
			id:       uint16(18),
			addr:     nodeAddr,
			requests: make(chan []byte),
		}
	)
	node.read(packets, lost, kadelay)
	lid := <-lost
	if lid != uint16(18) {
		t.Fatalf("lost node id is not valid. Expected: %v - Got: %v\n", uint16(1), lid)
	}
	node.Close()
	close(lost)
	close(packets)
}
