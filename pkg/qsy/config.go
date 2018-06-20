package qsy

import (
	"net"

	"golang.org/x/net/ipv4"
)

// Config is the configuration used for starting a
// qsy server.
type Config struct {
	Group     net.IP
	Inf       *net.Interface
	LocalAddr *net.TCPAddr
	PConn     *ipv4.PacketConn
}

// Just for doc
// laddr, err := net.ResolveTCPAddr(tcpv, localAddress+":0")
// if err != nil {
// return nil, errors.Wrap(err, "invalid local address")
// }
// c, err := net.ListenPacket(udpv, net.JoinHostPort(route, fmt.Sprintf("%v", QSYPort)))
// if err != nil {
// return nil, errors.Wrap(err, "invalid route config")
// }
// p := ipv4.NewPacketConn(c)
