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
	var zcmr uint32
	zcmr, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zcmr > 0 {
		zcmr--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Rows":
			var zajw uint32
			zajw, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Rows) >= int(zajw) {
				z.Rows = (z.Rows)[:zajw]
			} else {
				z.Rows = make([]ResultMap, zajw)
			}
			for zxvk := range z.Rows {
				var zwht uint32
				zwht, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if z.Rows[zxvk] == nil && zwht > 0 {
					z.Rows[zxvk] = make(ResultMap, zwht)
				} else if len(z.Rows[zxvk]) > 0 {
					for key, _ := range z.Rows[zxvk] {
						delete(z.Rows[zxvk], key)
					}
				}
				for zwht > 0 {
					zwht--
					var zbzg string
					var zbai turtle.URI
					zbzg, err = dc.ReadString()
					if err != nil {
						return
					}
					err = zbai.DecodeMsg(dc)
					if err != nil {
						return
					}
					z.Rows[zxvk][zbzg] = zbai
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
	for zxvk := range z.Rows {
		err = en.WriteMapHeader(uint32(len(z.Rows[zxvk])))
		if err != nil {
			return
		}
		for zbzg, zbai := range z.Rows[zxvk] {
			err = en.WriteString(zbzg)
			if err != nil {
				return
			}
			err = zbai.EncodeMsg(en)
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
	for zxvk := range z.Rows {
		o = msgp.AppendMapHeader(o, uint32(len(z.Rows[zxvk])))
		for zbzg, zbai := range z.Rows[zxvk] {
			o = msgp.AppendString(o, zbzg)
			o, err = zbai.MarshalMsg(o)
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
	var zhct uint32
	zhct, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Rows":
			var zcua uint32
			zcua, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Rows) >= int(zcua) {
				z.Rows = (z.Rows)[:zcua]
			} else {
				z.Rows = make([]ResultMap, zcua)
			}
			for zxvk := range z.Rows {
				var zxhx uint32
				zxhx, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if z.Rows[zxvk] == nil && zxhx > 0 {
					z.Rows[zxvk] = make(ResultMap, zxhx)
				} else if len(z.Rows[zxvk]) > 0 {
					for key, _ := range z.Rows[zxvk] {
						delete(z.Rows[zxvk], key)
					}
				}
				for zxhx > 0 {
					var zbzg string
					var zbai turtle.URI
					zxhx--
					zbzg, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					bts, err = zbai.UnmarshalMsg(bts)
					if err != nil {
						return
					}
					z.Rows[zxvk][zbzg] = zbai
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *QueryResult) Msgsize() (s int) {
	s = 1 + 5 + msgp.ArrayHeaderSize
	for zxvk := range z.Rows {
		s += msgp.MapHeaderSize
		if z.Rows[zxvk] != nil {
			for zbzg, zbai := range z.Rows[zxvk] {
				_ = zbai
				s += msgp.StringPrefixSize + len(zbzg) + zbai.Msgsize()
			}
		}
	}
	s += 6 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ResultMap) DecodeMsg(dc *msgp.Reader) (err error) {
	var zcxo uint32
	zcxo, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zcxo > 0 {
		(*z) = make(ResultMap, zcxo)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zcxo > 0 {
		zcxo--
		var zpks string
		var zjfb turtle.URI
		zpks, err = dc.ReadString()
		if err != nil {
			return
		}
		err = zjfb.DecodeMsg(dc)
		if err != nil {
			return
		}
		(*z)[zpks] = zjfb
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ResultMap) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zeff, zrsw := range z {
		err = en.WriteString(zeff)
		if err != nil {
			return
		}
		err = zrsw.EncodeMsg(en)
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
	for zeff, zrsw := range z {
		o = msgp.AppendString(o, zeff)
		o, err = zrsw.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ResultMap) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zobc uint32
	zobc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zobc > 0 {
		(*z) = make(ResultMap, zobc)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zobc > 0 {
		var zxpk string
		var zdnj turtle.URI
		zobc--
		zxpk, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		bts, err = zdnj.UnmarshalMsg(bts)
		if err != nil {
			return
		}
		(*z)[zxpk] = zdnj
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ResultMap) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zsnv, zkgt := range z {
			_ = zkgt
			s += msgp.StringPrefixSize + len(zsnv) + zkgt.Msgsize()
		}
	}
	return
}
