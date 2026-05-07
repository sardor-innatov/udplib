package udplib

import (
	"encoding/binary"
	"net"
)

type Packet struct {
	Seq      uint32
	Ack      uint32
	IsAck    bool
	IsNack   bool 
	IsFinal  bool
	Checksum uint16
	Data     []byte
	Addr     net.Addr
}

// this function serializes p Packet
// into bytes and returns buffer []byte
//   - [0-3]   - Seq (4 bytes)
//   - [4-7]   - Ack (4 bytes)
//   - [8]     - IsAck (1 byte)
//   - [9]	   - IsNack (1 byte)
//   - [10]	   - IsFinal (1 byte)
//   - [11-12]  - Checksum (2 bytes)
//   - [13-14] - Data Len (2 bytes)
//   - [15-N]  - The Payload (actual data)
func serialize(p Packet) []byte {

	// creating buffer with length 15 byte (Header) + Payload(p.Data) length
	buf := make([]byte, 15+len(p.Data))

	// write Seq (4 byte) into bytes 0-3
	binary.BigEndian.PutUint32(buf[0:4], p.Seq)

	// write Ack (4 byte) into bytes 4-7
	binary.BigEndian.PutUint32(buf[4:8], p.Ack)

	// write IsAck (1 byte) into byte 8
	if p.IsAck {
		buf[8] = 1
	} else {
		buf[8] = 0
	}

	// write IsFinal (1 byte) into byte 9
	if p.IsNack {
		buf[9] = 1
	} else {
		buf[9] = 0
	}

	// write IsFinal (1 byte) into byte 10
	if p.IsFinal {
		buf[10] = 1
	} else {
		buf[10] = 0
	}

	// write Checksum (2 byte) int bytes 11-12
	binary.BigEndian.PutUint16(buf[11:13], p.Checksum)

	// write data length into bytes 13-14
	// so receiver knows how many bytes to read after the header
	binary.BigEndian.PutUint16(buf[13:15], uint16(len(p.Data)))

	// copy actual data starting at byte 15
	copy(buf[15:], p.Data)

	return buf
}

// This function deserialize buf.
// Returns Packet and nil err if error is not resived.
func deserialize(buf []byte) Packet {
	p := Packet{}

	p.Seq = binary.BigEndian.Uint32(buf[0:4])
	p.Ack = binary.BigEndian.Uint32(buf[4:8])
	p.IsAck = buf[8] == 1
	p.IsNack = buf[9] == 1
	p.IsFinal = buf[10] == 1

	p.Checksum = binary.BigEndian.Uint16(buf[11:13])

	dataLen := binary.BigEndian.Uint16(buf[13:15])
	p.Data = buf[15 : 15+dataLen]

	return p
}
