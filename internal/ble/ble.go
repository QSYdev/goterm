package ble

import (
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

// State indicates the state of the connection with
// the central
type State byte

const (
	name = "terminal"
)

var (
	// Connected represents the state when the central gets connected.
	Connected State = 0x00
	// Disconnected represents the state when the central gets disconnected.
	Disconnected State = 0x01
)

// ConnListener has a method ConnState that it's called when
// the state of the connection with the central changes.
type ConnListener interface {
	ConnState(state State)
}

// Device is the device that holds the connection to the central.
type Device struct {
	c            gatt.Central
	connListener ConnListener
}

// Init initializes the BLE device.
func Init(connListener ConnListener, services ...*gatt.Service) (*Device, error) {
	d, err := gatt.NewDevice(option.DefaultServerOptions...)
	if err != nil {
		return nil, err
	}
	dvc := &Device{connListener: connListener}
	d.Handle(
		gatt.CentralConnected(dvc.centralConnected),
		gatt.CentralDisconnected(dvc.centralDisconnected),
	)
	onStateChanged := func(d gatt.Device, s gatt.State) {
		switch s {
		case gatt.StatePoweredOn:
			uuids := []gatt.UUID{}
			for _, svc := range services {
				uuids = append(uuids, svc.UUID())
			}
			d.AdvertiseNameAndServices(name, uuids)
		}
	}
	return dvc, d.Init(onStateChanged)
}

func (d *Device) centralConnected(c gatt.Central) {
	d.connListener.ConnState(Connected)
	d.c = c
}

func (d *Device) centralDisconnected(c gatt.Central) {
	d.connListener.ConnState(Disconnected)
	d.c = c
}

// Close closes the connection to the central.
func (d *Device) Close() error {
	return d.c.Close()
}
