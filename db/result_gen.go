package db

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *QueryResult) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var cmr uint32
	cmr, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for cmr > 0 {
		cmr--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Rows":
			var ajw uint32
			ajw, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Rows) >= int(ajw) {
				z.Rows = z.Rows[:ajw]
			} else {
				z.Rows = make([]ResultMap, ajw)
			}
			for xvk := range z.Rows {
				var wht uint32
				wht, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if z.Rows[xvk] == nil && wht > 0 {
					z.Rows[xvk] = make(ResultMap, wht)
				} else if len(z.Rows[xvk]) > 0 {
					for key, _ := range z.Rows[xvk] {
						delete(z.Rows[xvk], key)
					}
				}
				for wht > 0 {
					wht--
					var bzg string
					var bai turtle.URI
					bzg, err = dc.ReadString()
					if err != nil {
						return
					}
					err = bai.DecodeMsg(dc)
					if err != nil {
						return
					}
					z.Rows[xvk][bzg] = bai
				}
			}
		case "Count":
			z.Count, err = dc.ReadInt()
			if err != nil {
				return
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
func (z *QueryResult) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Rows"
	err = en.Append(0x82, 0xa4, 0x52, 0x6f, 0x77, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Rows)))
	if err != nil {
		return
	}
	for xvk := range z.Rows {
		err = en.WriteMapHeader(uint32(len(z.Rows[xvk])))
		if err != nil {
			return
		}
		for bzg, bai := range z.Rows[xvk] {
			err = en.WriteString(bzg)
			if err != nil {
				return
			}
			err = bai.EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	// write "Count"
	err = en.Append(0xa5, 0x43, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Count)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *QueryResult) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Rows"
	o = append(o, 0x82, 0xa4, 0x52, 0x6f, 0x77, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Rows)))
	for xvk := range z.Rows {
		o = msgp.AppendMapHeader(o, uint32(len(z.Rows[xvk])))
		for bzg, bai := range z.Rows[xvk] {
			o = msgp.AppendString(o, bzg)
			o, err = bai.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "Count"
	o = append(o, 0xa5, 0x43, 0x6f, 0x75, 0x6e, 0x74)
	o = msgp.AppendInt(o, z.Count)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *QueryResult) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var hct uint32
	hct, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for hct > 0 {
		hct--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Rows":
			var cua uint32
			cua, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Rows) >= int(cua) {
				z.Rows = z.Rows[:cua]
			} else {
				z.Rows = make([]ResultMap, cua)
			}
			for xvk := range z.Rows {
				var xhx uint32
				xhx, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if z.Rows[xvk] == nil && xhx > 0 {
					z.Rows[xvk] = make(ResultMap, xhx)
				} else if len(z.Rows[xvk]) > 0 {
					for key, _ := range z.Rows[xvk] {
						delete(z.Rows[xvk], key)
					}
				}
				for xhx > 0 {
					var bzg string
					var bai turtle.URI
					xhx--
					bzg, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					bts, err = bai.UnmarshalMsg(bts)
					if err != nil {
						return
					}
					z.Rows[xvk][bzg] = bai
				}
			}
		case "Count":
			z.Count, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
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

func (z *QueryResult) Msgsize() (s int) {
	s = 1 + 5 + msgp.ArrayHeaderSize
	for xvk := range z.Rows {
		s += msgp.MapHeaderSize
		if z.Rows[xvk] != nil {
			for bzg, bai := range z.Rows[xvk] {
				_ = bai
				s += msgp.StringPrefixSize + len(bzg) + bai.Msgsize()
			}
		}
	}
	s += 6 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ResultMap) DecodeMsg(dc *msgp.Reader) (err error) {
	var cxo uint32
	cxo, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && cxo > 0 {
		(*z) = make(ResultMap, cxo)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for cxo > 0 {
		cxo--
		var pks string
		var jfb turtle.URI
		pks, err = dc.ReadString()
		if err != nil {
			return
		}
		err = jfb.DecodeMsg(dc)
		if err != nil {
			return
		}
		(*z)[pks] = jfb
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ResultMap) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for eff, rsw := range z {
		err = en.WriteString(eff)
		if err != nil {
			return
		}
		err = rsw.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ResultMap) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for eff, rsw := range z {
		o = msgp.AppendString(o, eff)
		o, err = rsw.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ResultMap) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var obc uint32
	obc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && obc > 0 {
		(*z) = make(ResultMap, obc)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for obc > 0 {
		var xpk string
		var dnj turtle.URI
		obc--
		xpk, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		bts, err = dnj.UnmarshalMsg(bts)
		if err != nil {
			return
		}
		(*z)[xpk] = dnj
	}
	o = bts
	return
}

func (z ResultMap) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for snv, kgt := range z {
			_ = kgt
			s += msgp.StringPrefixSize + len(snv) + kgt.Msgsize()
		}
	}
	return
}
