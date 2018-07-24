package ble

import (
	"sync"
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

type receiver struct {
	mu      sync.RWMutex
	lastRcv int64

	ticker *time.Ticker
	data   []byte
}

func (r *receiver) received(data []byte) {
	r.mu.Lock()
	r.lastRcv = time.Now().Unix() * 1000
	if len(r.data) == 0 {
		r.ticker = time.NewTicker(1 * time.Second)
		go r.check()
	}
	r.mu.Unlock()
	r.data = append(r.data, data...)
}

func (r *receiver) check() {
	for n := range r.ticker.C {
		r.mu.RLock()
		if (n.Unix()*1000)-r.lastRcv < 500 {
			r.mu.RUnlock()
			continue
		}
		r.mu.RUnlock()
		r.data = []byte{}
		return
	}
}

func (r *receiver) reset() {
	r.data = []byte{}
	r.ticker.Stop()
}

// NewService returns a new gatt service with the characteristics
// of the IDK protocol.
func NewService(client Client) *gatt.Service {
	rcv := &receiver{data: []byte{}}
	s := gatt.NewService(gatt.MustParseUUID(serviceUUID.String()))
	s.AddCharacteristic(gatt.MustParseUUID(writeUUID.String())).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			rcv.received(data[1:])
			var s byte = gatt.StatusSuccess
			if data[0] == continuePacket {
				return s
			}
			if err := client.Write(rcv.data); err != nil {
				s = gatt.StatusUnexpectedError
			}
			rcv.reset()
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
