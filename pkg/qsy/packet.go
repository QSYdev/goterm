package qsy

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

	Red     = Color(61440)
	Green   = Color(3840)
	Blue    = Color(240)
	Cyan    = Color(4080)
	Magenta = Color(61680)
	Yellow  = Color(65280)
	White   = Color(65520)
	NoColor = Color(0)
)

// Color is a RGB color encoded with 16 bits.
type Color uint16

// String implements the Stringer interface.
func (c Color) String() string {
	switch c {
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	case Cyan:
		return "Cyan"
	case Magenta:
		return "Magenta"
	case Yellow:
		return "Yellow"
	case White:
		return "White"
	case NoColor:
		return "NoColor"
	default:
		return "Unknown"
	}
}

// newConfig returns a config with the specified
// configurations set.
func newConfig(sound, distance bool) uint16 {
	c := uint16(0)
	if sound {
		c = c | 1<<1
	}
	if distance {
		c = c | 1<<0
	}
	return c
}

// Packet represents an incoming or outgoing QSY Packet.
type Packet struct {
	Signature [3]byte
	T         uint8
	ID        uint16
	Color     Color
	Delay     uint32
	Step      uint16
	Config    uint16
}

// NewPacket returns a new packet given the specified parameters.
func NewPacket(t uint8, id uint16, color Color, delay uint32, step uint16, sound, distance bool) Packet {
	return Packet{
		Signature: [3]byte{'Q', 'S', 'Y'},
		T:         t,
		ID:        id,
		Color:     color,
		Delay:     delay,
		Step:      step,
		Config:    newConfig(sound, distance),
	}
}

// Encode returns the encoded bytes of the given packet.
func (pkt Packet) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, pkt); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (pkt Packet) String() string {
	return fmt.Sprintf("Type: %v - ID: %v - Color: %s", pkt.T, pkt.ID, pkt.Color)
}

// Decode decodes the bytes into the packet struct.
func Decode(b []byte, pkt *Packet) error {
	buf := bytes.NewBuffer(b)
	return binary.Read(buf, binary.BigEndian, pkt)
}
