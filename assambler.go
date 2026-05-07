package udplib

type assembler struct {
	chunks map[uint32][]byte
	final  uint32
	done   bool
}

func newAssembler() *assembler {
	return &assembler{
		chunks: make(map[uint32][]byte),
		final:  0,
		done:   false,
	}
}

func (a *assembler) add(p Packet) {
	a.chunks[p.Seq] = p.Data
	if p.IsFinal {
		a.done = true
		a.final = p.Seq
	}
}

func (a *assembler) ready() bool {
    if !a.done {
        return false  // not get final package yet
    }
    // check if every seq from 0 to final exists
    for i := uint32(0); i <= a.final; i++ {
        if _, exists := a.chunks[i]; !exists {
            return false  // missing a chunk
        }
    }
    return true
}

func (a *assembler) assemble() []byte {
	var res []byte

	for i := uint32(0); i <= a.final; i++ {
		res = append(res, a.chunks[0]...)
	}

	return res
}
