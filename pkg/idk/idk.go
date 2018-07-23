package idk

import (
	"log"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/paypal/gatt"
)

// Client has the methods used by the
type Client interface {
	NewCustomExecutor(ce *CustomExecutor)
	NewPlayerExecutor(pe *PlayerExecutor)
	StopExecutor() error
	NotifyStep() <-chan Event
	NotifyDone() <-chan *Result
}

const (
	continuePacket      = 0x00
	endPacket           = 0x01
	errorDecodingPacket = 0x02
	resultPacketSize    = 0x13
	packetInterval      = 50
)

var (
	// ServiceUUID is the UUID of the idk service
	ServiceUUID = gatt.UUID16(0xAAAA)
	// PlayerUUID is the UUID of the create new player executor characteristic
	PlayerUUID = gatt.UUID16(0xBBBB)
	// CustomUUID is the UUID of the create new custom executor characteristic
	CustomUUID = gatt.UUID16(0xCCCC)
	// NotifyStepUUID is the UUID of the characteristic that notifies when a step is done
	NotifyStepUUID = gatt.UUID16(0xDDDD)
	// NotifyDoneUUID is the UUID of the characteristic that notifies when the routine is done
	NotifyDoneUUID = gatt.UUID16(0xEEEE)
	// StopExecutorUUID is the UUID of the characteristic that stops the current executor.
	StopExecutorUUID = gatt.UUID16(0xFFFF)
)

// NewService returns a new gatt service with the characteristics
// of the IDK protocol.
func NewService(client Client) *gatt.Service {
	bytes := []byte{}
	s := gatt.NewService(gatt.MustParseUUID(ServiceUUID.String()))
	s.AddCharacteristic(gatt.MustParseUUID(PlayerUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			p := &PlayerExecutor{}
			log.Printf("len(data)=%v", len(data))
			if err := proto.Unmarshal(data, p); err != nil {
				// invalid bytes
				return gatt.StatusUnexpectedError
			}
			client.NewPlayerExecutor(p)
			return gatt.StatusSuccess
		})
	s.AddCharacteristic(gatt.MustParseUUID(CustomUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			if data[0] == continuePacket {
				bytes = append(bytes, data[1:]...)
				return gatt.StatusSuccess
			}
			bytes = append(bytes, data[1:]...)
			c := &CustomExecutor{}
			if err := proto.Unmarshal(bytes, c); err != nil {
				bytes = []byte{}
				return gatt.StatusUnexpectedError
			}
			client.NewCustomExecutor(c)
			bytes = []byte{}
			return gatt.StatusSuccess
		})
	s.AddCharacteristic(gatt.MustParseUUID(StopExecutorUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			if err := client.StopExecutor(); err != nil {
				log.Printf("failed to stop executor: %s", err)
			}
			return gatt.StatusUnexpectedError
		})
	s.AddCharacteristic(gatt.MustParseUUID(NotifyStepUUID.String())).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			for e := range client.NotifyStep() {
				b, err := proto.Marshal(&e)
				if err != nil {
					n.Write([]byte{errorDecodingPacket})
					return
				}
				n.Write(b)
			}
		})
	s.AddCharacteristic(gatt.MustParseUUID(NotifyDoneUUID.String())).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			results, err := proto.Marshal(<-client.NotifyDone())
			if err != nil {
				results = []byte{}
				n.Write([]byte{errorDecodingPacket})
				return
			}
			for i := 0; i <= len(results)/resultPacketSize; i++ {
				b := []byte{continuePacket}
				packetSize := resultPacketSize
				if i == len(results)/resultPacketSize {
					b[0] = endPacket
					packetSize = len(results) - resultPacketSize*i
				}
				b = append(b, results[i*resultPacketSize:i*resultPacketSize+packetSize]...)
				n.Write(b)
				time.Sleep(packetInterval * time.Millisecond)
			}
		})
	return s
}
