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

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var kgt uint32
	kgt, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for kgt > 0 {
		kgt--
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
		case "s":
			var ema uint32
			ema, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && ema > 0 {
				z.Subjects = make(map[string]map[string]uint32, ema)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for ema > 0 {
				ema--
				var jfb string
				var cxo map[string]uint32
				jfb, err = dc.ReadString()
				if err != nil {
					return
				}
				var pez uint32
				pez, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if cxo == nil && pez > 0 {
					cxo = make(map[string]uint32, pez)
				} else if len(cxo) > 0 {
					for key, _ := range cxo {
						delete(cxo, key)
					}
				}
				for pez > 0 {
					pez--
					var eff string
					var rsw uint32
					eff, err = dc.ReadString()
					if err != nil {
						return
					}
					rsw, err = dc.ReadUint32()
					if err != nil {
						return
					}
					cxo[eff] = rsw
				}
				z.Subjects[jfb] = cxo
			}
		case "o":
			var qke uint32
			qke, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && qke > 0 {
				z.Objects = make(map[string]map[string]uint32, qke)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for qke > 0 {
				qke--
				var xpk string
				var dnj map[string]uint32
				xpk, err = dc.ReadString()
				if err != nil {
					return
				}
				var qyh uint32
				qyh, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if dnj == nil && qyh > 0 {
					dnj = make(map[string]uint32, qyh)
				} else if len(dnj) > 0 {
					for key, _ := range dnj {
						delete(dnj, key)
					}
				}
				for qyh > 0 {
					qyh--
					var obc string
					var snv uint32
					obc, err = dc.ReadString()
					if err != nil {
						return
					}
					snv, err = dc.ReadUint32()
					if err != nil {
						return
					}
					dnj[obc] = snv
				}
				z.Objects[xpk] = dnj
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
func (z *PredicateEntity) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "p"
	err = en.Append(0x83, 0xa1, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.PK[:])
	if err != nil {
		return
	}
	// write "s"
	err = en.Append(0xa1, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Subjects)))
	if err != nil {
		return
	}
	for jfb, cxo := range z.Subjects {
		err = en.WriteString(jfb)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(cxo)))
		if err != nil {
			return
		}
		for eff, rsw := range cxo {
			err = en.WriteString(eff)
			if err != nil {
				return
			}
			err = en.WriteUint32(rsw)
			if err != nil {
				return
			}
		}
	}
	// write "o"
	err = en.Append(0xa1, 0x6f)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Objects)))
	if err != nil {
		return
	}
	for xpk, dnj := range z.Objects {
		err = en.WriteString(xpk)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(dnj)))
		if err != nil {
			return
		}
		for obc, snv := range dnj {
			err = en.WriteString(obc)
			if err != nil {
				return
			}
			err = en.WriteUint32(snv)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PredicateEntity) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "p"
	o = append(o, 0x83, 0xa1, 0x70)
	o = msgp.AppendBytes(o, z.PK[:])
	// string "s"
	o = append(o, 0xa1, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Subjects)))
	for jfb, cxo := range z.Subjects {
		o = msgp.AppendString(o, jfb)
		o = msgp.AppendMapHeader(o, uint32(len(cxo)))
		for eff, rsw := range cxo {
			o = msgp.AppendString(o, eff)
			o = msgp.AppendUint32(o, rsw)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for xpk, dnj := range z.Objects {
		o = msgp.AppendString(o, xpk)
		o = msgp.AppendMapHeader(o, uint32(len(dnj)))
		for obc, snv := range dnj {
			o = msgp.AppendString(o, obc)
			o = msgp.AppendUint32(o, snv)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var yzr uint32
	yzr, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for yzr > 0 {
		yzr--
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
		case "s":
			var ywj uint32
			ywj, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && ywj > 0 {
				z.Subjects = make(map[string]map[string]uint32, ywj)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for ywj > 0 {
				var jfb string
				var cxo map[string]uint32
				ywj--
				jfb, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var jpj uint32
				jpj, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if cxo == nil && jpj > 0 {
					cxo = make(map[string]uint32, jpj)
				} else if len(cxo) > 0 {
					for key, _ := range cxo {
						delete(cxo, key)
					}
				}
				for jpj > 0 {
					var eff string
					var rsw uint32
					jpj--
					eff, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					rsw, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					cxo[eff] = rsw
				}
				z.Subjects[jfb] = cxo
			}
		case "o":
			var zpf uint32
			zpf, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && zpf > 0 {
				z.Objects = make(map[string]map[string]uint32, zpf)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zpf > 0 {
				var xpk string
				var dnj map[string]uint32
				zpf--
				xpk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var rfe uint32
				rfe, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if dnj == nil && rfe > 0 {
					dnj = make(map[string]uint32, rfe)
				} else if len(dnj) > 0 {
					for key, _ := range dnj {
						delete(dnj, key)
					}
				}
				for rfe > 0 {
					var obc string
					var snv uint32
					rfe--
					obc, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					snv, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					dnj[obc] = snv
				}
				z.Objects[xpk] = dnj
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

func (z *PredicateEntity) Msgsize() (s int) {
	s = 1 + 2 + msgp.ArrayHeaderSize + (4 * (msgp.ByteSize)) + 2 + msgp.MapHeaderSize
	if z.Subjects != nil {
		for jfb, cxo := range z.Subjects {
			_ = cxo
			s += msgp.StringPrefixSize + len(jfb) + msgp.MapHeaderSize
			if cxo != nil {
				for eff, rsw := range cxo {
					_ = rsw
					s += msgp.StringPrefixSize + len(eff) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for xpk, dnj := range z.Objects {
			_ = dnj
			s += msgp.StringPrefixSize + len(xpk) + msgp.MapHeaderSize
			if dnj != nil {
				for obc, snv := range dnj {
					_ = snv
					s += msgp.StringPrefixSize + len(obc) + msgp.Uint32Size
				}
			}
		}
	}
	return
}
