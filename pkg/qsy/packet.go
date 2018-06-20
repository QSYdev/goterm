package qsy

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

const (
	// PacketSize represents the size of a QSY Packet
	PacketSize = 16

	QHeader       = 0x00
	SHeader       = 0x01
	YHeader       = 0x02
	TypeHeader    = 0x03
	IDHeader      = 0x04
	ColorRGHeader = 0x06
	ColorBHeader  = 0x07
	DelayHeader   = 0x08
	StepHeader    = 0x0C
	ConfigHeader  = 0x0E

	HelloT     = 0x00
	CommandT   = 0x01
	ToucheT    = 0x02
	KeepAliveT = 0x03
)

// Packet represents an incoming or outgoing QSY Packet.
type Packet struct {
	Signature [3]byte
	T         uint8
	ID        uint16
	Color     uint16
	Delay     uint32
	Step      uint16
	Config    uint16
}

// NewPacket returns a new packet given the specified parameters.
// TODO: color and config not yet implemented
func NewPacket(t uint8, id uint16, color string, delay uint32, step uint16, sound bool, touch bool) Packet {
	return Packet{
		Signature: [3]byte{'Q', 'S', 'Y'},
		T:         t,
		ID:        id,
		Color:     0, // TODO
		Delay:     delay,
		Step:      step,
		Config:    0, // TODO
	}
}

// Encode returns the encoded bytes of the given packet.
func (pkt Packet) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, pkt); err != nil {
		return nil, errors.Wrap(err, "failed to encode packet")
	}
	return buf.Bytes(), nil
}

// Decode decodes the bytes into the packet struct.
func Decode(b []byte, pkt *Packet) error {
	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.BigEndian, pkt); err != nil {
		return errors.Wrap(err, "failed to decode packet")
	}
	return nil
}

func (pkt Packet) String() string {
	return fmt.Sprintf("Type: %v\nID: %v\n", pkt.T, pkt.ID)
}

func ntoh(s uint16) uint16 {
	return (((0xFF) & s) << 8) | (((0xFF00) & s) >> 8)
}
