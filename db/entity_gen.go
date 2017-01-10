package db

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

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
			err = z.PK.DecodeMsg(dc)
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
				z.InEdges = make(map[string][]Key, msz)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for msz > 0 {
				msz--
				var xvk string
				var bzg []Key
				xvk, err = dc.ReadString()
				if err != nil {
					return
				}
				var xsz uint32
				xsz, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(bzg) >= int(xsz) {
					bzg = bzg[:xsz]
				} else {
					bzg = make([]Key, xsz)
				}
				for bai := range bzg {
					err = bzg[bai].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.InEdges[xvk] = bzg
			}
		case "eout":
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && msz > 0 {
				z.OutEdges = make(map[string][]Key, msz)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for msz > 0 {
				msz--
				var cmr string
				var ajw []Key
				cmr, err = dc.ReadString()
				if err != nil {
					return
				}
				var xsz uint32
				xsz, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(ajw) >= int(xsz) {
					ajw = ajw[:xsz]
				} else {
					ajw = make([]Key, xsz)
				}
				for wht := range ajw {
					err = ajw[wht].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.OutEdges[cmr] = ajw
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
	for xvk, bzg := range z.InEdges {
		err = en.WriteString(xvk)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(bzg)))
		if err != nil {
			return
		}
		for bai := range bzg {
			err = bzg[bai].EncodeMsg(en)
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
	for cmr, ajw := range z.OutEdges {
		err = en.WriteString(cmr)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(ajw)))
		if err != nil {
			return
		}
		for wht := range ajw {
			err = ajw[wht].EncodeMsg(en)
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
	for xvk, bzg := range z.InEdges {
		o = msgp.AppendString(o, xvk)
		o = msgp.AppendArrayHeader(o, uint32(len(bzg)))
		for bai := range bzg {
			o, err = bzg[bai].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "eout"
	o = append(o, 0xa4, 0x65, 0x6f, 0x75, 0x74)
	o = msgp.AppendMapHeader(o, uint32(len(z.OutEdges)))
	for cmr, ajw := range z.OutEdges {
		o = msgp.AppendString(o, cmr)
		o = msgp.AppendArrayHeader(o, uint32(len(ajw)))
		for wht := range ajw {
			o, err = ajw[wht].MarshalMsg(o)
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
			bts, err = z.PK.UnmarshalMsg(bts)
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
				z.InEdges = make(map[string][]Key, msz)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for msz > 0 {
				var xvk string
				var bzg []Key
				msz--
				xvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var xsz uint32
				xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(bzg) >= int(xsz) {
					bzg = bzg[:xsz]
				} else {
					bzg = make([]Key, xsz)
				}
				for bai := range bzg {
					bts, err = bzg[bai].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.InEdges[xvk] = bzg
			}
		case "eout":
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && msz > 0 {
				z.OutEdges = make(map[string][]Key, msz)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for msz > 0 {
				var cmr string
				var ajw []Key
				msz--
				cmr, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var xsz uint32
				xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(ajw) >= int(xsz) {
					ajw = ajw[:xsz]
				} else {
					ajw = make([]Key, xsz)
				}
				for wht := range ajw {
					bts, err = ajw[wht].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.OutEdges[cmr] = ajw
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
	s = 1 + 2 + z.PK.Msgsize() + 4 + msgp.MapHeaderSize
	if z.InEdges != nil {
		for xvk, bzg := range z.InEdges {
			_ = bzg
			s += msgp.StringPrefixSize + len(xvk) + msgp.ArrayHeaderSize
			for bai := range bzg {
				s += bzg[bai].Msgsize()
			}
		}
	}
	s += 5 + msgp.MapHeaderSize
	if z.OutEdges != nil {
		for cmr, ajw := range z.OutEdges {
			_ = ajw
			s += msgp.StringPrefixSize + len(cmr) + msgp.ArrayHeaderSize
			for wht := range ajw {
				s += ajw[wht].Msgsize()
			}
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
		var xhx string
		var lqf string
		xhx, err = dc.ReadString()
		if err != nil {
			return
		}
		lqf, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[xhx] = lqf
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z NamespaceIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for daf, pks := range z {
		err = en.WriteString(daf)
		if err != nil {
			return
		}
		err = en.WriteString(pks)
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
	for daf, pks := range z {
		o = msgp.AppendString(o, daf)
		o = msgp.AppendString(o, pks)
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
		var jfb string
		var cxo string
		msz--
		jfb, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		cxo, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[jfb] = cxo
	}
	o = bts
	return
}

func (z NamespaceIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for eff, rsw := range z {
			_ = rsw
			s += msgp.StringPrefixSize + len(eff) + msgp.StringPrefixSize + len(rsw)
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
		var obc string
		var snv *PredicateEntity
		obc, err = dc.ReadString()
		if err != nil {
			return
		}
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			snv = nil
		} else {
			if snv == nil {
				snv = new(PredicateEntity)
			}
			err = snv.DecodeMsg(dc)
			if err != nil {
				return
			}
		}
		(*z)[obc] = snv
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PredIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for kgt, ema := range z {
		err = en.WriteString(kgt)
		if err != nil {
			return
		}
		if ema == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = ema.EncodeMsg(en)
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
	for kgt, ema := range z {
		o = msgp.AppendString(o, kgt)
		if ema == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = ema.MarshalMsg(o)
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
		var pez string
		var qke *PredicateEntity
		msz--
		pez, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			qke = nil
		} else {
			if qke == nil {
				qke = new(PredicateEntity)
			}
			bts, err = qke.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
		(*z)[pez] = qke
	}
	o = bts
	return
}

func (z PredIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for qyh, yzr := range z {
			_ = yzr
			s += msgp.StringPrefixSize + len(qyh)
			if yzr == nil {
				s += msgp.NilSize
			} else {
				s += yzr.Msgsize()
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
			err = z.PK.DecodeMsg(dc)
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
				var ywj string
				var jpj map[string]uint32
				ywj, err = dc.ReadString()
				if err != nil {
					return
				}
				var msz uint32
				msz, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if jpj == nil && msz > 0 {
					jpj = make(map[string]uint32, msz)
				} else if len(jpj) > 0 {
					for key, _ := range jpj {
						delete(jpj, key)
					}
				}
				for msz > 0 {
					msz--
					var zpf string
					var rfe uint32
					zpf, err = dc.ReadString()
					if err != nil {
						return
					}
					rfe, err = dc.ReadUint32()
					if err != nil {
						return
					}
					jpj[zpf] = rfe
				}
				z.Subjects[ywj] = jpj
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
				z.Objects[gmo] = taf
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
	for ywj, jpj := range z.Subjects {
		err = en.WriteString(ywj)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(jpj)))
		if err != nil {
			return
		}
		for zpf, rfe := range jpj {
			err = en.WriteString(zpf)
			if err != nil {
				return
			}
			err = en.WriteUint32(rfe)
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
	for gmo, taf := range z.Objects {
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
	for ywj, jpj := range z.Subjects {
		o = msgp.AppendString(o, ywj)
		o = msgp.AppendMapHeader(o, uint32(len(jpj)))
		for zpf, rfe := range jpj {
			o = msgp.AppendString(o, zpf)
			o = msgp.AppendUint32(o, rfe)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for gmo, taf := range z.Objects {
		o = msgp.AppendString(o, gmo)
		o = msgp.AppendMapHeader(o, uint32(len(taf)))
		for eth, sbz := range taf {
			o = msgp.AppendString(o, eth)
			o = msgp.AppendUint32(o, sbz)
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
			bts, err = z.PK.UnmarshalMsg(bts)
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
				var ywj string
				var jpj map[string]uint32
				msz--
				ywj, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var msz uint32
				msz, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if jpj == nil && msz > 0 {
					jpj = make(map[string]uint32, msz)
				} else if len(jpj) > 0 {
					for key, _ := range jpj {
						delete(jpj, key)
					}
				}
				for msz > 0 {
					var zpf string
					var rfe uint32
					msz--
					zpf, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					rfe, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					jpj[zpf] = rfe
				}
				z.Subjects[ywj] = jpj
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
				z.Objects[gmo] = taf
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
	s = 1 + 2 + z.PK.Msgsize() + 2 + msgp.MapHeaderSize
	if z.Subjects != nil {
		for ywj, jpj := range z.Subjects {
			_ = jpj
			s += msgp.StringPrefixSize + len(ywj) + msgp.MapHeaderSize
			if jpj != nil {
				for zpf, rfe := range jpj {
					_ = rfe
					s += msgp.StringPrefixSize + len(zpf) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for gmo, taf := range z.Objects {
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
		var wel string
		var rbe string
		wel, err = dc.ReadString()
		if err != nil {
			return
		}
		rbe, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[wel] = rbe
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RelshipIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for mfd, zdc := range z {
		err = en.WriteString(mfd)
		if err != nil {
			return
		}
		err = en.WriteString(zdc)
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
	for mfd, zdc := range z {
		o = msgp.AppendString(o, mfd)
		o = msgp.AppendString(o, zdc)
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
		var elx string
		var bal string
		msz--
		elx, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		bal, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[elx] = bal
	}
	o = bts
	return
}

func (z RelshipIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for jqz, kct := range z {
			_ = kct
			s += msgp.StringPrefixSize + len(jqz) + msgp.StringPrefixSize + len(kct)
		}
	}
	return
}
