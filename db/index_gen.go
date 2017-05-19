package db

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *NamespaceIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zajw uint32
	zajw, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zajw > 0 {
		(*z) = make(NamespaceIndex, zajw)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zajw > 0 {
		zajw--
		var zbai string
		var zcmr string
		zbai, err = dc.ReadString()
		if err != nil {
			return
		}
		zcmr, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[zbai] = zcmr
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z NamespaceIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zwht, zhct := range z {
		err = en.WriteString(zwht)
		if err != nil {
			return
		}
		err = en.WriteString(zhct)
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
	for zwht, zhct := range z {
		o = msgp.AppendString(o, zwht)
		o = msgp.AppendString(o, zhct)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NamespaceIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zlqf uint32
	zlqf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zlqf > 0 {
		(*z) = make(NamespaceIndex, zlqf)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zlqf > 0 {
		var zcua string
		var zxhx string
		zlqf--
		zcua, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zxhx, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[zcua] = zxhx
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z NamespaceIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zdaf, zpks := range z {
			_ = zpks
			s += msgp.StringPrefixSize + len(zdaf) + msgp.StringPrefixSize + len(zpks)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PredIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zxpk uint32
	zxpk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zxpk > 0 {
		(*z) = make(PredIndex, zxpk)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zxpk > 0 {
		zxpk--
		var zeff string
		var zrsw *PredicateEntity
		zeff, err = dc.ReadString()
		if err != nil {
			return
		}
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			zrsw = nil
		} else {
			if zrsw == nil {
				zrsw = new(PredicateEntity)
			}
			err = zrsw.DecodeMsg(dc)
			if err != nil {
				return
			}
		}
		(*z)[zeff] = zrsw
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PredIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zdnj, zobc := range z {
		err = en.WriteString(zdnj)
		if err != nil {
			return
		}
		if zobc == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zobc.EncodeMsg(en)
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
	for zdnj, zobc := range z {
		o = msgp.AppendString(o, zdnj)
		if zobc == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zobc.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PredIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zema uint32
	zema, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zema > 0 {
		(*z) = make(PredIndex, zema)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zema > 0 {
		var zsnv string
		var zkgt *PredicateEntity
		zema--
		zsnv, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			zkgt = nil
		} else {
			if zkgt == nil {
				zkgt = new(PredicateEntity)
			}
			bts, err = zkgt.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
		(*z)[zsnv] = zkgt
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z PredIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zpez, zqke := range z {
			_ = zqke
			s += msgp.StringPrefixSize + len(zpez)
			if zqke == nil {
				s += msgp.NilSize
			} else {
				s += zqke.Msgsize()
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RelshipIndex) DecodeMsg(dc *msgp.Reader) (err error) {
	var zzpf uint32
	zzpf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zzpf > 0 {
		(*z) = make(RelshipIndex, zzpf)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zzpf > 0 {
		zzpf--
		var zywj string
		var zjpj string
		zywj, err = dc.ReadString()
		if err != nil {
			return
		}
		zjpj, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z)[zywj] = zjpj
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RelshipIndex) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zrfe, zgmo := range z {
		err = en.WriteString(zrfe)
		if err != nil {
			return
		}
		err = en.WriteString(zgmo)
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
	for zrfe, zgmo := range z {
		o = msgp.AppendString(o, zrfe)
		o = msgp.AppendString(o, zgmo)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RelshipIndex) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zsbz uint32
	zsbz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zsbz > 0 {
		(*z) = make(RelshipIndex, zsbz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for zsbz > 0 {
		var ztaf string
		var zeth string
		zsbz--
		ztaf, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zeth, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z)[ztaf] = zeth
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RelshipIndex) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zrjx, zawn := range z {
			_ = zawn
			s += msgp.StringPrefixSize + len(zrjx) + msgp.StringPrefixSize + len(zawn)
		}
	}
	return
}
