package udplib

import (
	"net"
	"time"
)

func (c *Conn) WriteTo(data []byte, addr net.Addr) uint32 {

	var chunks [][]byte

	for i := 0; i < len(data); i += 1400 {
		if len(data)-i <= 1400 {
			chunks = append(chunks, data[i:])
			break
		}

		chunks = append(chunks, data[i:i+1400])
	}

	var packets []Packet

	for i, chunk := range chunks {

		var packet Packet
		{
			packet.Seq = uint32(i)
			packet.IsFinal = i+1 == len(chunks)
			packet.Checksum = calculateChecksum(chunk)
			packet.Data = chunk
		}

		packets = append(packets, packet)
	}

	var lostbytes uint32

	for i := 0; i < len(packets); {

		delivered := false
		for attempt := 0; attempt < 7; attempt++ {

			buf := serialize(packets[i])

			c.conn.WriteTo(buf, addr)

			select {
			case ack := <-c.ackCh:
				if ack.IsAck {
					i++
					delivered = true
				} else if ack.IsNack {
					// just continue
				}

			case <-time.After(time.Duration(100<<attempt) * time.Millisecond):// 1 attempt timeout 100ms, 2 attempt timeout 200ms, 3 attempt timeout 300ms ...
				if attempt == 6 {
					lostbytes += uint32(len(packets[i].Data)) // count of lost bytes
					i++
				}
				continue
			}

			if delivered {
				break
			}
		}
	}

	return lostbytes

}
