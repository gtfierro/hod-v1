package db

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Entity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var wht uint32
	wht, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for wht > 0 {
		wht--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			err = dc.ReadExactBytes(z.PK[:])
			if err != nil {
				return
			}
		case "e":
			var hct uint32
			hct, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Edges == nil && hct > 0 {
				z.Edges = make(map[string][][4]byte, hct)
			} else if len(z.Edges) > 0 {
				for key, _ := range z.Edges {
					delete(z.Edges, key)
				}
			}
			for hct > 0 {
				hct--
				var bzg string
				var bai [][4]byte
				bzg, err = dc.ReadString()
				if err != nil {
					return
				}
				var cua uint32
				cua, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(bai) >= int(cua) {
					bai = bai[:cua]
				} else {
					bai = make([][4]byte, cua)
				}
				for cmr := range bai {
					err = dc.ReadExactBytes(bai[cmr][:])
					if err != nil {
						return
					}
				}
				z.Edges[bzg] = bai
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Entity) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "p"
	err = en.Append(0x82, 0xa1, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.PK[:])
	if err != nil {
		return
	}
	// write "e"
	err = en.Append(0xa1, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Edges)))
	if err != nil {
		return
	}
	for bzg, bai := range z.Edges {
		err = en.WriteString(bzg)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(bai)))
		if err != nil {
			return
		}
		for cmr := range bai {
			err = en.WriteBytes(bai[cmr][:])
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Entity) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "p"
	o = append(o, 0x82, 0xa1, 0x70)
	o = msgp.AppendBytes(o, z.PK[:])
	// string "e"
	o = append(o, 0xa1, 0x65)
	o = msgp.AppendMapHeader(o, uint32(len(z.Edges)))
	for bzg, bai := range z.Edges {
		o = msgp.AppendString(o, bzg)
		o = msgp.AppendArrayHeader(o, uint32(len(bai)))
		for cmr := range bai {
			o = msgp.AppendBytes(o, bai[cmr][:])
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Entity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var xhx uint32
	xhx, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for xhx > 0 {
		xhx--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			bts, err = msgp.ReadExactBytes(bts, z.PK[:])
			if err != nil {
				return
			}
		case "e":
			var lqf uint32
			lqf, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Edges == nil && lqf > 0 {
				z.Edges = make(map[string][][4]byte, lqf)
			} else if len(z.Edges) > 0 {
				for key, _ := range z.Edges {
					delete(z.Edges, key)
				}
			}
			for lqf > 0 {
				var bzg string
				var bai [][4]byte
				lqf--
				bzg, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var daf uint32
				daf, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(bai) >= int(daf) {
					bai = bai[:daf]
				} else {
					bai = make([][4]byte, daf)
				}
				for cmr := range bai {
					bts, err = msgp.ReadExactBytes(bts, bai[cmr][:])
					if err != nil {
						return
					}
				}
				z.Edges[bzg] = bai
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Entity) Msgsize() (s int) {
	s = 1 + 2 + msgp.ArrayHeaderSize + (4 * (msgp.ByteSize)) + 2 + msgp.MapHeaderSize
	if z.Edges != nil {
		for bzg, bai := range z.Edges {
			_ = bai
			s += msgp.StringPrefixSize + len(bzg) + msgp.ArrayHeaderSize + (len(bai) * (4 * (msgp.ByteSize)))
		}
	}
	return
}
