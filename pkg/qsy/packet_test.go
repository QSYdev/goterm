package qsy

import "testing"

func TestEncode(t *testing.T) {
	var (
		pkt         = NewPacket(HelloT, uint16(1), "", uint32(0), uint16(1), false, false)
		validPacket = []byte{}
	)
	b, err := pkt.Encode()
	if err != nil {
		t.Fatalf("failed to encode packet: %s", err)
	}
	for k, v := range b {
		t.Fatalf("byte in position %v is wrong. Got %v expected %v", k, v, validPacket[k])
	}
}

func TestDecode(t *testing.T) {
	var (
		pkt            = Packet{}
		expectedPacket = NewPacket(HelloT, uint16(1), "", uint32(0), uint16(1), false, false)
		encodedPacket  = []byte{}
	)
	if err := Decode(encodedPacket, &pkt); err != nil {
		t.Fatalf("failed to decode packet: %s", err)
	}
	if pkt != expectedPacket {
		t.Fatalf("wrong decoding.\n\tExpected: %v\n\tGot: %v", expectedPacket, pkt)
	}
}
