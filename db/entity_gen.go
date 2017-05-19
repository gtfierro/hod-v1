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
	var zhct uint32
	zhct, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			err = z.PK.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "i":
			var zcua uint32
			zcua, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && zcua > 0 {
				z.InEdges = make(map[string][]Key, zcua)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zcua > 0 {
				zcua--
				var zxvk string
				var zbzg []Key
				zxvk, err = dc.ReadString()
				if err != nil {
					return
				}
				var zxhx uint32
				zxhx, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zbzg) >= int(zxhx) {
					zbzg = (zbzg)[:zxhx]
				} else {
					zbzg = make([]Key, zxhx)
				}
				for zbai := range zbzg {
					err = zbzg[zbai].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.InEdges[zxvk] = zbzg
			}
		case "o":
			var zlqf uint32
			zlqf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && zlqf > 0 {
				z.OutEdges = make(map[string][]Key, zlqf)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zlqf > 0 {
				zlqf--
				var zcmr string
				var zajw []Key
				zcmr, err = dc.ReadString()
				if err != nil {
					return
				}
				var zdaf uint32
				zdaf, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zajw) >= int(zdaf) {
					zajw = (zajw)[:zdaf]
				} else {
					zajw = make([]Key, zdaf)
				}
				for zwht := range zajw {
					err = zajw[zwht].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.OutEdges[zcmr] = zajw
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
	err = z.PK.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "i"
	err = en.Append(0xa1, 0x69)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InEdges)))
	if err != nil {
		return
	}
	for zxvk, zbzg := range z.InEdges {
		err = en.WriteString(zxvk)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zbzg)))
		if err != nil {
			return
		}
		for zbai := range zbzg {
			err = zbzg[zbai].EncodeMsg(en)
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
	err = en.WriteMapHeader(uint32(len(z.OutEdges)))
	if err != nil {
		return
	}
	for zcmr, zajw := range z.OutEdges {
		err = en.WriteString(zcmr)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zajw)))
		if err != nil {
			return
		}
		for zwht := range zajw {
			err = zajw[zwht].EncodeMsg(en)
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
	o, err = z.PK.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "i"
	o = append(o, 0xa1, 0x69)
	o = msgp.AppendMapHeader(o, uint32(len(z.InEdges)))
	for zxvk, zbzg := range z.InEdges {
		o = msgp.AppendString(o, zxvk)
		o = msgp.AppendArrayHeader(o, uint32(len(zbzg)))
		for zbai := range zbzg {
			o, err = zbzg[zbai].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutEdges)))
	for zcmr, zajw := range z.OutEdges {
		o = msgp.AppendString(o, zcmr)
		o = msgp.AppendArrayHeader(o, uint32(len(zajw)))
		for zwht := range zajw {
			o, err = zajw[zwht].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Entity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zpks uint32
	zpks, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zpks > 0 {
		zpks--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			bts, err = z.PK.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "i":
			var zjfb uint32
			zjfb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && zjfb > 0 {
				z.InEdges = make(map[string][]Key, zjfb)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zjfb > 0 {
				var zxvk string
				var zbzg []Key
				zjfb--
				zxvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zcxo uint32
				zcxo, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zbzg) >= int(zcxo) {
					zbzg = (zbzg)[:zcxo]
				} else {
					zbzg = make([]Key, zcxo)
				}
				for zbai := range zbzg {
					bts, err = zbzg[zbai].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.InEdges[zxvk] = zbzg
			}
		case "o":
			var zeff uint32
			zeff, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && zeff > 0 {
				z.OutEdges = make(map[string][]Key, zeff)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zeff > 0 {
				var zcmr string
				var zajw []Key
				zeff--
				zcmr, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zrsw uint32
				zrsw, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zajw) >= int(zrsw) {
					zajw = (zajw)[:zrsw]
				} else {
					zajw = make([]Key, zrsw)
				}
				for zwht := range zajw {
					bts, err = zajw[zwht].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.OutEdges[zcmr] = zajw
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
func (z *Entity) Msgsize() (s int) {
	s = 1 + 2 + z.PK.Msgsize() + 2 + msgp.MapHeaderSize
	if z.InEdges != nil {
		for zxvk, zbzg := range z.InEdges {
			_ = zbzg
			s += msgp.StringPrefixSize + len(zxvk) + msgp.ArrayHeaderSize
			for zbai := range zbzg {
				s += zbzg[zbai].Msgsize()
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.OutEdges != nil {
		for zcmr, zajw := range z.OutEdges {
			_ = zajw
			s += msgp.StringPrefixSize + len(zcmr) + msgp.ArrayHeaderSize
			for zwht := range zajw {
				s += zajw[zwht].Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *EntityExtendedIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zpez uint32
	zpez, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zpez > 0 {
		zpez--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			err = z.PK.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "i+":
			var zqke uint32
			zqke, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InPlusEdges == nil && zqke > 0 {
				z.InPlusEdges = make(map[string][]Key, zqke)
			} else if len(z.InPlusEdges) > 0 {
				for key, _ := range z.InPlusEdges {
					delete(z.InPlusEdges, key)
				}
			}
			for zqke > 0 {
				zqke--
				var zxpk string
				var zdnj []Key
				zxpk, err = dc.ReadString()
				if err != nil {
					return
				}
				var zqyh uint32
				zqyh, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zdnj) >= int(zqyh) {
					zdnj = (zdnj)[:zqyh]
				} else {
					zdnj = make([]Key, zqyh)
				}
				for zobc := range zdnj {
					err = zdnj[zobc].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.InPlusEdges[zxpk] = zdnj
			}
		case "o+":
			var zyzr uint32
			zyzr, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutPlusEdges == nil && zyzr > 0 {
				z.OutPlusEdges = make(map[string][]Key, zyzr)
			} else if len(z.OutPlusEdges) > 0 {
				for key, _ := range z.OutPlusEdges {
					delete(z.OutPlusEdges, key)
				}
			}
			for zyzr > 0 {
				zyzr--
				var zsnv string
				var zkgt []Key
				zsnv, err = dc.ReadString()
				if err != nil {
					return
				}
				var zywj uint32
				zywj, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zkgt) >= int(zywj) {
					zkgt = (zkgt)[:zywj]
				} else {
					zkgt = make([]Key, zywj)
				}
				for zema := range zkgt {
					err = zkgt[zema].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.OutPlusEdges[zsnv] = zkgt
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
func (z *EntityExtendedIndex) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "p"
	err = en.Append(0x83, 0xa1, 0x70)
	if err != nil {
		return err
	}
	err = z.PK.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "i+"
	err = en.Append(0xa2, 0x69, 0x2b)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InPlusEdges)))
	if err != nil {
		return
	}
	for zxpk, zdnj := range z.InPlusEdges {
		err = en.WriteString(zxpk)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zdnj)))
		if err != nil {
			return
		}
		for zobc := range zdnj {
			err = zdnj[zobc].EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	// write "o+"
	err = en.Append(0xa2, 0x6f, 0x2b)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.OutPlusEdges)))
	if err != nil {
		return
	}
	for zsnv, zkgt := range z.OutPlusEdges {
		err = en.WriteString(zsnv)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zkgt)))
		if err != nil {
			return
		}
		for zema := range zkgt {
			err = zkgt[zema].EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *EntityExtendedIndex) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "p"
	o = append(o, 0x83, 0xa1, 0x70)
	o, err = z.PK.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "i+"
	o = append(o, 0xa2, 0x69, 0x2b)
	o = msgp.AppendMapHeader(o, uint32(len(z.InPlusEdges)))
	for zxpk, zdnj := range z.InPlusEdges {
		o = msgp.AppendString(o, zxpk)
		o = msgp.AppendArrayHeader(o, uint32(len(zdnj)))
		for zobc := range zdnj {
			o, err = zdnj[zobc].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "o+"
	o = append(o, 0xa2, 0x6f, 0x2b)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutPlusEdges)))
	for zsnv, zkgt := range z.OutPlusEdges {
		o = msgp.AppendString(o, zsnv)
		o = msgp.AppendArrayHeader(o, uint32(len(zkgt)))
		for zema := range zkgt {
			o, err = zkgt[zema].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *EntityExtendedIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zjpj uint32
	zjpj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zjpj > 0 {
		zjpj--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			bts, err = z.PK.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "i+":
			var zzpf uint32
			zzpf, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InPlusEdges == nil && zzpf > 0 {
				z.InPlusEdges = make(map[string][]Key, zzpf)
			} else if len(z.InPlusEdges) > 0 {
				for key, _ := range z.InPlusEdges {
					delete(z.InPlusEdges, key)
				}
			}
			for zzpf > 0 {
				var zxpk string
				var zdnj []Key
				zzpf--
				zxpk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zrfe uint32
				zrfe, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zdnj) >= int(zrfe) {
					zdnj = (zdnj)[:zrfe]
				} else {
					zdnj = make([]Key, zrfe)
				}
				for zobc := range zdnj {
					bts, err = zdnj[zobc].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.InPlusEdges[zxpk] = zdnj
			}
		case "o+":
			var zgmo uint32
			zgmo, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutPlusEdges == nil && zgmo > 0 {
				z.OutPlusEdges = make(map[string][]Key, zgmo)
			} else if len(z.OutPlusEdges) > 0 {
				for key, _ := range z.OutPlusEdges {
					delete(z.OutPlusEdges, key)
				}
			}
			for zgmo > 0 {
				var zsnv string
				var zkgt []Key
				zgmo--
				zsnv, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var ztaf uint32
				ztaf, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zkgt) >= int(ztaf) {
					zkgt = (zkgt)[:ztaf]
				} else {
					zkgt = make([]Key, ztaf)
				}
				for zema := range zkgt {
					bts, err = zkgt[zema].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.OutPlusEdges[zsnv] = zkgt
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
func (z *EntityExtendedIndex) Msgsize() (s int) {
	s = 1 + 2 + z.PK.Msgsize() + 3 + msgp.MapHeaderSize
	if z.InPlusEdges != nil {
		for zxpk, zdnj := range z.InPlusEdges {
			_ = zdnj
			s += msgp.StringPrefixSize + len(zxpk) + msgp.ArrayHeaderSize
			for zobc := range zdnj {
				s += zdnj[zobc].Msgsize()
			}
		}
	}
	s += 3 + msgp.MapHeaderSize
	if z.OutPlusEdges != nil {
		for zsnv, zkgt := range z.OutPlusEdges {
			_ = zkgt
			s += msgp.StringPrefixSize + len(zsnv) + msgp.ArrayHeaderSize
			for zema := range zkgt {
				s += zkgt[zema].Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NamespaceIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zwel uint32
	zwel, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zwel > 0 {
		(*z) = make(NamespaceIndex, zwel)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zwel > 0 {
		zwel--
		var zrjx string
		var zawn string
		zrjx, err = dc.ReadString()
		if err != nil {
			return
		}
		zawn, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[zrjx] = zawn
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z NamespaceIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zrbe, zmfd := range z {
		err = en.WriteString(zrbe)
		if err != nil {
			return
		}
		err = en.WriteString(zmfd)
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
	for zrbe, zmfd := range z {
		o = msgp.AppendString(o, zrbe)
		o = msgp.AppendString(o, zmfd)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NamespaceIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zbal uint32
	zbal, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zbal > 0 {
		(*z) = make(NamespaceIndex, zbal)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zbal > 0 {
		var zzdc string
		var zelx string
		zbal--
		zzdc, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zelx, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[zzdc] = zelx
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z NamespaceIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zjqz, zkct := range z {
			_ = zkct
			s += msgp.StringPrefixSize + len(zjqz) + msgp.StringPrefixSize + len(zkct)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zinl uint32
	zinl, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zinl > 0 {
		(*z) = make(PredIndex, zinl)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zinl > 0 {
		zinl--
		var zana string
		var ztyy *PredicateEntity
		zana, err = dc.ReadString()
		if err != nil {
			return
		}
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			ztyy = nil
		} else {
			if ztyy == nil {
				ztyy = new(PredicateEntity)
			}
			err = ztyy.DecodeMsg(dc)
			if err != nil {
				return
			}
		}
		(*z)[zana] = ztyy
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PredIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zare, zljy := range z {
		err = en.WriteString(zare)
		if err != nil {
			return
		}
		if zljy == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zljy.EncodeMsg(en)
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
	for zare, zljy := range z {
		o = msgp.AppendString(o, zare)
		if zljy == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zljy.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zctn uint32
	zctn, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zctn > 0 {
		(*z) = make(PredIndex, zctn)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zctn > 0 {
		var zixj string
		var zrsc *PredicateEntity
		zctn--
		zixj, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			zrsc = nil
		} else {
			if zrsc == nil {
				zrsc = new(PredicateEntity)
			}
			bts, err = zrsc.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
		(*z)[zixj] = zrsc
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z PredIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zswy, znsg := range z {
			_ = znsg
			s += msgp.StringPrefixSize + len(zswy)
			if znsg == nil {
				s += msgp.NilSize
			} else {
				s += znsg.Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var ztls uint32
	ztls, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for ztls > 0 {
		ztls--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			err = z.PK.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "s":
			var zmvo uint32
			zmvo, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && zmvo > 0 {
				z.Subjects = make(map[string]map[string]uint32, zmvo)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zmvo > 0 {
				zmvo--
				var zrus string
				var zsvm map[string]uint32
				zrus, err = dc.ReadString()
				if err != nil {
					return
				}
				var zigk uint32
				zigk, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if zsvm == nil && zigk > 0 {
					zsvm = make(map[string]uint32, zigk)
				} else if len(zsvm) > 0 {
					for key, _ := range zsvm {
						delete(zsvm, key)
					}
				}
				for zigk > 0 {
					zigk--
					var zaoz string
					var zfzb uint32
					zaoz, err = dc.ReadString()
					if err != nil {
						return
					}
					zfzb, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zsvm[zaoz] = zfzb
				}
				z.Subjects[zrus] = zsvm
			}
		case "o":
			var zopb uint32
			zopb, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && zopb > 0 {
				z.Objects = make(map[string]map[string]uint32, zopb)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zopb > 0 {
				zopb--
				var zsbo string
				var zjif map[string]uint32
				zsbo, err = dc.ReadString()
				if err != nil {
					return
				}
				var zuop uint32
				zuop, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if zjif == nil && zuop > 0 {
					zjif = make(map[string]uint32, zuop)
				} else if len(zjif) > 0 {
					for key, _ := range zjif {
						delete(zjif, key)
					}
				}
				for zuop > 0 {
					zuop--
					var zqgz string
					var zsnw uint32
					zqgz, err = dc.ReadString()
					if err != nil {
						return
					}
					zsnw, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zjif[zqgz] = zsnw
				}
				z.Objects[zsbo] = zjif
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
	err = z.PK.EncodeMsg(en)
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
	for zrus, zsvm := range z.Subjects {
		err = en.WriteString(zrus)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(zsvm)))
		if err != nil {
			return
		}
		for zaoz, zfzb := range zsvm {
			err = en.WriteString(zaoz)
			if err != nil {
				return
			}
			err = en.WriteUint32(zfzb)
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
	for zsbo, zjif := range z.Objects {
		err = en.WriteString(zsbo)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(zjif)))
		if err != nil {
			return
		}
		for zqgz, zsnw := range zjif {
			err = en.WriteString(zqgz)
			if err != nil {
				return
			}
			err = en.WriteUint32(zsnw)
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
	o, err = z.PK.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "s"
	o = append(o, 0xa1, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Subjects)))
	for zrus, zsvm := range z.Subjects {
		o = msgp.AppendString(o, zrus)
		o = msgp.AppendMapHeader(o, uint32(len(zsvm)))
		for zaoz, zfzb := range zsvm {
			o = msgp.AppendString(o, zaoz)
			o = msgp.AppendUint32(o, zfzb)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for zsbo, zjif := range z.Objects {
		o = msgp.AppendString(o, zsbo)
		o = msgp.AppendMapHeader(o, uint32(len(zjif)))
		for zqgz, zsnw := range zjif {
			o = msgp.AppendString(o, zqgz)
			o = msgp.AppendUint32(o, zsnw)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zedl uint32
	zedl, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zedl > 0 {
		zedl--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "p":
			bts, err = z.PK.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "s":
			var zupd uint32
			zupd, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && zupd > 0 {
				z.Subjects = make(map[string]map[string]uint32, zupd)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zupd > 0 {
				var zrus string
				var zsvm map[string]uint32
				zupd--
				zrus, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zome uint32
				zome, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if zsvm == nil && zome > 0 {
					zsvm = make(map[string]uint32, zome)
				} else if len(zsvm) > 0 {
					for key, _ := range zsvm {
						delete(zsvm, key)
					}
				}
				for zome > 0 {
					var zaoz string
					var zfzb uint32
					zome--
					zaoz, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					zfzb, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zsvm[zaoz] = zfzb
				}
				z.Subjects[zrus] = zsvm
			}
		case "o":
			var zrvj uint32
			zrvj, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && zrvj > 0 {
				z.Objects = make(map[string]map[string]uint32, zrvj)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zrvj > 0 {
				var zsbo string
				var zjif map[string]uint32
				zrvj--
				zsbo, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zarz uint32
				zarz, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if zjif == nil && zarz > 0 {
					zjif = make(map[string]uint32, zarz)
				} else if len(zjif) > 0 {
					for key, _ := range zjif {
						delete(zjif, key)
					}
				}
				for zarz > 0 {
					var zqgz string
					var zsnw uint32
					zarz--
					zqgz, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					zsnw, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zjif[zqgz] = zsnw
				}
				z.Objects[zsbo] = zjif
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
func (z *PredicateEntity) Msgsize() (s int) {
	s = 1 + 2 + z.PK.Msgsize() + 2 + msgp.MapHeaderSize
	if z.Subjects != nil {
		for zrus, zsvm := range z.Subjects {
			_ = zsvm
			s += msgp.StringPrefixSize + len(zrus) + msgp.MapHeaderSize
			if zsvm != nil {
				for zaoz, zfzb := range zsvm {
					_ = zfzb
					s += msgp.StringPrefixSize + len(zaoz) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for zsbo, zjif := range z.Objects {
			_ = zjif
			s += msgp.StringPrefixSize + len(zsbo) + msgp.MapHeaderSize
			if zjif != nil {
				for zqgz, zsnw := range zjif {
					_ = zsnw
					s += msgp.StringPrefixSize + len(zqgz) + msgp.Uint32Size
				}
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RelshipIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zbgy uint32
	zbgy, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zbgy > 0 {
		(*z) = make(RelshipIndex, zbgy)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zbgy > 0 {
		zbgy--
		var zucw string
		var zlsx string
		zucw, err = dc.ReadString()
		if err != nil {
			return
		}
		zlsx, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[zucw] = zlsx
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RelshipIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zrao, zmbt := range z {
		err = en.WriteString(zrao)
		if err != nil {
			return
		}
		err = en.WriteString(zmbt)
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
	for zrao, zmbt := range z {
		o = msgp.AppendString(o, zrao)
		o = msgp.AppendString(o, zmbt)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RelshipIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zzak uint32
	zzak, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zzak > 0 {
		(*z) = make(RelshipIndex, zzak)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zzak > 0 {
		var zvls string
		var zjfj string
		zzak--
		zvls, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zjfj, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[zvls] = zjfj
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RelshipIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zbtz, zsym := range z {
			_ = zsym
			s += msgp.StringPrefixSize + len(zbtz) + msgp.StringPrefixSize + len(zsym)
		}
	}
	return
}
