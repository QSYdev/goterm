package qsy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
	"golang.org/x/sync/syncmap"
)

const (
	// QSYPort is the port used to communicate with nodes.
	QSYPort = 3000
	// DefaultDelay is the default amount of seconds to
	// wait for keep alive cleanups.
	DefaultDelay = 5
	defaultRoute = "0.0.0.0"
	tcpv         = "tcp4"
	udpv         = "udp4"
)

var (
	// ErrNotExist is an error when a given node is
	// not in the pool.
	ErrNotExist = errors.New("node does not exist")
)

type newConn struct {
	id   uint16
	addr string
}

// Listener has a Receive method called when a new packet
// comes in and a Lost method called when a node gets
// disconnected.
type Listener interface {
	Receive(Packet)
	LostNode(nodeID uint16)
	NewNode(id uint16)
}

// Server handles all-things QSY.
// It is in charge of:
// 	* Managing the pool of connected nodes
//	* Forward incoming packets to whomever is listening
//	* Send packets to connected nodes
//	* Searching for nodes
// The zero-value configuration for server IS NOT valid.
type Server struct {
	pool syncmap.Map

	ctx context.Context

	pconn     *ipv4.PacketConn
	laddr     *net.TCPAddr
	route     string
	listeners []Listener

	incoming     chan newConn
	packets      chan Packet
	lost         chan uint16
	connected    chan uint16
	disconnected chan uint16

	delay int64

	run bool

	mu        sync.RWMutex
	searching bool
}

// NewServer returns a new QSY server.
// The parameters for configuring the server are:
//      * ctx: context used for cancellation.
//	* inf: specifies the network interface where
//	  the addresses live.
//	* group: the IP for the multicast group used for listening
// 	  hello packets over udp.
//	* route: an IPv4 address for configuring the UDP server.
//	  If no route is provided then the default route will be used.
//	* localAddress: the tcp address associated with the network
//	  interface.
func NewServer(ctx context.Context, logger io.Writer, inf string, group net.IP, route, localAddress string, listeners ...Listener) (*Server, error) {
	if inf == "" || group == nil || localAddress == "" {
		return nil, errors.New("please provide the network interface, multicast group and local tcp address")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if route == "" {
		route = defaultRoute
	}
	i, err := net.InterfaceByName(inf)
	if err != nil {
		return nil, errors.Wrap(err, "invalid network interface")
	}
	laddr, err := net.ResolveTCPAddr(tcpv, localAddress+":0")
	if err != nil {
		return nil, errors.Wrap(err, "invalid local address")
	}
	c, err := net.ListenPacket(udpv, net.JoinHostPort(route, fmt.Sprintf("%v", QSYPort)))
	if err != nil {
		return nil, errors.Wrap(err, "invalid route config")
	}
	p := ipv4.NewPacketConn(c)
	if err = p.JoinGroup(i, &net.UDPAddr{IP: group}); err != nil {
		return nil, errors.Wrap(err, "failed to join group")
	}
	if logger == nil {
		logger = ioutil.Discard
	}
	log.SetOutput(logger)
	srv := &Server{
		ctx:       ctx,
		pconn:     p,
		route:     route,
		laddr:     laddr,
		delay:     DefaultDelay,
		listeners: listeners,
	}
	return srv, nil
}

// Send sends the given packet, the node to sent to is
// specified within the packet. It returns an error if
// does not exist or packet is incorrect. The writing
// to the connection may or may not succeed, either way
// you will find that if you implement Listener interface.
func (srv *Server) Send(packet Packet) error {
	b, err := packet.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode packet")
	}
	n, ok := srv.pool.Load(packet.ID)
	if !ok {
		return errors.Wrapf(ErrNotExist, "id %v", packet.ID)
	}
	node := n.(*node)
	node.Send(b)
	return nil
}

// Nodes returns an array with all the ids of the currently
// connected nodes.
func (srv *Server) Nodes() []uint16 {
	ids := []uint16{}
	srv.pool.Range(func(key interface{}, value interface{}) bool {
		ids = append(ids, key.(uint16))
		return true
	})
	return ids
}

