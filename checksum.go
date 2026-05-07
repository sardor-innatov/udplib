package udplib

// Valid checks if the packet data is corrupted or not.
// It compares the calculated checksum of the current data with the stored p.Checksum.
// Returns true if checksums are equal, otherwise returns false.
func (p *Packet) Valid() bool {
    return calculateChecksum(p.Data) == p.Checksum
}

// This func calculatest all bytes of data.
// Returns sum of bytes
func calculateChecksum(data []byte) uint16 {
    var sum uint32
    
    // sum every 2 bytes as a 16 bit number
    for i := 0; i+1 < len(data); i += 2 {
        sum += uint32(data[i])<<8 | uint32(data[i+1])
    }
    
    // odd byte left over
    if len(data)%2 != 0 {
        sum += uint32(data[len(data)-1]) << 8
    }

    // fold carry bits back in — ones complement
    for sum>>16 != 0 {
        sum = (sum & 0xffff) + (sum >> 16)
    }

    // flip all bits — ones complement
    return ^uint16(sum)
}