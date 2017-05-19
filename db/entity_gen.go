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
	var zjfb uint32
	zjfb, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zjfb > 0 {
		zjfb--
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
			var zcxo uint32
			zcxo, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && zcxo > 0 {
				z.InEdges = make(map[string][]Key, zcxo)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zcxo > 0 {
				zcxo--
				var zxvk string
				var zbzg []Key
				zxvk, err = dc.ReadString()
				if err != nil {
					return
				}
				var zeff uint32
				zeff, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zbzg) >= int(zeff) {
					zbzg = (zbzg)[:zeff]
				} else {
					zbzg = make([]Key, zeff)
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
			var zrsw uint32
			zrsw, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && zrsw > 0 {
				z.OutEdges = make(map[string][]Key, zrsw)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zrsw > 0 {
				zrsw--
				var zcmr string
				var zajw []Key
				zcmr, err = dc.ReadString()
				if err != nil {
					return
				}
				var zxpk uint32
				zxpk, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zajw) >= int(zxpk) {
					zajw = (zajw)[:zxpk]
				} else {
					zajw = make([]Key, zxpk)
				}
				for zwht := range zajw {
					err = zajw[zwht].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.OutEdges[zcmr] = zajw
			}
		case "i+":
			var zdnj uint32
			zdnj, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InPlusEdges == nil && zdnj > 0 {
				z.InPlusEdges = make(map[string][]Key, zdnj)
			} else if len(z.InPlusEdges) > 0 {
				for key, _ := range z.InPlusEdges {
					delete(z.InPlusEdges, key)
				}
			}
			for zdnj > 0 {
				zdnj--
				var zhct string
				var zcua []Key
				zhct, err = dc.ReadString()
				if err != nil {
					return
				}
				var zobc uint32
				zobc, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zcua) >= int(zobc) {
					zcua = (zcua)[:zobc]
				} else {
					zcua = make([]Key, zobc)
				}
				for zxhx := range zcua {
					err = zcua[zxhx].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.InPlusEdges[zhct] = zcua
			}
		case "o+":
			var zsnv uint32
			zsnv, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutPlusEdges == nil && zsnv > 0 {
				z.OutPlusEdges = make(map[string][]Key, zsnv)
			} else if len(z.OutPlusEdges) > 0 {
				for key, _ := range z.OutPlusEdges {
					delete(z.OutPlusEdges, key)
				}
			}
			for zsnv > 0 {
				zsnv--
				var zlqf string
				var zdaf []Key
				zlqf, err = dc.ReadString()
				if err != nil {
					return
				}
				var zkgt uint32
				zkgt, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zdaf) >= int(zkgt) {
					zdaf = (zdaf)[:zkgt]
				} else {
					zdaf = make([]Key, zkgt)
				}
				for zpks := range zdaf {
					err = zdaf[zpks].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.OutPlusEdges[zlqf] = zdaf
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
	// map header, size 5
	// write "p"
	err = en.Append(0x85, 0xa1, 0x70)
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
	// write "i+"
	err = en.Append(0xa2, 0x69, 0x2b)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InPlusEdges)))
	if err != nil {
		return
	}
	for zhct, zcua := range z.InPlusEdges {
		err = en.WriteString(zhct)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zcua)))
		if err != nil {
			return
		}
		for zxhx := range zcua {
			err = zcua[zxhx].EncodeMsg(en)
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
	for zlqf, zdaf := range z.OutPlusEdges {
		err = en.WriteString(zlqf)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zdaf)))
		if err != nil {
			return
		}
		for zpks := range zdaf {
			err = zdaf[zpks].EncodeMsg(en)
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
	// map header, size 5
	// string "p"
	o = append(o, 0x85, 0xa1, 0x70)
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
	// string "i+"
	o = append(o, 0xa2, 0x69, 0x2b)
	o = msgp.AppendMapHeader(o, uint32(len(z.InPlusEdges)))
	for zhct, zcua := range z.InPlusEdges {
		o = msgp.AppendString(o, zhct)
		o = msgp.AppendArrayHeader(o, uint32(len(zcua)))
		for zxhx := range zcua {
			o, err = zcua[zxhx].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "o+"
	o = append(o, 0xa2, 0x6f, 0x2b)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutPlusEdges)))
	for zlqf, zdaf := range z.OutPlusEdges {
		o = msgp.AppendString(o, zlqf)
		o = msgp.AppendArrayHeader(o, uint32(len(zdaf)))
		for zpks := range zdaf {
			o, err = zdaf[zpks].MarshalMsg(o)
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
	var zema uint32
	zema, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zema > 0 {
		zema--
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
			var zpez uint32
			zpez, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && zpez > 0 {
				z.InEdges = make(map[string][]Key, zpez)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zpez > 0 {
				var zxvk string
				var zbzg []Key
				zpez--
				zxvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zqke uint32
				zqke, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zbzg) >= int(zqke) {
					zbzg = (zbzg)[:zqke]
				} else {
					zbzg = make([]Key, zqke)
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
			var zqyh uint32
			zqyh, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && zqyh > 0 {
				z.OutEdges = make(map[string][]Key, zqyh)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zqyh > 0 {
				var zcmr string
				var zajw []Key
				zqyh--
				zcmr, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zyzr uint32
				zyzr, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zajw) >= int(zyzr) {
					zajw = (zajw)[:zyzr]
				} else {
					zajw = make([]Key, zyzr)
				}
				for zwht := range zajw {
					bts, err = zajw[zwht].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.OutEdges[zcmr] = zajw
			}
		case "i+":
			var zywj uint32
			zywj, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InPlusEdges == nil && zywj > 0 {
				z.InPlusEdges = make(map[string][]Key, zywj)
			} else if len(z.InPlusEdges) > 0 {
				for key, _ := range z.InPlusEdges {
					delete(z.InPlusEdges, key)
				}
			}
			for zywj > 0 {
				var zhct string
				var zcua []Key
				zywj--
				zhct, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zjpj uint32
				zjpj, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zcua) >= int(zjpj) {
					zcua = (zcua)[:zjpj]
				} else {
					zcua = make([]Key, zjpj)
				}
				for zxhx := range zcua {
					bts, err = zcua[zxhx].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.InPlusEdges[zhct] = zcua
			}
		case "o+":
			var zzpf uint32
			zzpf, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutPlusEdges == nil && zzpf > 0 {
				z.OutPlusEdges = make(map[string][]Key, zzpf)
			} else if len(z.OutPlusEdges) > 0 {
				for key, _ := range z.OutPlusEdges {
					delete(z.OutPlusEdges, key)
				}
			}
			for zzpf > 0 {
				var zlqf string
				var zdaf []Key
				zzpf--
				zlqf, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zrfe uint32
				zrfe, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zdaf) >= int(zrfe) {
					zdaf = (zdaf)[:zrfe]
				} else {
					zdaf = make([]Key, zrfe)
				}
				for zpks := range zdaf {
					bts, err = zdaf[zpks].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.OutPlusEdges[zlqf] = zdaf
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
	s += 3 + msgp.MapHeaderSize
	if z.InPlusEdges != nil {
		for zhct, zcua := range z.InPlusEdges {
			_ = zcua
			s += msgp.StringPrefixSize + len(zhct) + msgp.ArrayHeaderSize
			for zxhx := range zcua {
				s += zcua[zxhx].Msgsize()
			}
		}
	}
	s += 3 + msgp.MapHeaderSize
	if z.OutPlusEdges != nil {
		for zlqf, zdaf := range z.OutPlusEdges {
			_ = zdaf
			s += msgp.StringPrefixSize + len(zlqf) + msgp.ArrayHeaderSize
			for zpks := range zdaf {
				s += zdaf[zpks].Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NamespaceIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zrjx uint32
	zrjx, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zrjx > 0 {
		(*z) = make(NamespaceIndex, zrjx)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zrjx > 0 {
		zrjx--
		var zeth string
		var zsbz string
		zeth, err = dc.ReadString()
		if err != nil {
			return
		}
		zsbz, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[zeth] = zsbz
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z NamespaceIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zawn, zwel := range z {
		err = en.WriteString(zawn)
		if err != nil {
			return
		}
		err = en.WriteString(zwel)
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
	for zawn, zwel := range z {
		o = msgp.AppendString(o, zawn)
		o = msgp.AppendString(o, zwel)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NamespaceIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zzdc uint32
	zzdc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zzdc > 0 {
		(*z) = make(NamespaceIndex, zzdc)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zzdc > 0 {
		var zrbe string
		var zmfd string
		zzdc--
		zrbe, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zmfd, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[zrbe] = zmfd
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z NamespaceIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zelx, zbal := range z {
			_ = zbal
			s += msgp.StringPrefixSize + len(zelx) + msgp.StringPrefixSize + len(zbal)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zana uint32
	zana, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zana > 0 {
		(*z) = make(PredIndex, zana)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zana > 0 {
		zana--
		var ztmt string
		var ztco *PredicateEntity
		ztmt, err = dc.ReadString()
		if err != nil {
			return
		}
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			ztco = nil
		} else {
			if ztco == nil {
				ztco = new(PredicateEntity)
			}
			err = ztco.DecodeMsg(dc)
			if err != nil {
				return
			}
		}
		(*z)[ztmt] = ztco
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PredIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for ztyy, zinl := range z {
		err = en.WriteString(ztyy)
		if err != nil {
			return
		}
		if zinl == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zinl.EncodeMsg(en)
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
	for ztyy, zinl := range z {
		o = msgp.AppendString(o, ztyy)
		if zinl == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zinl.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zixj uint32
	zixj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zixj > 0 {
		(*z) = make(PredIndex, zixj)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zixj > 0 {
		var zare string
		var zljy *PredicateEntity
		zixj--
		zare, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			zljy = nil
		} else {
			if zljy == nil {
				zljy = new(PredicateEntity)
			}
			bts, err = zljy.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
		(*z)[zare] = zljy
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z PredIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zrsc, zctn := range z {
			_ = zctn
			s += msgp.StringPrefixSize + len(zrsc)
			if zctn == nil {
				s += msgp.NilSize
			} else {
				s += zctn.Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zqgz uint32
	zqgz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zqgz > 0 {
		zqgz--
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
			var zsnw uint32
			zsnw, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && zsnw > 0 {
				z.Subjects = make(map[string]map[string]uint32, zsnw)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zsnw > 0 {
				zsnw--
				var zswy string
				var znsg map[string]uint32
				zswy, err = dc.ReadString()
				if err != nil {
					return
				}
				var ztls uint32
				ztls, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if znsg == nil && ztls > 0 {
					znsg = make(map[string]uint32, ztls)
				} else if len(znsg) > 0 {
					for key, _ := range znsg {
						delete(znsg, key)
					}
				}
				for ztls > 0 {
					ztls--
					var zrus string
					var zsvm uint32
					zrus, err = dc.ReadString()
					if err != nil {
						return
					}
					zsvm, err = dc.ReadUint32()
					if err != nil {
						return
					}
					znsg[zrus] = zsvm
				}
				z.Subjects[zswy] = znsg
			}
		case "o":
			var zmvo uint32
			zmvo, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && zmvo > 0 {
				z.Objects = make(map[string]map[string]uint32, zmvo)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zmvo > 0 {
				zmvo--
				var zaoz string
				var zfzb map[string]uint32
				zaoz, err = dc.ReadString()
				if err != nil {
					return
				}
				var zigk uint32
				zigk, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if zfzb == nil && zigk > 0 {
					zfzb = make(map[string]uint32, zigk)
				} else if len(zfzb) > 0 {
					for key, _ := range zfzb {
						delete(zfzb, key)
					}
				}
				for zigk > 0 {
					zigk--
					var zsbo string
					var zjif uint32
					zsbo, err = dc.ReadString()
					if err != nil {
						return
					}
					zjif, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zfzb[zsbo] = zjif
				}
				z.Objects[zaoz] = zfzb
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
	for zswy, znsg := range z.Subjects {
		err = en.WriteString(zswy)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(znsg)))
		if err != nil {
			return
		}
		for zrus, zsvm := range znsg {
			err = en.WriteString(zrus)
			if err != nil {
				return
			}
			err = en.WriteUint32(zsvm)
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
	for zaoz, zfzb := range z.Objects {
		err = en.WriteString(zaoz)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(zfzb)))
		if err != nil {
			return
		}
		for zsbo, zjif := range zfzb {
			err = en.WriteString(zsbo)
			if err != nil {
				return
			}
			err = en.WriteUint32(zjif)
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
	for zswy, znsg := range z.Subjects {
		o = msgp.AppendString(o, zswy)
		o = msgp.AppendMapHeader(o, uint32(len(znsg)))
		for zrus, zsvm := range znsg {
			o = msgp.AppendString(o, zrus)
			o = msgp.AppendUint32(o, zsvm)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for zaoz, zfzb := range z.Objects {
		o = msgp.AppendString(o, zaoz)
		o = msgp.AppendMapHeader(o, uint32(len(zfzb)))
		for zsbo, zjif := range zfzb {
			o = msgp.AppendString(o, zsbo)
			o = msgp.AppendUint32(o, zjif)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zopb uint32
	zopb, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zopb > 0 {
		zopb--
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
			var zuop uint32
			zuop, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && zuop > 0 {
				z.Subjects = make(map[string]map[string]uint32, zuop)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zuop > 0 {
				var zswy string
				var znsg map[string]uint32
				zuop--
				zswy, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zedl uint32
				zedl, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if znsg == nil && zedl > 0 {
					znsg = make(map[string]uint32, zedl)
				} else if len(znsg) > 0 {
					for key, _ := range znsg {
						delete(znsg, key)
					}
				}
				for zedl > 0 {
					var zrus string
					var zsvm uint32
					zedl--
					zrus, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					zsvm, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					znsg[zrus] = zsvm
				}
				z.Subjects[zswy] = znsg
			}
		case "o":
			var zupd uint32
			zupd, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && zupd > 0 {
				z.Objects = make(map[string]map[string]uint32, zupd)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zupd > 0 {
				var zaoz string
				var zfzb map[string]uint32
				zupd--
				zaoz, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zome uint32
				zome, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if zfzb == nil && zome > 0 {
					zfzb = make(map[string]uint32, zome)
				} else if len(zfzb) > 0 {
					for key, _ := range zfzb {
						delete(zfzb, key)
					}
				}
				for zome > 0 {
					var zsbo string
					var zjif uint32
					zome--
					zsbo, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					zjif, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zfzb[zsbo] = zjif
				}
				z.Objects[zaoz] = zfzb
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
		for zswy, znsg := range z.Subjects {
			_ = znsg
			s += msgp.StringPrefixSize + len(zswy) + msgp.MapHeaderSize
			if znsg != nil {
				for zrus, zsvm := range znsg {
					_ = zsvm
					s += msgp.StringPrefixSize + len(zrus) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for zaoz, zfzb := range z.Objects {
			_ = zfzb
			s += msgp.StringPrefixSize + len(zaoz) + msgp.MapHeaderSize
			if zfzb != nil {
				for zsbo, zjif := range zfzb {
					_ = zjif
					s += msgp.StringPrefixSize + len(zsbo) + msgp.Uint32Size
				}
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RelshipIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zucw uint32
	zucw, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zucw > 0 {
		(*z) = make(RelshipIndex, zucw)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zucw > 0 {
		zucw--
		var zknt string
		var zxye string
		zknt, err = dc.ReadString()
		if err != nil {
			return
		}
		zxye, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[zknt] = zxye
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RelshipIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zlsx, zbgy := range z {
		err = en.WriteString(zlsx)
		if err != nil {
			return
		}
		err = en.WriteString(zbgy)
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
	for zlsx, zbgy := range z {
		o = msgp.AppendString(o, zlsx)
		o = msgp.AppendString(o, zbgy)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RelshipIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zvls uint32
	zvls, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zvls > 0 {
		(*z) = make(RelshipIndex, zvls)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zvls > 0 {
		var zrao string
		var zmbt string
		zvls--
		zrao, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zmbt, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[zrao] = zmbt
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RelshipIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zjfj, zzak := range z {
			_ = zzak
			s += msgp.StringPrefixSize + len(zjfj) + msgp.StringPrefixSize + len(zzak)
		}
	}
	return
}
