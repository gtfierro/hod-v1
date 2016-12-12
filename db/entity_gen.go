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
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
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
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && msz > 0 {
				z.InEdges = make(map[string][][4]byte, msz)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for msz > 0 {
				msz--
				var bzg string
				var bai [][4]byte
				bzg, err = dc.ReadString()
				if err != nil {
					return
				}
				var xsz uint32
				xsz, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(bai) >= int(xsz) {
					bai = bai[:xsz]
				} else {
					bai = make([][4]byte, xsz)
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
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && msz > 0 {
				z.OutEdges = make(map[string][][4]byte, msz)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for msz > 0 {
				msz--
				var wht string
				var hct [][4]byte
				wht, err = dc.ReadString()
				if err != nil {
					return
				}
				var xsz uint32
				xsz, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(hct) >= int(xsz) {
					hct = hct[:xsz]
				} else {
					hct = make([][4]byte, xsz)
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
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
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
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && msz > 0 {
				z.InEdges = make(map[string][][4]byte, msz)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for msz > 0 {
				var bzg string
				var bai [][4]byte
				msz--
				bzg, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var xsz uint32
				xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(bai) >= int(xsz) {
					bai = bai[:xsz]
				} else {
					bai = make([][4]byte, xsz)
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
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && msz > 0 {
				z.OutEdges = make(map[string][][4]byte, msz)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for msz > 0 {
				var wht string
				var hct [][4]byte
				msz--
				wht, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var xsz uint32
				xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(hct) >= int(xsz) {
					hct = hct[:xsz]
				} else {
					hct = make([][4]byte, xsz)
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
func (z *NamespaceIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var msz uint32
	msz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(NamespaceIndex, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		msz--
		var pks string
		var jfb string
		pks, err = dc.ReadString()
		if err != nil {
			return
		}
		jfb, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[pks] = jfb
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z NamespaceIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for cxo, eff := range z {
		err = en.WriteString(cxo)
		if err != nil {
			return
		}
		err = en.WriteString(eff)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z NamespaceIndex) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for cxo, eff := range z {
		o = msgp.AppendString(o, cxo)
		o = msgp.AppendString(o, eff)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NamespaceIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var msz uint32
	msz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(NamespaceIndex, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		var rsw string
		var xpk string
		msz--
		rsw, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		xpk, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[rsw] = xpk
	}
	o = bts
	return
}

func (z NamespaceIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for dnj, obc := range z {
			_ = obc
			s += msgp.StringPrefixSize + len(dnj) + msgp.StringPrefixSize + len(obc)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var msz uint32
	msz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(PredIndex, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		msz--
		var ema string
		var pez *PredicateEntity
		ema, err = dc.ReadString()
		if err != nil {
			return
		}
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			pez = nil
		} else {
			if pez == nil {
				pez = new(PredicateEntity)
			}
			err = pez.DecodeMsg(dc)
			if err != nil {
				return
			}
		}
		(*z)[ema] = pez
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PredIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for qke, qyh := range z {
		err = en.WriteString(qke)
		if err != nil {
			return
		}
		if qyh == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = qyh.EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z PredIndex) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for qke, qyh := range z {
		o = msgp.AppendString(o, qke)
		if qyh == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = qyh.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var msz uint32
	msz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(PredIndex, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		var yzr string
		var ywj *PredicateEntity
		msz--
		yzr, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			ywj = nil
		} else {
			if ywj == nil {
				ywj = new(PredicateEntity)
			}
			bts, err = ywj.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
		(*z)[yzr] = ywj
	}
	o = bts
	return
}

func (z PredIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for jpj, zpf := range z {
			_ = zpf
			s += msgp.StringPrefixSize + len(jpj)
			if zpf == nil {
				s += msgp.NilSize
			} else {
				s += zpf.Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
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
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && msz > 0 {
				z.Subjects = make(map[string]map[string]uint32, msz)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for msz > 0 {
				msz--
				var gmo string
				var taf map[string]uint32
				gmo, err = dc.ReadString()
				if err != nil {
					return
				}
				var msz uint32
				msz, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if taf == nil && msz > 0 {
					taf = make(map[string]uint32, msz)
				} else if len(taf) > 0 {
					for key, _ := range taf {
						delete(taf, key)
					}
				}
				for msz > 0 {
					msz--
					var eth string
					var sbz uint32
					eth, err = dc.ReadString()
					if err != nil {
						return
					}
					sbz, err = dc.ReadUint32()
					if err != nil {
						return
					}
					taf[eth] = sbz
				}
				z.Subjects[gmo] = taf
			}
		case "o":
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && msz > 0 {
				z.Objects = make(map[string]map[string]uint32, msz)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for msz > 0 {
				msz--
				var rjx string
				var awn map[string]uint32
				rjx, err = dc.ReadString()
				if err != nil {
					return
				}
				var msz uint32
				msz, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if awn == nil && msz > 0 {
					awn = make(map[string]uint32, msz)
				} else if len(awn) > 0 {
					for key, _ := range awn {
						delete(awn, key)
					}
				}
				for msz > 0 {
					msz--
					var wel string
					var rbe uint32
					wel, err = dc.ReadString()
					if err != nil {
						return
					}
					rbe, err = dc.ReadUint32()
					if err != nil {
						return
					}
					awn[wel] = rbe
				}
				z.Objects[rjx] = awn
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
	for gmo, taf := range z.Subjects {
		err = en.WriteString(gmo)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(taf)))
		if err != nil {
			return
		}
		for eth, sbz := range taf {
			err = en.WriteString(eth)
			if err != nil {
				return
			}
			err = en.WriteUint32(sbz)
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
	for rjx, awn := range z.Objects {
		err = en.WriteString(rjx)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(awn)))
		if err != nil {
			return
		}
		for wel, rbe := range awn {
			err = en.WriteString(wel)
			if err != nil {
				return
			}
			err = en.WriteUint32(rbe)
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
	for gmo, taf := range z.Subjects {
		o = msgp.AppendString(o, gmo)
		o = msgp.AppendMapHeader(o, uint32(len(taf)))
		for eth, sbz := range taf {
			o = msgp.AppendString(o, eth)
			o = msgp.AppendUint32(o, sbz)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for rjx, awn := range z.Objects {
		o = msgp.AppendString(o, rjx)
		o = msgp.AppendMapHeader(o, uint32(len(awn)))
		for wel, rbe := range awn {
			o = msgp.AppendString(o, wel)
			o = msgp.AppendUint32(o, rbe)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
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
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && msz > 0 {
				z.Subjects = make(map[string]map[string]uint32, msz)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for msz > 0 {
				var gmo string
				var taf map[string]uint32
				msz--
				gmo, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var msz uint32
				msz, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if taf == nil && msz > 0 {
					taf = make(map[string]uint32, msz)
				} else if len(taf) > 0 {
					for key, _ := range taf {
						delete(taf, key)
					}
				}
				for msz > 0 {
					var eth string
					var sbz uint32
					msz--
					eth, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					sbz, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					taf[eth] = sbz
				}
				z.Subjects[gmo] = taf
			}
		case "o":
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && msz > 0 {
				z.Objects = make(map[string]map[string]uint32, msz)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for msz > 0 {
				var rjx string
				var awn map[string]uint32
				msz--
				rjx, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var msz uint32
				msz, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if awn == nil && msz > 0 {
					awn = make(map[string]uint32, msz)
				} else if len(awn) > 0 {
					for key, _ := range awn {
						delete(awn, key)
					}
				}
				for msz > 0 {
					var wel string
					var rbe uint32
					msz--
					wel, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					rbe, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					awn[wel] = rbe
				}
				z.Objects[rjx] = awn
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
		for gmo, taf := range z.Subjects {
			_ = taf
			s += msgp.StringPrefixSize + len(gmo) + msgp.MapHeaderSize
			if taf != nil {
				for eth, sbz := range taf {
					_ = sbz
					s += msgp.StringPrefixSize + len(eth) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for rjx, awn := range z.Objects {
			_ = awn
			s += msgp.StringPrefixSize + len(rjx) + msgp.MapHeaderSize
			if awn != nil {
				for wel, rbe := range awn {
					_ = rbe
					s += msgp.StringPrefixSize + len(wel) + msgp.Uint32Size
				}
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RelshipIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var msz uint32
	msz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(RelshipIndex, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		msz--
		var elx string
		var bal string
		elx, err = dc.ReadString()
		if err != nil {
			return
		}
		bal, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[elx] = bal
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RelshipIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for jqz, kct := range z {
		err = en.WriteString(jqz)
		if err != nil {
			return
		}
		err = en.WriteString(kct)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RelshipIndex) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for jqz, kct := range z {
		o = msgp.AppendString(o, jqz)
		o = msgp.AppendString(o, kct)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RelshipIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var msz uint32
	msz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(RelshipIndex, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		var tmt string
		var tco string
		msz--
		tmt, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		tco, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[tmt] = tco
	}
	o = bts
	return
}

func (z RelshipIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for ana, tyy := range z {
			_ = tyy
			s += msgp.StringPrefixSize + len(ana) + msgp.StringPrefixSize + len(tyy)
		}
	}
	return
}
