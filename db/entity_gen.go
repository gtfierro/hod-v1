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
		case "ein":
			var zcua uint32
			zcua, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && zcua > 0 {
				z.InEdges = make(map[uint32][]Key, zcua)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zcua > 0 {
				zcua--
				var zxvk uint32
				var zbzg []Key
				zxvk, err = dc.ReadUint32()
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
		case "eout":
			var zlqf uint32
			zlqf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && zlqf > 0 {
				z.OutEdges = make(map[uint32][]Key, zlqf)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zlqf > 0 {
				zlqf--
				var zcmr uint32
				var zajw []Key
				zcmr, err = dc.ReadUint32()
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
	// write "ein"
	err = en.Append(0xa3, 0x65, 0x69, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InEdges)))
	if err != nil {
		return
	}
	for zxvk, zbzg := range z.InEdges {
		err = en.WriteUint32(zxvk)
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
	// write "eout"
	err = en.Append(0xa4, 0x65, 0x6f, 0x75, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.OutEdges)))
	if err != nil {
		return
	}
	for zcmr, zajw := range z.OutEdges {
		err = en.WriteUint32(zcmr)
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
	// string "ein"
	o = append(o, 0xa3, 0x65, 0x69, 0x6e)
	o = msgp.AppendMapHeader(o, uint32(len(z.InEdges)))
	for zxvk, zbzg := range z.InEdges {
		o = msgp.AppendUint32(o, zxvk)
		o = msgp.AppendArrayHeader(o, uint32(len(zbzg)))
		for zbai := range zbzg {
			o, err = zbzg[zbai].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "eout"
	o = append(o, 0xa4, 0x65, 0x6f, 0x75, 0x74)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutEdges)))
	for zcmr, zajw := range z.OutEdges {
		o = msgp.AppendUint32(o, zcmr)
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
		case "ein":
			var zjfb uint32
			zjfb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && zjfb > 0 {
				z.InEdges = make(map[uint32][]Key, zjfb)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zjfb > 0 {
				var zxvk uint32
				var zbzg []Key
				zjfb--
				zxvk, bts, err = msgp.ReadUint32Bytes(bts)
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
		case "eout":
			var zeff uint32
			zeff, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && zeff > 0 {
				z.OutEdges = make(map[uint32][]Key, zeff)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zeff > 0 {
				var zcmr uint32
				var zajw []Key
				zeff--
				zcmr, bts, err = msgp.ReadUint32Bytes(bts)
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
	s = 1 + 2 + z.PK.Msgsize() + 4 + msgp.MapHeaderSize
	if z.InEdges != nil {
		for _, zbzg := range z.InEdges {
			_ = zbzg
			s += msgp.Uint32Size + msgp.ArrayHeaderSize
			for zbai := range zbzg {
				s += zbzg[zbai].Msgsize()
			}
		}
	}
	s += 5 + msgp.MapHeaderSize
	if z.OutEdges != nil {
		for _, zajw := range z.OutEdges {
			_ = zajw
			s += msgp.Uint32Size + msgp.ArrayHeaderSize
			for zwht := range zajw {
				s += zajw[zwht].Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zqyh uint32
	zqyh, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zqyh > 0 {
		zqyh--
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
			var zyzr uint32
			zyzr, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && zyzr > 0 {
				z.Subjects = make(map[uint32]map[uint32]uint32, zyzr)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zyzr > 0 {
				zyzr--
				var zxpk uint32
				var zdnj map[uint32]uint32
				zxpk, err = dc.ReadUint32()
				if err != nil {
					return
				}
				var zywj uint32
				zywj, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if zdnj == nil && zywj > 0 {
					zdnj = make(map[uint32]uint32, zywj)
				} else if len(zdnj) > 0 {
					for key, _ := range zdnj {
						delete(zdnj, key)
					}
				}
				for zywj > 0 {
					zywj--
					var zobc uint32
					var zsnv uint32
					zobc, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zsnv, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zdnj[zobc] = zsnv
				}
				z.Subjects[zxpk] = zdnj
			}
		case "o":
			var zjpj uint32
			zjpj, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && zjpj > 0 {
				z.Objects = make(map[uint32]map[uint32]uint32, zjpj)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zjpj > 0 {
				zjpj--
				var zkgt uint32
				var zema map[uint32]uint32
				zkgt, err = dc.ReadUint32()
				if err != nil {
					return
				}
				var zzpf uint32
				zzpf, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if zema == nil && zzpf > 0 {
					zema = make(map[uint32]uint32, zzpf)
				} else if len(zema) > 0 {
					for key, _ := range zema {
						delete(zema, key)
					}
				}
				for zzpf > 0 {
					zzpf--
					var zpez uint32
					var zqke uint32
					zpez, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zqke, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zema[zpez] = zqke
				}
				z.Objects[zkgt] = zema
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
	for zxpk, zdnj := range z.Subjects {
		err = en.WriteUint32(zxpk)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(zdnj)))
		if err != nil {
			return
		}
		for zobc, zsnv := range zdnj {
			err = en.WriteUint32(zobc)
			if err != nil {
				return
			}
			err = en.WriteUint32(zsnv)
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
	for zkgt, zema := range z.Objects {
		err = en.WriteUint32(zkgt)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(zema)))
		if err != nil {
			return
		}
		for zpez, zqke := range zema {
			err = en.WriteUint32(zpez)
			if err != nil {
				return
			}
			err = en.WriteUint32(zqke)
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
	for zxpk, zdnj := range z.Subjects {
		o = msgp.AppendUint32(o, zxpk)
		o = msgp.AppendMapHeader(o, uint32(len(zdnj)))
		for zobc, zsnv := range zdnj {
			o = msgp.AppendUint32(o, zobc)
			o = msgp.AppendUint32(o, zsnv)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for zkgt, zema := range z.Objects {
		o = msgp.AppendUint32(o, zkgt)
		o = msgp.AppendMapHeader(o, uint32(len(zema)))
		for zpez, zqke := range zema {
			o = msgp.AppendUint32(o, zpez)
			o = msgp.AppendUint32(o, zqke)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zrfe uint32
	zrfe, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zrfe > 0 {
		zrfe--
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
			var zgmo uint32
			zgmo, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && zgmo > 0 {
				z.Subjects = make(map[uint32]map[uint32]uint32, zgmo)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zgmo > 0 {
				var zxpk uint32
				var zdnj map[uint32]uint32
				zgmo--
				zxpk, bts, err = msgp.ReadUint32Bytes(bts)
				if err != nil {
					return
				}
				var ztaf uint32
				ztaf, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if zdnj == nil && ztaf > 0 {
					zdnj = make(map[uint32]uint32, ztaf)
				} else if len(zdnj) > 0 {
					for key, _ := range zdnj {
						delete(zdnj, key)
					}
				}
				for ztaf > 0 {
					var zobc uint32
					var zsnv uint32
					ztaf--
					zobc, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zsnv, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zdnj[zobc] = zsnv
				}
				z.Subjects[zxpk] = zdnj
			}
		case "o":
			var zeth uint32
			zeth, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && zeth > 0 {
				z.Objects = make(map[uint32]map[uint32]uint32, zeth)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zeth > 0 {
				var zkgt uint32
				var zema map[uint32]uint32
				zeth--
				zkgt, bts, err = msgp.ReadUint32Bytes(bts)
				if err != nil {
					return
				}
				var zsbz uint32
				zsbz, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if zema == nil && zsbz > 0 {
					zema = make(map[uint32]uint32, zsbz)
				} else if len(zema) > 0 {
					for key, _ := range zema {
						delete(zema, key)
					}
				}
				for zsbz > 0 {
					var zpez uint32
					var zqke uint32
					zsbz--
					zpez, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zqke, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zema[zpez] = zqke
				}
				z.Objects[zkgt] = zema
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
		for _, zdnj := range z.Subjects {
			_ = zdnj
			s += msgp.Uint32Size + msgp.MapHeaderSize
			if zdnj != nil {
				for _, zsnv := range zdnj {
					_ = zsnv
					s += msgp.Uint32Size + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for _, zema := range z.Objects {
			_ = zema
			s += msgp.Uint32Size + msgp.MapHeaderSize
			if zema != nil {
				for _, zqke := range zema {
					_ = zqke
					s += msgp.Uint32Size + msgp.Uint32Size
				}
			}
		}
	}
	return
}
