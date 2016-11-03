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
		case "e":
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Edges == nil && msz > 0 {
				z.Edges = make(map[string][][4]byte, msz)
			} else if len(z.Edges) > 0 {
				for key, _ := range z.Edges {
					delete(z.Edges, key)
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
		case "e":
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Edges == nil && msz > 0 {
				z.Edges = make(map[string][][4]byte, msz)
			} else if len(z.Edges) > 0 {
				for key, _ := range z.Edges {
					delete(z.Edges, key)
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
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Subjects) >= int(xsz) {
				z.Subjects = z.Subjects[:xsz]
			} else {
				z.Subjects = make([][4]byte, xsz)
			}
			for hct := range z.Subjects {
				err = dc.ReadExactBytes(z.Subjects[hct][:])
				if err != nil {
					return
				}
			}
		case "o":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Objects) >= int(xsz) {
				z.Objects = z.Objects[:xsz]
			} else {
				z.Objects = make([][4]byte, xsz)
			}
			for xhx := range z.Objects {
				err = dc.ReadExactBytes(z.Objects[xhx][:])
				if err != nil {
					return
				}
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
	err = en.WriteArrayHeader(uint32(len(z.Subjects)))
	if err != nil {
		return
	}
	for hct := range z.Subjects {
		err = en.WriteBytes(z.Subjects[hct][:])
		if err != nil {
			return
		}
	}
	// write "o"
	err = en.Append(0xa1, 0x6f)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Objects)))
	if err != nil {
		return
	}
	for xhx := range z.Objects {
		err = en.WriteBytes(z.Objects[xhx][:])
		if err != nil {
			return
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
	o = msgp.AppendArrayHeader(o, uint32(len(z.Subjects)))
	for hct := range z.Subjects {
		o = msgp.AppendBytes(o, z.Subjects[hct][:])
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Objects)))
	for xhx := range z.Objects {
		o = msgp.AppendBytes(o, z.Objects[xhx][:])
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
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Subjects) >= int(xsz) {
				z.Subjects = z.Subjects[:xsz]
			} else {
				z.Subjects = make([][4]byte, xsz)
			}
			for hct := range z.Subjects {
				bts, err = msgp.ReadExactBytes(bts, z.Subjects[hct][:])
				if err != nil {
					return
				}
			}
		case "o":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Objects) >= int(xsz) {
				z.Objects = z.Objects[:xsz]
			} else {
				z.Objects = make([][4]byte, xsz)
			}
			for xhx := range z.Objects {
				bts, err = msgp.ReadExactBytes(bts, z.Objects[xhx][:])
				if err != nil {
					return
				}
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
	s = 1 + 2 + msgp.ArrayHeaderSize + (4 * (msgp.ByteSize)) + 2 + msgp.ArrayHeaderSize + (len(z.Subjects) * (4 * (msgp.ByteSize))) + 2 + msgp.ArrayHeaderSize + (len(z.Objects) * (4 * (msgp.ByteSize)))
	return
}
