package db

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *Entity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var hct uint32
	hct, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for hct > 0 {
		hct--
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
			var cua uint32
			cua, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.InEdges == nil && cua > 0 {
				z.InEdges = make(map[string][]Key, cua)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for cua > 0 {
				cua--
				var xvk string
				var bzg []Key
				xvk, err = dc.ReadString()
				if err != nil {
					return
				}
				var xhx uint32
				xhx, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(bzg) >= int(xhx) {
					bzg = bzg[:xhx]
				} else {
					bzg = make([]Key, xhx)
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
			var lqf uint32
			lqf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.OutEdges == nil && lqf > 0 {
				z.OutEdges = make(map[string][]Key, lqf)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for lqf > 0 {
				lqf--
				var cmr string
				var ajw []Key
				cmr, err = dc.ReadString()
				if err != nil {
					return
				}
				var daf uint32
				daf, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if cap(ajw) >= int(daf) {
					ajw = ajw[:daf]
				} else {
					ajw = make([]Key, daf)
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
	var pks uint32
	pks, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for pks > 0 {
		pks--
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
			var jfb uint32
			jfb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.InEdges == nil && jfb > 0 {
				z.InEdges = make(map[string][]Key, jfb)
			} else if len(z.InEdges) > 0 {
				for key, _ := range z.InEdges {
					delete(z.InEdges, key)
				}
			}
			for jfb > 0 {
				var xvk string
				var bzg []Key
				jfb--
				xvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var cxo uint32
				cxo, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(bzg) >= int(cxo) {
					bzg = bzg[:cxo]
				} else {
					bzg = make([]Key, cxo)
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
			var eff uint32
			eff, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.OutEdges == nil && eff > 0 {
				z.OutEdges = make(map[string][]Key, eff)
			} else if len(z.OutEdges) > 0 {
				for key, _ := range z.OutEdges {
					delete(z.OutEdges, key)
				}
			}
			for eff > 0 {
				var cmr string
				var ajw []Key
				eff--
				cmr, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var rsw uint32
				rsw, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if cap(ajw) >= int(rsw) {
					ajw = ajw[:rsw]
				} else {
					ajw = make([]Key, rsw)
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
	var kgt uint32
	kgt, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && kgt > 0 {
		(*z) = make(NamespaceIndex, kgt)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for kgt > 0 {
		kgt--
		var obc string
		var snv string
		obc, err = dc.ReadString()
		if err != nil {
			return
		}
		snv, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[obc] = snv
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z NamespaceIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for ema, pez := range z {
		err = en.WriteString(ema)
		if err != nil {
			return
		}
		err = en.WriteString(pez)
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
	for ema, pez := range z {
		o = msgp.AppendString(o, ema)
		o = msgp.AppendString(o, pez)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NamespaceIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var yzr uint32
	yzr, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && yzr > 0 {
		(*z) = make(NamespaceIndex, yzr)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for yzr > 0 {
		var qke string
		var qyh string
		yzr--
		qke, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		qyh, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[qke] = qyh
	}
	o = bts
	return
}

func (z NamespaceIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for ywj, jpj := range z {
			_ = jpj
			s += msgp.StringPrefixSize + len(ywj) + msgp.StringPrefixSize + len(jpj)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var eth uint32
	eth, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && eth > 0 {
		(*z) = make(PredIndex, eth)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for eth > 0 {
		eth--
		var gmo string
		var taf *PredicateEntity
		gmo, err = dc.ReadString()
		if err != nil {
			return
		}
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			taf = nil
		} else {
			if taf == nil {
				taf = new(PredicateEntity)
			}
			err = taf.DecodeMsg(dc)
			if err != nil {
				return
			}
		}
		(*z)[gmo] = taf
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PredIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for sbz, rjx := range z {
		err = en.WriteString(sbz)
		if err != nil {
			return
		}
		if rjx == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = rjx.EncodeMsg(en)
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
	for sbz, rjx := range z {
		o = msgp.AppendString(o, sbz)
		if rjx == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = rjx.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var rbe uint32
	rbe, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && rbe > 0 {
		(*z) = make(PredIndex, rbe)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for rbe > 0 {
		var awn string
		var wel *PredicateEntity
		rbe--
		awn, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			wel = nil
		} else {
			if wel == nil {
				wel = new(PredicateEntity)
			}
			bts, err = wel.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
		(*z)[awn] = wel
	}
	o = bts
	return
}

func (z PredIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for mfd, zdc := range z {
			_ = zdc
			s += msgp.StringPrefixSize + len(mfd)
			if zdc == nil {
				s += msgp.NilSize
			} else {
				s += zdc.Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredicateEntity) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var inl uint32
	inl, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for inl > 0 {
		inl--
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
			var are uint32
			are, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Subjects == nil && are > 0 {
				z.Subjects = make(map[string]map[string]uint32, are)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for are > 0 {
				are--
				var elx string
				var bal map[string]uint32
				elx, err = dc.ReadString()
				if err != nil {
					return
				}
				var ljy uint32
				ljy, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if bal == nil && ljy > 0 {
					bal = make(map[string]uint32, ljy)
				} else if len(bal) > 0 {
					for key, _ := range bal {
						delete(bal, key)
					}
				}
				for ljy > 0 {
					ljy--
					var jqz string
					var kct uint32
					jqz, err = dc.ReadString()
					if err != nil {
						return
					}
					kct, err = dc.ReadUint32()
					if err != nil {
						return
					}
					bal[jqz] = kct
				}
				z.Subjects[elx] = bal
			}
		case "o":
			var ixj uint32
			ixj, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Objects == nil && ixj > 0 {
				z.Objects = make(map[string]map[string]uint32, ixj)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for ixj > 0 {
				ixj--
				var tmt string
				var tco map[string]uint32
				tmt, err = dc.ReadString()
				if err != nil {
					return
				}
				var rsc uint32
				rsc, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				if tco == nil && rsc > 0 {
					tco = make(map[string]uint32, rsc)
				} else if len(tco) > 0 {
					for key, _ := range tco {
						delete(tco, key)
					}
				}
				for rsc > 0 {
					rsc--
					var ana string
					var tyy uint32
					ana, err = dc.ReadString()
					if err != nil {
						return
					}
					tyy, err = dc.ReadUint32()
					if err != nil {
						return
					}
					tco[ana] = tyy
				}
				z.Objects[tmt] = tco
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
	for elx, bal := range z.Subjects {
		err = en.WriteString(elx)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(bal)))
		if err != nil {
			return
		}
		for jqz, kct := range bal {
			err = en.WriteString(jqz)
			if err != nil {
				return
			}
			err = en.WriteUint32(kct)
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
	for tmt, tco := range z.Objects {
		err = en.WriteString(tmt)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(tco)))
		if err != nil {
			return
		}
		for ana, tyy := range tco {
			err = en.WriteString(ana)
			if err != nil {
				return
			}
			err = en.WriteUint32(tyy)
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
	for elx, bal := range z.Subjects {
		o = msgp.AppendString(o, elx)
		o = msgp.AppendMapHeader(o, uint32(len(bal)))
		for jqz, kct := range bal {
			o = msgp.AppendString(o, jqz)
			o = msgp.AppendUint32(o, kct)
		}
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendMapHeader(o, uint32(len(z.Objects)))
	for tmt, tco := range z.Objects {
		o = msgp.AppendString(o, tmt)
		o = msgp.AppendMapHeader(o, uint32(len(tco)))
		for ana, tyy := range tco {
			o = msgp.AppendString(o, ana)
			o = msgp.AppendUint32(o, tyy)
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredicateEntity) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var ctn uint32
	ctn, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for ctn > 0 {
		ctn--
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
			var swy uint32
			swy, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Subjects == nil && swy > 0 {
				z.Subjects = make(map[string]map[string]uint32, swy)
			} else if len(z.Subjects) > 0 {
				for key, _ := range z.Subjects {
					delete(z.Subjects, key)
				}
			}
			for swy > 0 {
				var elx string
				var bal map[string]uint32
				swy--
				elx, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var nsg uint32
				nsg, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if bal == nil && nsg > 0 {
					bal = make(map[string]uint32, nsg)
				} else if len(bal) > 0 {
					for key, _ := range bal {
						delete(bal, key)
					}
				}
				for nsg > 0 {
					var jqz string
					var kct uint32
					nsg--
					jqz, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					kct, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					bal[jqz] = kct
				}
				z.Subjects[elx] = bal
			}
		case "o":
			var rus uint32
			rus, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Objects == nil && rus > 0 {
				z.Objects = make(map[string]map[string]uint32, rus)
			} else if len(z.Objects) > 0 {
				for key, _ := range z.Objects {
					delete(z.Objects, key)
				}
			}
			for rus > 0 {
				var tmt string
				var tco map[string]uint32
				rus--
				tmt, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				var svm uint32
				svm, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				if tco == nil && svm > 0 {
					tco = make(map[string]uint32, svm)
				} else if len(tco) > 0 {
					for key, _ := range tco {
						delete(tco, key)
					}
				}
				for svm > 0 {
					var ana string
					var tyy uint32
					svm--
					ana, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					tyy, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						return
					}
					tco[ana] = tyy
				}
				z.Objects[tmt] = tco
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
		for elx, bal := range z.Subjects {
			_ = bal
			s += msgp.StringPrefixSize + len(elx) + msgp.MapHeaderSize
			if bal != nil {
				for jqz, kct := range bal {
					_ = kct
					s += msgp.StringPrefixSize + len(jqz) + msgp.Uint32Size
				}
			}
		}
	}
	s += 2 + msgp.MapHeaderSize
	if z.Objects != nil {
		for tmt, tco := range z.Objects {
			_ = tco
			s += msgp.StringPrefixSize + len(tmt) + msgp.MapHeaderSize
			if tco != nil {
				for ana, tyy := range tco {
					_ = tyy
					s += msgp.StringPrefixSize + len(ana) + msgp.Uint32Size
				}
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RelshipIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var qgz uint32
	qgz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && qgz > 0 {
		(*z) = make(RelshipIndex, qgz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for qgz > 0 {
		qgz--
		var sbo string
		var jif string
		sbo, err = dc.ReadString()
		if err != nil {
			return
		}
		jif, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[sbo] = jif
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RelshipIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for snw, tls := range z {
		err = en.WriteString(snw)
		if err != nil {
			return
		}
		err = en.WriteString(tls)
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
	for snw, tls := range z {
		o = msgp.AppendString(o, snw)
		o = msgp.AppendString(o, tls)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RelshipIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var opb uint32
	opb, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && opb > 0 {
		(*z) = make(RelshipIndex, opb)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for opb > 0 {
		var mvo string
		var igk string
		opb--
		mvo, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		igk, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[mvo] = igk
	}
	o = bts
	return
}

func (z RelshipIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for uop, edl := range z {
			_ = edl
			s += msgp.StringPrefixSize + len(uop) + msgp.StringPrefixSize + len(edl)
		}
	}
	return
}
