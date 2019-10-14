package storage

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *BytesEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zlqf uint32
	zlqf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zlqf > 0 {
		zlqf--
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
		case "i":
			var zdaf uint32
			zdaf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && zdaf > 0 {
				z.InEdges = make(map[string][]HashKey, zdaf)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zdaf > 0 {
				zdaf--
				var zbzg string
				var zbai []HashKey
				zbzg, err = dc.ReadString()
				if err != nil {
					return
				}
				var zpks uint32
				zpks, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zbai) >= int(zpks) {
					zbai = (zbai)[:zpks]
				} else {
					zbai = make([]HashKey, zpks)
				}
				for zcmr := range zbai {
					err = dc.ReadExactBytes(zbai[zcmr][:])
					if err != nil {
						return
					}
				}
				z.InEdges[zbzg] = zbai
			}
		case "o":
			var zjfb uint32
			zjfb, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && zjfb > 0 {
				z.OutEdges = make(map[string][]HashKey, zjfb)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zjfb > 0 {
				zjfb--
				var zwht string
				var zhct []HashKey
				zwht, err = dc.ReadString()
				if err != nil {
					return
				}
				var zcxo uint32
				zcxo, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zhct) >= int(zcxo) {
					zhct = (zhct)[:zcxo]
				} else {
					zhct = make([]HashKey, zcxo)
				}
				for zcua := range zhct {
					err = dc.ReadExactBytes(zhct[zcua][:])
					if err != nil {
						return
					}
				}
				z.OutEdges[zwht] = zhct
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
func (z *BytesEntity) EncodeMsg(en *msgp.Writer) (err error) {
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
	// write "i"
	err = en.Append(0xa1, 0x69)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InEdges)))
	if err != nil {
		return
	}
	for zbzg, zbai := range z.InEdges {
		err = en.WriteString(zbzg)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zbai)))
		if err != nil {
			return
		}
		for zcmr := range zbai {
			err = en.WriteBytes(zbai[zcmr][:])
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
	for zwht, zhct := range z.OutEdges {
		err = en.WriteString(zwht)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zhct)))
		if err != nil {
			return
		}
		for zcua := range zhct {
			err = en.WriteBytes(zhct[zcua][:])
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BytesEntity) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "p"
	o = append(o, 0x83, 0xa1, 0x70)
	o = msgp.AppendBytes(o, z.PK[:])
	// string "i"
	o = append(o, 0xa1, 0x69)
	o = msgp.AppendMapHeader(o, uint32(len(z.InEdges)))
	for zbzg, zbai := range z.InEdges {
		o = msgp.AppendString(o, zbzg)
		o = msgp.AppendArrayHeader(o, uint32(len(zbai)))
		for zcmr := range zbai {
			o = msgp.AppendBytes(o, zbai[zcmr][:])
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutEdges)))
	for zwht, zhct := range z.OutEdges {
		o = msgp.AppendString(o, zwht)
		o = msgp.AppendArrayHeader(o, uint32(len(zhct)))
		for zcua := range zhct {
			o = msgp.AppendBytes(o, zhct[zcua][:])
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BytesEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zeff uint32
	zeff, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zeff > 0 {
		zeff--
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
		case "i":
			var zrsw uint32
			zrsw, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && zrsw > 0 {
				z.InEdges = make(map[string][]HashKey, zrsw)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for zrsw > 0 {
				var zbzg string
				var zbai []HashKey
				zrsw--
				zbzg, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zxpk uint32
				zxpk, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zbai) >= int(zxpk) {
					zbai = (zbai)[:zxpk]
				} else {
					zbai = make([]HashKey, zxpk)
				}
				for zcmr := range zbai {
					bts, err = msgp.ReadExactBytes(bts, zbai[zcmr][:])
					if err != nil {
						return
					}
				}
				z.InEdges[zbzg] = zbai
			}
		case "o":
			var zdnj uint32
			zdnj, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && zdnj > 0 {
				z.OutEdges = make(map[string][]HashKey, zdnj)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for zdnj > 0 {
				var zwht string
				var zhct []HashKey
				zdnj--
				zwht, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zobc uint32
				zobc, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zhct) >= int(zobc) {
					zhct = (zhct)[:zobc]
				} else {
					zhct = make([]HashKey, zobc)
				}
				for zcua := range zhct {
					bts, err = msgp.ReadExactBytes(bts, zhct[zcua][:])
					if err != nil {
						return
					}
				}
				z.OutEdges[zwht] = zhct
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
func (z *BytesEntity) Msgsize() (s int) {
	s = 1 + 2 + msgp.ArrayHeaderSize + (8 * (msgp.ByteSize)) + 2 + msgp.MapHeaderSize
	if z.InEdges != nil {
		for zbzg, zbai := range z.InEdges {
			_ = zbai
			s += msgp.StringPrefixSize + len(zbzg) + msgp.ArrayHeaderSize + (len(zbai) * (8 * (msgp.ByteSize)))
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.OutEdges != nil {
		for zwht, zhct := range z.OutEdges {
			_ = zhct
			s += msgp.StringPrefixSize + len(zwht) + msgp.ArrayHeaderSize + (len(zhct) * (8 * (msgp.ByteSize)))
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BytesEntityExtendedIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zzpf uint32
	zzpf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zzpf > 0 {
		zzpf--
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
		case "i+":
			var zrfe uint32
			zrfe, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InPlusEdges == nil && zrfe > 0 {
				z.InPlusEdges = make(map[string][]HashKey, zrfe)
			} else if len(z.InPlusEdges) > 0 {
				for key, _ := range z.InPlusEdges {
					delete(z.InPlusEdges, key)
				}
			}
			for zrfe > 0 {
				zrfe--
				var zkgt string
				var zema []HashKey
				zkgt, err = dc.ReadString()
				if err != nil {
					return
				}
				var zgmo uint32
				zgmo, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zema) >= int(zgmo) {
					zema = (zema)[:zgmo]
				} else {
					zema = make([]HashKey, zgmo)
				}
				for zpez := range zema {
					err = dc.ReadExactBytes(zema[zpez][:])
					if err != nil {
						return
					}
				}
				z.InPlusEdges[zkgt] = zema
			}
		case "o+":
			var ztaf uint32
			ztaf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutPlusEdges == nil && ztaf > 0 {
				z.OutPlusEdges = make(map[string][]HashKey, ztaf)
			} else if len(z.OutPlusEdges) > 0 {
				for key, _ := range z.OutPlusEdges {
					delete(z.OutPlusEdges, key)
				}
			}
			for ztaf > 0 {
				ztaf--
				var zqyh string
				var zyzr []HashKey
				zqyh, err = dc.ReadString()
				if err != nil {
					return
				}
				var zeth uint32
				zeth, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(zyzr) >= int(zeth) {
					zyzr = (zyzr)[:zeth]
				} else {
					zyzr = make([]HashKey, zeth)
				}
				for zywj := range zyzr {
					err = dc.ReadExactBytes(zyzr[zywj][:])
					if err != nil {
						return
					}
				}
				z.OutPlusEdges[zqyh] = zyzr
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
func (z *BytesEntityExtendedIndex) EncodeMsg(en *msgp.Writer) (err error) {
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
	// write "i+"
	err = en.Append(0xa2, 0x69, 0x2b)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.InPlusEdges)))
	if err != nil {
		return
	}
	for zkgt, zema := range z.InPlusEdges {
		err = en.WriteString(zkgt)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zema)))
		if err != nil {
			return
		}
		for zpez := range zema {
			err = en.WriteBytes(zema[zpez][:])
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
	for zqyh, zyzr := range z.OutPlusEdges {
		err = en.WriteString(zqyh)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(zyzr)))
		if err != nil {
			return
		}
		for zywj := range zyzr {
			err = en.WriteBytes(zyzr[zywj][:])
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BytesEntityExtendedIndex) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "p"
	o = append(o, 0x83, 0xa1, 0x70)
	o = msgp.AppendBytes(o, z.PK[:])
	// string "i+"
	o = append(o, 0xa2, 0x69, 0x2b)
	o = msgp.AppendMapHeader(o, uint32(len(z.InPlusEdges)))
	for zkgt, zema := range z.InPlusEdges {
		o = msgp.AppendString(o, zkgt)
		o = msgp.AppendArrayHeader(o, uint32(len(zema)))
		for zpez := range zema {
			o = msgp.AppendBytes(o, zema[zpez][:])
		}
	}
	// string "o+"
	o = append(o, 0xa2, 0x6f, 0x2b)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutPlusEdges)))
	for zqyh, zyzr := range z.OutPlusEdges {
		o = msgp.AppendString(o, zqyh)
		o = msgp.AppendArrayHeader(o, uint32(len(zyzr)))
		for zywj := range zyzr {
			o = msgp.AppendBytes(o, zyzr[zywj][:])
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BytesEntityExtendedIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zsbz uint32
	zsbz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zsbz > 0 {
		zsbz--
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
		case "i+":
			var zrjx uint32
			zrjx, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InPlusEdges == nil && zrjx > 0 {
				z.InPlusEdges = make(map[string][]HashKey, zrjx)
			} else if len(z.InPlusEdges) > 0 {
				for key, _ := range z.InPlusEdges {
					delete(z.InPlusEdges, key)
				}
			}
			for zrjx > 0 {
				var zkgt string
				var zema []HashKey
				zrjx--
				zkgt, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zawn uint32
				zawn, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zema) >= int(zawn) {
					zema = (zema)[:zawn]
				} else {
					zema = make([]HashKey, zawn)
				}
				for zpez := range zema {
					bts, err = msgp.ReadExactBytes(bts, zema[zpez][:])
					if err != nil {
						return
					}
				}
				z.InPlusEdges[zkgt] = zema
			}
		case "o+":
			var zwel uint32
			zwel, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutPlusEdges == nil && zwel > 0 {
				z.OutPlusEdges = make(map[string][]HashKey, zwel)
			} else if len(z.OutPlusEdges) > 0 {
				for key, _ := range z.OutPlusEdges {
					delete(z.OutPlusEdges, key)
				}
			}
			for zwel > 0 {
				var zqyh string
				var zyzr []HashKey
				zwel--
				zqyh, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zrbe uint32
				zrbe, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(zyzr) >= int(zrbe) {
					zyzr = (zyzr)[:zrbe]
				} else {
					zyzr = make([]HashKey, zrbe)
				}
				for zywj := range zyzr {
					bts, err = msgp.ReadExactBytes(bts, zyzr[zywj][:])
					if err != nil {
						return
					}
				}
				z.OutPlusEdges[zqyh] = zyzr
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
func (z *BytesEntityExtendedIndex) Msgsize() (s int) {
	s = 1 + 2 + msgp.ArrayHeaderSize + (8 * (msgp.ByteSize)) + 3 + msgp.MapHeaderSize
	if z.InPlusEdges != nil {
		for zkgt, zema := range z.InPlusEdges {
			_ = zema
			s += msgp.StringPrefixSize + len(zkgt) + msgp.ArrayHeaderSize + (len(zema) * (8 * (msgp.ByteSize)))
		}
	}
	s += 3 + msgp.MapHeaderSize
	if z.OutPlusEdges != nil {
		for zqyh, zyzr := range z.OutPlusEdges {
			_ = zyzr
			s += msgp.StringPrefixSize + len(zqyh) + msgp.ArrayHeaderSize + (len(zyzr) * (8 * (msgp.ByteSize)))
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BytesPredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var ztyy uint32
	ztyy, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for ztyy > 0 {
		ztyy--
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
			var zinl uint32
			zinl, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && zinl > 0 {
				z.Subjects = make(map[string]map[string]uint32, zinl)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zinl > 0 {
				zinl--
				var zzdc string
				var zelx map[string]uint32
				zzdc, err = dc.ReadString()
				if err != nil {
					return
				}
				var zare uint32
				zare, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if zelx == nil && zare > 0 {
					zelx = make(map[string]uint32, zare)
				} else if len(zelx) > 0 {
					for key, _ := range zelx {
						delete(zelx, key)
					}
				}
				for zare > 0 {
					zare--
					var zbal string
					var zjqz uint32
					zbal, err = dc.ReadString()
					if err != nil {
						return
					}
					zjqz, err = dc.ReadUint32()
					if err != nil {
						return
					}
					zelx[zbal] = zjqz
				}
				z.Subjects[zzdc] = zelx
			}
		case "o":
			var zljy uint32
			zljy, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && zljy > 0 {
				z.Objects = make(map[string]map[string]uint32, zljy)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for zljy > 0 {
				zljy--
				var zkct string
				var ztmt map[string]uint32
				zkct, err = dc.ReadString()
				if err != nil {
					return
				}
				var zixj uint32
				zixj, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if ztmt == nil && zixj > 0 {
					ztmt = make(map[string]uint32, zixj)
				} else if len(ztmt) > 0 {
					for key, _ := range ztmt {
						delete(ztmt, key)
					}
				}
				for zixj > 0 {
					zixj--
					var ztco string
					var zana uint32
					ztco, err = dc.ReadString()
					if err != nil {
						return
					}
					zana, err = dc.ReadUint32()
					if err != nil {
						return
					}
					ztmt[ztco] = zana
				}
				z.Objects[zkct] = ztmt
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
func (z *BytesPredicateEntity) EncodeMsg(en *msgp.Writer) (err error) {
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
	for zzdc, zelx := range z.Subjects {
		err = en.WriteString(zzdc)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(zelx)))
		if err != nil {
			return
		}
		for zbal, zjqz := range zelx {
			err = en.WriteString(zbal)
			if err != nil {
				return
			}
			err = en.WriteUint32(zjqz)
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
	for zkct, ztmt := range z.Objects {
		err = en.WriteString(zkct)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(ztmt)))
		if err != nil {
			return
		}
		for ztco, zana := range ztmt {
			err = en.WriteString(ztco)
			if err != nil {
				return
			}
			err = en.WriteUint32(zana)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BytesPredicateEntity) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "p"
	o = append(o, 0x83, 0xa1, 0x70)
	o = msgp.AppendBytes(o, z.PK[:])
	// string "s"
	o = append(o, 0xa1, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Subjects)))
	for zzdc, zelx := range z.Subjects {
		o = msgp.AppendString(o, zzdc)
		o = msgp.AppendMapHeader(o, uint32(len(zelx)))
		for zbal, zjqz := range zelx {
			o = msgp.AppendString(o, zbal)
			o = msgp.AppendUint32(o, zjqz)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for zkct, ztmt := range z.Objects {
		o = msgp.AppendString(o, zkct)
		o = msgp.AppendMapHeader(o, uint32(len(ztmt)))
		for ztco, zana := range ztmt {
			o = msgp.AppendString(o, ztco)
			o = msgp.AppendUint32(o, zana)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BytesPredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zrsc uint32
	zrsc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zrsc > 0 {
		zrsc--
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
			var zctn uint32
			zctn, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && zctn > 0 {
				z.Subjects = make(map[string]map[string]uint32, zctn)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for zctn > 0 {
				var zzdc string
				var zelx map[string]uint32
				zctn--
				zzdc, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zswy uint32
				zswy, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if zelx == nil && zswy > 0 {
					zelx = make(map[string]uint32, zswy)
				} else if len(zelx) > 0 {
					for key, _ := range zelx {
						delete(zelx, key)
					}
				}
				for zswy > 0 {
					var zbal string
					var zjqz uint32
					zswy--
					zbal, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					zjqz, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					zelx[zbal] = zjqz
				}
				z.Subjects[zzdc] = zelx
			}
		case "o":
			var znsg uint32
			znsg, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && znsg > 0 {
				z.Objects = make(map[string]map[string]uint32, znsg)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for znsg > 0 {
				var zkct string
				var ztmt map[string]uint32
				znsg--
				zkct, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var zrus uint32
				zrus, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if ztmt == nil && zrus > 0 {
					ztmt = make(map[string]uint32, zrus)
				} else if len(ztmt) > 0 {
					for key, _ := range ztmt {
						delete(ztmt, key)
					}
				}
				for zrus > 0 {
					var ztco string
					var zana uint32
					zrus--
					ztco, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					zana, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					ztmt[ztco] = zana
				}
				z.Objects[zkct] = ztmt
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
func (z *BytesPredicateEntity) Msgsize() (s int) {
	s = 1 + 2 + msgp.ArrayHeaderSize + (8 * (msgp.ByteSize)) + 2 + msgp.MapHeaderSize
	if z.Subjects != nil {
		for zzdc, zelx := range z.Subjects {
			_ = zelx
			s += msgp.StringPrefixSize + len(zzdc) + msgp.MapHeaderSize
			if zelx != nil {
				for zbal, zjqz := range zelx {
					_ = zjqz
					s += msgp.StringPrefixSize + len(zbal) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for zkct, ztmt := range z.Objects {
			_ = ztmt
			s += msgp.StringPrefixSize + len(zkct) + msgp.MapHeaderSize
			if ztmt != nil {
				for ztco, zana := range ztmt {
					_ = zana
					s += msgp.StringPrefixSize + len(ztco) + msgp.Uint32Size
				}
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *HashKey) DecodeMsg(dc *msgp.Reader) (err error) {
	err = dc.ReadExactBytes(z[:])
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *HashKey) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes(z[:])
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *HashKey) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, z[:])
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *HashKey) UnmarshalMsg(bts []byte) (o []byte, err error) {
	bts, err = msgp.ReadExactBytes(bts, z[:])
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *HashKey) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (8 * (msgp.ByteSize))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *KeyType) DecodeMsg(dc *msgp.Reader) (err error) {
	err = dc.ReadExactBytes(z[:])
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *KeyType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes(z[:])
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *KeyType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, z[:])
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *KeyType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	bts, err = msgp.ReadExactBytes(bts, z[:])
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *KeyType) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (4 * (msgp.ByteSize))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Version) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zfzb uint32
	zfzb, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zfzb > 0 {
		zfzb--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Timestamp":
			z.Timestamp, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Name":
			z.Name, err = dc.ReadString()
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
func (z Version) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Timestamp"
	err = en.Append(0x82, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.Timestamp)
	if err != nil {
		return
	}
	// write "Name"
	err = en.Append(0xa4, 0x4e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Name)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Version) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Timestamp"
	o = append(o, 0x82, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o = msgp.AppendUint64(o, z.Timestamp)
	// string "Name"
	o = append(o, 0xa4, 0x4e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Name)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Version) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zsbo uint32
	zsbo, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zsbo > 0 {
		zsbo--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Timestamp":
			z.Timestamp, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Name":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
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
func (z Version) Msgsize() (s int) {
	s = 1 + 10 + msgp.Uint64Size + 5 + msgp.StringPrefixSize + len(z.Name)
	return
}
