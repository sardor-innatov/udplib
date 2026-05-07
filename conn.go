package udplib

import "net"

type Conn struct {
	conn   net.PacketConn
	addr   net.Addr
	ackCh  chan Packet
	dataCh chan Packet
	seq    uint32
}

func ListenOn(port string) (*Conn, error) {

	conn, err := net.ListenPacket("udp", port)
	{
		if err != nil {
			return nil, err
		}
	}

	c := &Conn{
		conn:   conn,
		addr:   nil,
		ackCh:  make(chan Packet, 10),
		dataCh: make(chan Packet, 10),
		seq:    0,
	}

	go c.readloop()

	return c, nil
}

func Dial(address string) (*Conn, error) {

	target, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}

	c := &Conn{
		conn:   conn,
		addr:   target,
		ackCh:  make(chan Packet, 10),
		dataCh: make(chan Packet, 10),
		seq:    0,
	}

	go c.readloop()
	return c, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
