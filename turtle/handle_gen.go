package turtle

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"C"

	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Triple) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var xvk uint32
	xvk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for xvk > 0 {
		xvk--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "s":
			var bzg uint32
			bzg, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for bzg > 0 {
				bzg--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "n":
					z.Subject.Namespace, err = dc.ReadString()
					if err != nil {
						return
					}
				case "v":
					z.Subject.Value, err = dc.ReadString()
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
		case "p":
			var bai uint32
			bai, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for bai > 0 {
				bai--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "n":
					z.Predicate.Namespace, err = dc.ReadString()
					if err != nil {
						return
					}
				case "v":
					z.Predicate.Value, err = dc.ReadString()
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
		case "o":
			var cmr uint32
			cmr, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for cmr > 0 {
				cmr--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "n":
					z.Object.Namespace, err = dc.ReadString()
					if err != nil {
						return
					}
				case "v":
					z.Object.Value, err = dc.ReadString()
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
func (z *Triple) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "s"
	// map header, size 2
	// write "n"
	err = en.Append(0x83, 0xa1, 0x73, 0x82, 0xa1, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Subject.Namespace)
	if err != nil {
		return
	}
	// write "v"
	err = en.Append(0xa1, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Subject.Value)
	if err != nil {
		return
	}
	// write "p"
	// map header, size 2
	// write "n"
	err = en.Append(0xa1, 0x70, 0x82, 0xa1, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Predicate.Namespace)
	if err != nil {
		return
	}
	// write "v"
	err = en.Append(0xa1, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Predicate.Value)
	if err != nil {
		return
	}
	// write "o"
	// map header, size 2
	// write "n"
	err = en.Append(0xa1, 0x6f, 0x82, 0xa1, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Object.Namespace)
	if err != nil {
		return
	}
	// write "v"
	err = en.Append(0xa1, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Object.Value)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Triple) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "s"
	// map header, size 2
	// string "n"
	o = append(o, 0x83, 0xa1, 0x73, 0x82, 0xa1, 0x6e)
	o = msgp.AppendString(o, z.Subject.Namespace)
	// string "v"
	o = append(o, 0xa1, 0x76)
	o = msgp.AppendString(o, z.Subject.Value)
	// string "p"
	// map header, size 2
	// string "n"
	o = append(o, 0xa1, 0x70, 0x82, 0xa1, 0x6e)
	o = msgp.AppendString(o, z.Predicate.Namespace)
	// string "v"
	o = append(o, 0xa1, 0x76)
	o = msgp.AppendString(o, z.Predicate.Value)
	// string "o"
	// map header, size 2
	// string "n"
	o = append(o, 0xa1, 0x6f, 0x82, 0xa1, 0x6e)
	o = msgp.AppendString(o, z.Object.Namespace)
	// string "v"
	o = append(o, 0xa1, 0x76)
	o = msgp.AppendString(o, z.Object.Value)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Triple) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var ajw uint32
	ajw, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for ajw > 0 {
		ajw--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "s":
			var wht uint32
			wht, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for wht > 0 {
				wht--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "n":
					z.Subject.Namespace, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "v":
					z.Subject.Value, bts, err = msgp.ReadStringBytes(bts)
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
		case "p":
			var hct uint32
			hct, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for hct > 0 {
				hct--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "n":
					z.Predicate.Namespace, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "v":
					z.Predicate.Value, bts, err = msgp.ReadStringBytes(bts)
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
		case "o":
			var cua uint32
			cua, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for cua > 0 {
				cua--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "n":
					z.Object.Namespace, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "v":
					z.Object.Value, bts, err = msgp.ReadStringBytes(bts)
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

func (z *Triple) Msgsize() (s int) {
	s = 1 + 2 + 1 + 2 + msgp.StringPrefixSize + len(z.Subject.Namespace) + 2 + msgp.StringPrefixSize + len(z.Subject.Value) + 2 + 1 + 2 + msgp.StringPrefixSize + len(z.Predicate.Namespace) + 2 + msgp.StringPrefixSize + len(z.Predicate.Value) + 2 + 1 + 2 + msgp.StringPrefixSize + len(z.Object.Namespace) + 2 + msgp.StringPrefixSize + len(z.Object.Value)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *URI) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var xhx uint32
	xhx, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for xhx > 0 {
		xhx--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "n":
			z.Namespace, err = dc.ReadString()
			if err != nil {
				return
			}
		case "v":
			z.Value, err = dc.ReadString()
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
func (z URI) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "n"
	err = en.Append(0x82, 0xa1, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Namespace)
	if err != nil {
		return
	}
	// write "v"
	err = en.Append(0xa1, 0x76)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Value)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z URI) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "n"
	o = append(o, 0x82, 0xa1, 0x6e)
	o = msgp.AppendString(o, z.Namespace)
	// string "v"
	o = append(o, 0xa1, 0x76)
	o = msgp.AppendString(o, z.Value)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *URI) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var lqf uint32
	lqf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for lqf > 0 {
		lqf--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "n":
			z.Namespace, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "v":
			z.Value, bts, err = msgp.ReadStringBytes(bts)
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

func (z URI) Msgsize() (s int) {
	s = 1 + 2 + msgp.StringPrefixSize + len(z.Namespace) + 2 + msgp.StringPrefixSize + len(z.Value)
	return
}
