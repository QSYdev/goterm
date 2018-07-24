package ble

import (
	"time"

	"github.com/paypal/gatt"
)

// Client has the methods used by the service when
// characteristics are requested.
type Client interface {
	Write(data []byte) error
	Notify() <-chan []byte
}

const (
	continuePacket = 0x00
	endPacket      = 0x01
	payloadSize    = 0x13
	packetInterval = 100
)

var (
	serviceUUID = gatt.UUID16(0xAAAA)
	writeUUID   = gatt.UUID16(0xBBBB)
	notifyUUID  = gatt.UUID16(0xCCCC)
)

// NewService returns a new gatt service with the characteristics
// of the IDK protocol.
func NewService(client Client) *gatt.Service {
	bytes := []byte{}
	s := gatt.NewService(gatt.MustParseUUID(serviceUUID.String()))
	s.AddCharacteristic(gatt.MustParseUUID(writeUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			var s byte = gatt.StatusSuccess
			if data[0] == continuePacket {
				bytes = append(bytes, data[1:]...)
				return s
			}
			bytes = append(bytes, data[1:]...)
			if err := client.Write(bytes); err != nil {
				s = gatt.StatusUnexpectedError
			}
			bytes = []byte{}
			return s
		})
	s.AddCharacteristic(gatt.MustParseUUID(notifyUUID.String())).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			for data := range client.Notify() {
				size := len(data)
				for i := 0; i <= size/payloadSize; i++ {
					b := []byte{continuePacket}
					packetSize := payloadSize
					if i == size/payloadSize {
						b[0] = endPacket
						packetSize = size - payloadSize*i
					}
					b = append(b, data[i*payloadSize:i*payloadSize+packetSize]...)
					n.Write(b)
					time.Sleep(packetInterval * time.Millisecond)
				}

			}
		})
	return s
}
