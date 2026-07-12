package bambulabs_api

import (
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Decoder struct{} // stateless

func (d *Decoder) Apply(state *State, msg *protocol.Report) error { // mutates state based on protocol contents
	if msg.Print == nil {
		return nil
	}

	// d.decodeAMS(state, msg)
	// d.decodeHMS(state, msg)
	// d.decodeExtruder(state, msg)
	// d.decodeNozzles(state, msg)

	return nil
}
