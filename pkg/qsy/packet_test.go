package qsy

import "testing"

func TestEncode(t *testing.T) {
	var (
		pkt         = NewPacket(HelloT, uint16(18), "", uint32(0), uint16(0), false, false)
		validPacket = helloPacket()
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
		expectedPacket = NewPacket(HelloT, uint16(18), "", uint32(0), uint16(0), false, false)
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
	p[0x05] = uint8(18)
	p[ColorRGHeader] = uint8(0)
	p[ColorBHeader] = uint8(0)
	p[DelayHeader] = uint8(0)
	p[0x09] = uint8(0)
	p[0x0A] = uint8(0)
	p[0x0B] = uint8(0)
	p[StepHeader] = uint8(0)
	p[0x0D] = uint8(0)
	p[ConfigHeader] = uint8(0)
	return p
}
