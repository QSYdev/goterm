package qsy

import "testing"

func TestEncode(t *testing.T) {
	var (
		pkt         = NewPacket(ToucheT, uint16(18), Red, uint32(0), uint16(0), false, true)
		validPacket = touchePacket()
	)
	b, err := pkt.Encode()
	if err != nil {
		t.Fatalf("failed to encode packet: %s", err)
	}
	for k, v := range b {
		if v != validPacket[k] {
			t.Fatalf("byte in position %v is wrong. Got %v expected %v", k, v, validPacket[k])
		}
	}
}

func TestDecode(t *testing.T) {
	var (
		pkt            = Packet{}
		expectedPacket = NewPacket(HelloT, uint16(18), Cyan, uint32(0), uint16(0), true, true)
		encodedPacket  = helloPacket()
	)
	if err := Decode(encodedPacket, &pkt); err != nil {
		t.Fatalf("failed to decode packet: %s", err)
	}
	if pkt != expectedPacket {
		t.Fatalf("wrong decoding.\n\tExpected: %v\n\tGot: %v", expectedPacket, pkt)
	}
}

func helloPacket() []byte {
	p := make([]byte, PacketSize)
	p[QHeader] = 'Q'
	p[SHeader] = 'S'
	p[YHeader] = 'Y'
	p[TypeHeader] = HelloT
	p[IDHeader] = uint8(0)
	p[IDHeader+0x01] = uint8(18)
	// color cyan
	p[ColorRGHeader] = uint8(15)
	p[ColorBHeader] = uint8(240)
	p[DelayHeader] = uint8(0)
	p[DelayHeader+0x01] = uint8(0)
	p[DelayHeader+0x02] = uint8(0)
	p[DelayHeader+0x03] = uint8(0)
	p[StepHeader] = uint8(0)
	p[StepHeader+0x01] = uint8(0)
	p[ConfigHeader] = uint8(0)
	p[ConfigHeader+0x01] = uint8(3)
	return p
}

func touchePacket() []byte {
	p := make([]byte, PacketSize)
	p[QHeader] = 'Q'
	p[SHeader] = 'S'
	p[YHeader] = 'Y'
	p[TypeHeader] = ToucheT
	p[IDHeader] = uint8(0)
	p[IDHeader+0x01] = uint8(18)
	// color red
	p[ColorRGHeader] = uint8(240)
	p[ColorBHeader] = uint8(0)
	p[DelayHeader] = uint8(0)
	p[DelayHeader+0x01] = uint8(0)
	p[DelayHeader+0x02] = uint8(0)
	p[DelayHeader+0x03] = uint8(0)
	p[StepHeader] = uint8(0)
	p[StepHeader+0x01] = uint8(0)
	p[ConfigHeader] = uint8(0)
	p[ConfigHeader+0x01] = uint8(1)
	return p
}
