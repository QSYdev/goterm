package idk

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/paypal/gatt"
)

// Client has the methods used by the
type Client interface {
	NewCustomExecutor(ce *CustomExecutor)
	NewPlayerExecutor(pe *PlayerExecutor)
	// TODO: func to handle the step event
	NotifyStep()
	// TODO: func to handle the done event
	NotifyDone()
}

var (
	serviceUUID    = gatt.UUID16(0xAAAA)
	playerUUID     = gatt.UUID16(0xBBBB)
	customUUID     = gatt.UUID16(0xCCCC)
	notifyStepUUID = gatt.UUID16(0xDDDD)
	notifyDoneUUID = gatt.UUID16(0xEEEE)
)

// NewService returns a new gatt service with the characteristics
// of the IDK protocol.
func NewService(client Client) *gatt.Service {
	s := gatt.NewService(gatt.MustParseUUID(serviceUUID.String()))
	s.AddCharacteristic(gatt.MustParseUUID(playerUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			p := &PlayerExecutor{}
			if err := proto.Unmarshal(data, p); err != nil {
				// invalid bytes
				return gatt.StatusUnexpectedError
			}
			client.NewPlayerExecutor(p)
			return gatt.StatusSuccess
		})
	s.AddCharacteristic(gatt.MustParseUUID(customUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			c := &CustomExecutor{}
			if err := proto.Unmarshal(data, c); err != nil {
				// invalid bytes
				return gatt.StatusUnexpectedError
			}
			client.NewCustomExecutor(c)
			return gatt.StatusSuccess
		})
	return nil
}
