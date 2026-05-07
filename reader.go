package udplib

import "fmt"

func (c *Conn) readloop() {

	buf := make([]byte, 1024)

	for {

		n, addr, err := c.conn.ReadFrom(buf)
		{
			if err != nil && n <= 0 {
				continue
			}
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		packet := deserialize(data)

		packet.Addr = addr

		if packet.IsAck || packet.IsNack {
			c.ackCh <- packet
			continue
		}

		c.dataCh <- packet

	}
}

func (c *Conn) Read() ([]byte, error) {

	assembler := newAssembler()

	for packet := range c.dataCh {

		if !packet.Valid() {

			nack := Packet{
				IsNack: true,
				Ack:    packet.Seq, // tell sender which seq was corrupted
			}

			buf := serialize(nack)
			
			_, err := c.conn.WriteTo(buf, packet.Addr)
			{
				if err != nil {
					return nil, err
				}
			}
			continue
		}

		ack := Packet{
			IsAck: true,
			Ack:   packet.Seq + 1, // tell sender which to send next
		}

		buf := serialize(ack)

		_, err := c.conn.WriteTo(buf, packet.Addr)
		{
			if err != nil {
				return nil, err
			}
		}

		assembler.add(packet)

		if assembler.ready() {
			return assembler.assemble(), nil // full message — return to user
		}

	}

	return nil, fmt.Errorf("connection closed")
}