// ListenAndAccept runs the qsy server, listening over udp for incoming
// connection requests, and establishing new connections over tcp.
func (srv *Server) ListenAndAccept() error {
	if srv.run {
		return errors.New("server is already running")
	}
	if err := srv.ctx.Err(); err != nil {
		return errors.Wrap(err, "server was stopped")
	}
	srv.run = true
	srv.packets = make(chan Packet, 50)
	srv.lost = make(chan uint16, 50)
	srv.connected = make(chan uint16, 50)
	srv.disconnected = make(chan uint16, 50)
	srv.incoming = make(chan newConn, 50)
	srv.mu.Lock()
	srv.searching = true
	srv.mu.Unlock()
	go srv.listen()
	go srv.accept()
	if srv.listeners != nil {
		go srv.forward()
	}
	return nil
}

// forward forwards events to the listeners. Events so far are:
// * Incoming packet that it's not keep alive
// * Disconnected node
// * New node connection
func (srv *Server) forward() {
	for {
		select {
		case p := <-srv.packets:
			for _, v := range srv.listeners {
				go v.Receive(p)
			}
		case id := <-srv.disconnected:
			for _, v := range srv.listeners {
				go v.LostNode(id)
			}
		case id := <-srv.connected:
			for _, v := range srv.listeners {
				go v.NewNode(id)
			}
		case <-srv.ctx.Done():
			return
		}
	}
}

// accept listens on incoming connections and handles lost connections.
func (srv *Server) accept() {
	for {
		select {
		case nconn := <-srv.incoming:
			if _, ok := srv.pool.Load(nconn.id); ok {
				srv.lost <- nconn.id
				break
			}
			tconn, err := srv.dial(nconn)
			if err != nil {
				log.Printf("failed to dial new conn: %s", err)
				break
			}
			if err := tconn.SetNoDelay(true); err != nil {
				log.Printf("failed to set no delay on node conn: %s", err)
				srv.lost <- nconn.id
				break
			}
			n := newNode(tconn, nconn.id, nconn.addr)
			srv.pool.Store(n.id, n)
			srv.connected <- n.id
			n.Listen(srv.packets, srv.lost, srv.delay)
		case nid := <-srv.lost:
			n, ok := srv.pool.Load(nid)
			if !ok {
				break
			}
			node := n.(*node)
			if err := node.Close(); err != nil {
				log.Printf("failed to close node %v: %s", nid, err)
			}
			srv.pool.Delete(nid)
			srv.disconnected <- nid
		case <-srv.ctx.Done():
			srv.pool.Range(func(id interface{}, n interface{}) bool {
				node := n.(*node)
				if err := node.Close(); err != nil {
					log.Printf("failed to close node %v: %s", id, err)
				}
				return true
			})
			close(srv.lost)
			close(srv.packets)
			close(srv.disconnected)
			close(srv.connected)
			return
		}
	}
}

func (srv *Server) dial(n newConn) (*net.TCPConn, error) {
	taddr, err := net.ResolveTCPAddr(tcpv, n.addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve node addres")
	}
	taddr.Port = QSYPort
	tconn, err := net.DialTCP(tcpv, srv.laddr, taddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initiate connection to node")
	}
	return tconn, nil
}

// Search allows to accept incoming connection requests. If the server
// was stopped this is a NOP.
func (srv *Server) Search() {
	if err := srv.ctx.Err(); err != nil {
		return
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.searching = true
}

// StopSearch stops accepting incoming connection requests. If the
// server was stopped this is a NOP.
func (srv *Server) StopSearch() {
	if err := srv.ctx.Err(); err != nil {
		return
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.searching = false
}

// listen listens for new udp hello packets and forwards
// them through the incoming channel. Note that listen does
// not decide what to do with the incoming connection requests.
// listen runs in its own go routine.
func (srv *Server) listen() {
	for {
		srv.mu.RLock()
		if !srv.searching {
			srv.mu.RUnlock()
			continue
		}
		srv.mu.RUnlock()

		b := make([]byte, 16)
		_, _, src, err := srv.pconn.ReadFrom(b)
		if err != nil {
			log.Printf("failed to read from udp conn: %s", err)
			break
		}
		if b[QHeader] != 'Q' || b[SHeader] != 'S' || b[YHeader] != 'Y' {
			continue
		}
		pkt := Packet{}
		if err := Decode(b, &pkt); err != nil {
			log.Printf("failed to decode packet: %v\nPacket bytes: %v", err, b)
			continue
		}
		if pkt.T != HelloT {
			continue
		}
		srv.incoming <- newConn{id: pkt.ID, addr: src.String()}
	}
	close(srv.incoming)
	if err := srv.pconn.Close(); err != nil {
		log.Printf("failed to close udp conn: %s", err)
	}
}
