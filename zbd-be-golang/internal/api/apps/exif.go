package apps

// 사진의 촬영 위치(EXIF GPS) — 올리기의 ?meta= 실험 기능. JPEG 만 읽는다(못 읽으면 nil — 지금 서비스도 위치가 없으면 null)

import (
	"encoding/binary"
	"os"

	"zbd.kr/be/internal/jsx"
)

func exifGPS(p string) any {
	b, err := os.ReadFile(p)
	if err != nil || len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return nil
	}
	i := 2
	for i+4 <= len(b) {
		if b[i] != 0xFF {
			return nil
		}
		mk := b[i+1]
		ln := int(binary.BigEndian.Uint16(b[i+2:]))
		if mk == 0xE1 && i+4+ln-2 <= len(b) && ln > 8 && string(b[i+4:i+10]) == "Exif\x00\x00" {
			return parseTiffGPS(b[i+10 : i+2+ln])
		}
		if mk == 0xDA {
			return nil
		}
		i += 2 + ln
	}
	return nil
}

func parseTiffGPS(t []byte) any {
	if len(t) < 8 {
		return nil
	}
	var bo binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return nil
	}
	type ent struct{ typ, cnt, val uint32 }
	readIFD := func(off uint32) map[uint16]ent {
		m := map[uint16]ent{}
		if int(off)+2 > len(t) {
			return m
		}
		n := int(bo.Uint16(t[off:]))
		for k := 0; k < n; k++ {
			e := int(off) + 2 + k*12
			if e+12 > len(t) {
				break
			}
			m[bo.Uint16(t[e:])] = ent{uint32(bo.Uint16(t[e+2:])), bo.Uint32(t[e+4:]), bo.Uint32(t[e+8:])}
		}
		return m
	}
	ifd0 := readIFD(bo.Uint32(t[4:]))
	g, ok := ifd0[0x8825]
	if !ok {
		return nil
	}
	gps := readIFD(g.val)
	coord := func(tag uint16) (float64, bool) {
		e, ok := gps[tag]
		if !ok || e.typ != 5 || e.cnt < 3 || int(e.val)+24 > len(t) {
			return 0, false
		}
		var v [3]float64
		for k := 0; k < 3; k++ {
			o := int(e.val) + k*8
			n, d := bo.Uint32(t[o:]), bo.Uint32(t[o+4:])
			if d == 0 {
				return 0, false
			}
			v[k] = float64(n) / float64(d)
		}
		return v[0] + v[1]/60 + v[2]/3600, true
	}
	ref := func(tag uint16) byte {
		e, ok := gps[tag]
		if !ok {
			return 0
		}
		if bo == binary.LittleEndian {
			return byte(e.val)
		}
		return byte(e.val >> 24)
	}
	lat, ok1 := coord(2)
	lng, ok2 := coord(4)
	if !ok1 || !ok2 {
		return nil
	}
	if ref(1) == 'S' {
		lat = -lat
	}
	if ref(3) == 'W' {
		lng = -lng
	}
	return jsx.Obj("lat", lat, "lng", lng)
}
