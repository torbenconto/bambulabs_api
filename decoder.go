package bambulabs_api

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Decoder struct {
	ams AMSDecoder
}

func NewDecoder(model Model, commandClient CommandClient) *Decoder {
	return &Decoder{
		ams: *NewAMSDecoder(model, commandClient),
	}
}

func (d *Decoder) Apply(p *printer, msg *protocol.Report) error { // mutates state based on protocol contents
	d.ams.Apply(p, msg)
	return nil
}

func decodeColor(raw string) color.RGBA {
	defaultColor := color.RGBA{}

	if raw == "" {
		return defaultColor
	}

	clr, ok := parseColor(raw)
	if ok {
		return clr
	}

	return defaultColor
}

func parseColor(s string) (color.RGBA, bool) {
	s = strings.TrimPrefix(s, "#")

	switch len(s) {
	case 6:
		r, err1 := strconv.ParseUint(s[0:2], 16, 8)
		g, err2 := strconv.ParseUint(s[2:4], 16, 8)
		b, err3 := strconv.ParseUint(s[4:6], 16, 8)
		if err1 != nil || err2 != nil || err3 != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, true

	case 8:
		r, err1 := strconv.ParseUint(s[0:2], 16, 8)
		g, err2 := strconv.ParseUint(s[2:4], 16, 8)
		b, err3 := strconv.ParseUint(s[4:6], 16, 8)
		a, err4 := strconv.ParseUint(s[6:8], 16, 8)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, true
	}

	return color.RGBA{}, false
}

func parseFloat32(raw string) float32 {
	conv, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return 0.0
	}

	return float32(conv)
}

func parseInt(raw string) int {
	conv, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0
	}

	return int(conv)
}
