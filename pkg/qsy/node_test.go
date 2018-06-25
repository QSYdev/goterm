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
	read func() ([]byte, error)
}

// m.read is used to mock whatever operation we want
// to mock
func (m mockNode) Read(b []byte) (int, error) {
	var err error
	b, err = m.read()
	return len(b), err
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

func TestListenReceiving(t *testing.T) {
	t.Parallel()
	pkt := NewPacket(CommandT, uint16(1), "", uint32(0), uint16(1), false, false)
	packets := make(chan Packet)
	kadelay := int64(5)
	n := &node{
		conn: mockNode{
			read: func() ([]byte, error) {
				return pkt.Encode()
			},
		},
		id:       uint16(1),
		addr:     nodeAddr,
		requests: make(chan []byte),
	}
	go n.Listen(packets, nil, kadelay)
	p := <-packets
	if p != pkt {
		t.Fatalf("packet is not valid.\n\tExpected: %s - Got: %s\n", pkt, p)
	}
	n.Close()
	close(packets)
}

func TestListenLost(t *testing.T) {
	t.Parallel()
	lost := make(chan uint16)
	kadelay := int64(5)
	n := &node{
		conn: mockNode{
			read: func() ([]byte, error) {
				return nil, errors.New("uh oh")
			},
		},
		id:       uint16(1),
		addr:     nodeAddr,
		requests: make(chan []byte),
	}
	go n.Listen(nil, lost, kadelay)
	lid := <-lost
	if lid != uint16(1) {
		t.Fatalf("lost node id is not valid. Expected: %v - Got: %v\n", uint16(1), lid)
	}
	n.Close()
	close(lost)
}
