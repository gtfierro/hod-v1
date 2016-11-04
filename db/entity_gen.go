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
	var lqf uint32
	lqf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for lqf > 0 {
		lqf--
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
		case "ein":
			var daf uint32
			daf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && daf > 0 {
				z.InEdges = make(map[string][][4]byte, daf)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for daf > 0 {
				daf--
				var bzg string
				var bai [][4]byte
				bzg, err = dc.ReadString()
				if err != nil {
					return
				}
				var pks uint32
				pks, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(bai) >= int(pks) {
					bai = bai[:pks]
				} else {
					bai = make([][4]byte, pks)
				}
				for cmr := range bai {
					err = dc.ReadExactBytes(bai[cmr][:])
					if err != nil {
						return
					}
				}
				z.InEdges[bzg] = bai
			}
		case "eout":
			var jfb uint32
			jfb, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && jfb > 0 {
				z.OutEdges = make(map[string][][4]byte, jfb)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for jfb > 0 {
				jfb--
				var wht string
				var hct [][4]byte
				wht, err = dc.ReadString()
				if err != nil {
					return
				}
				var cxo uint32
				cxo, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(hct) >= int(cxo) {
					hct = hct[:cxo]
				} else {
					hct = make([][4]byte, cxo)
				}
				for cua := range hct {
					err = dc.ReadExactBytes(hct[cua][:])
					if err != nil {
						return
					}
				}
				z.OutEdges[wht] = hct
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
	// write "ein"
	err = en.Append(0xa3, 0x65, 0x69, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InEdges)))
	if err != nil {
		return
	}
	for bzg, bai := range z.InEdges {
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
	// write "eout"
	err = en.Append(0xa4, 0x65, 0x6f, 0x75, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.OutEdges)))
	if err != nil {
		return
	}
	for wht, hct := range z.OutEdges {
		err = en.WriteString(wht)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(hct)))
		if err != nil {
			return
		}
		for cua := range hct {
			err = en.WriteBytes(hct[cua][:])
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
	// map header, size 3
	// string "p"
	o = append(o, 0x83, 0xa1, 0x70)
	o = msgp.AppendBytes(o, z.PK[:])
	// string "ein"
	o = append(o, 0xa3, 0x65, 0x69, 0x6e)
	o = msgp.AppendMapHeader(o, uint32(len(z.InEdges)))
	for bzg, bai := range z.InEdges {
		o = msgp.AppendString(o, bzg)
		o = msgp.AppendArrayHeader(o, uint32(len(bai)))
		for cmr := range bai {
			o = msgp.AppendBytes(o, bai[cmr][:])
		}
	}
	// string "eout"
	o = append(o, 0xa4, 0x65, 0x6f, 0x75, 0x74)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutEdges)))
	for wht, hct := range z.OutEdges {
		o = msgp.AppendString(o, wht)
		o = msgp.AppendArrayHeader(o, uint32(len(hct)))
		for cua := range hct {
			o = msgp.AppendBytes(o, hct[cua][:])
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Entity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var eff uint32
	eff, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for eff > 0 {
		eff--
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
		case "ein":
			var rsw uint32
			rsw, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && rsw > 0 {
				z.InEdges = make(map[string][][4]byte, rsw)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for rsw > 0 {
				var bzg string
				var bai [][4]byte
				rsw--
				bzg, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var xpk uint32
				xpk, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(bai) >= int(xpk) {
					bai = bai[:xpk]
				} else {
					bai = make([][4]byte, xpk)
				}
				for cmr := range bai {
					bts, err = msgp.ReadExactBytes(bts, bai[cmr][:])
					if err != nil {
						return
					}
				}
				z.InEdges[bzg] = bai
			}
		case "eout":
			var dnj uint32
			dnj, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && dnj > 0 {
				z.OutEdges = make(map[string][][4]byte, dnj)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for dnj > 0 {
				var wht string
				var hct [][4]byte
				dnj--
				wht, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var obc uint32
				obc, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(hct) >= int(obc) {
					hct = hct[:obc]
				} else {
					hct = make([][4]byte, obc)
				}
				for cua := range hct {
					bts, err = msgp.ReadExactBytes(bts, hct[cua][:])
					if err != nil {
						return
					}
				}
				z.OutEdges[wht] = hct
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
	s = 1 + 2 + msgp.ArrayHeaderSize + (4 * (msgp.ByteSize)) + 4 + msgp.MapHeaderSize
	if z.InEdges != nil {
		for bzg, bai := range z.InEdges {
			_ = bai
			s += msgp.StringPrefixSize + len(bzg) + msgp.ArrayHeaderSize + (len(bai) * (4 * (msgp.ByteSize)))
		}
	}
	s += 5 + msgp.MapHeaderSize
	if z.OutEdges != nil {
		for wht, hct := range z.OutEdges {
			_ = hct
			s += msgp.StringPrefixSize + len(wht) + msgp.ArrayHeaderSize + (len(hct) * (4 * (msgp.ByteSize)))
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zpf uint32
	zpf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zpf > 0 {
		zpf--
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
			var rfe uint32
			rfe, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && rfe > 0 {
				z.Subjects = make(map[string]map[string]uint32, rfe)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for rfe > 0 {
				rfe--
				var kgt string
				var ema map[string]uint32
				kgt, err = dc.ReadString()
				if err != nil {
					return
				}
				var gmo uint32
				gmo, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if ema == nil && gmo > 0 {
					ema = make(map[string]uint32, gmo)
				} else if len(ema) > 0 {
					for key, _ := range ema {
						delete(ema, key)
					}
				}
				for gmo > 0 {
					gmo--
					var pez string
					var qke uint32
					pez, err = dc.ReadString()
					if err != nil {
						return
					}
					qke, err = dc.ReadUint32()
					if err != nil {
						return
					}
					ema[pez] = qke
				}
				z.Subjects[kgt] = ema
			}
		case "o":
			var taf uint32
			taf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && taf > 0 {
				z.Objects = make(map[string]map[string]uint32, taf)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for taf > 0 {
				taf--
				var qyh string
				var yzr map[string]uint32
				qyh, err = dc.ReadString()
				if err != nil {
					return
				}
				var eth uint32
				eth, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if yzr == nil && eth > 0 {
					yzr = make(map[string]uint32, eth)
				} else if len(yzr) > 0 {
					for key, _ := range yzr {
						delete(yzr, key)
					}
				}
				for eth > 0 {
					eth--
					var ywj string
					var jpj uint32
					ywj, err = dc.ReadString()
					if err != nil {
						return
					}
					jpj, err = dc.ReadUint32()
					if err != nil {
						return
					}
					yzr[ywj] = jpj
				}
				z.Objects[qyh] = yzr
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
	for kgt, ema := range z.Subjects {
		err = en.WriteString(kgt)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(ema)))
		if err != nil {
			return
		}
		for pez, qke := range ema {
			err = en.WriteString(pez)
			if err != nil {
				return
			}
			err = en.WriteUint32(qke)
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
	for qyh, yzr := range z.Objects {
		err = en.WriteString(qyh)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(yzr)))
		if err != nil {
			return
		}
		for ywj, jpj := range yzr {
			err = en.WriteString(ywj)
			if err != nil {
				return
			}
			err = en.WriteUint32(jpj)
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
	for kgt, ema := range z.Subjects {
		o = msgp.AppendString(o, kgt)
		o = msgp.AppendMapHeader(o, uint32(len(ema)))
		for pez, qke := range ema {
			o = msgp.AppendString(o, pez)
			o = msgp.AppendUint32(o, qke)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for qyh, yzr := range z.Objects {
		o = msgp.AppendString(o, qyh)
		o = msgp.AppendMapHeader(o, uint32(len(yzr)))
		for ywj, jpj := range yzr {
			o = msgp.AppendString(o, ywj)
			o = msgp.AppendUint32(o, jpj)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var sbz uint32
	sbz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for sbz > 0 {
		sbz--
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
			var rjx uint32
			rjx, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && rjx > 0 {
				z.Subjects = make(map[string]map[string]uint32, rjx)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for rjx > 0 {
				var kgt string
				var ema map[string]uint32
				rjx--
				kgt, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var awn uint32
				awn, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if ema == nil && awn > 0 {
					ema = make(map[string]uint32, awn)
				} else if len(ema) > 0 {
					for key, _ := range ema {
						delete(ema, key)
					}
				}
				for awn > 0 {
					var pez string
					var qke uint32
					awn--
					pez, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					qke, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					ema[pez] = qke
				}
				z.Subjects[kgt] = ema
			}
		case "o":
			var wel uint32
			wel, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && wel > 0 {
				z.Objects = make(map[string]map[string]uint32, wel)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for wel > 0 {
				var qyh string
				var yzr map[string]uint32
				wel--
				qyh, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var rbe uint32
				rbe, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if yzr == nil && rbe > 0 {
					yzr = make(map[string]uint32, rbe)
				} else if len(yzr) > 0 {
					for key, _ := range yzr {
						delete(yzr, key)
					}
				}
				for rbe > 0 {
					var ywj string
					var jpj uint32
					rbe--
					ywj, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					jpj, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					yzr[ywj] = jpj
				}
				z.Objects[qyh] = yzr
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
		for kgt, ema := range z.Subjects {
			_ = ema
			s += msgp.StringPrefixSize + len(kgt) + msgp.MapHeaderSize
			if ema != nil {
				for pez, qke := range ema {
					_ = qke
					s += msgp.StringPrefixSize + len(pez) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for qyh, yzr := range z.Objects {
			_ = yzr
			s += msgp.StringPrefixSize + len(qyh) + msgp.MapHeaderSize
			if yzr != nil {
				for ywj, jpj := range yzr {
					_ = jpj
					s += msgp.StringPrefixSize + len(ywj) + msgp.Uint32Size
				}
			}
		}
	}
	return
}
