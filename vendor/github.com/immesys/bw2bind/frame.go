package bw2bind

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/immesys/bw2/objects"
)

const (
	cmdHello        = "helo"
	cmdPublish      = "publ"
	cmdSubscribe    = "subs"
	cmdPersist      = "pers"
	cmdList         = "list"
	cmdQuery        = "quer"
	cmdTapSubscribe = "tsub"
	cmdTapQuery     = "tque"
	cmdMakeDot      = "makd"
	cmdMakeEntity   = "make"
	cmdMakeChain    = "makc"
	cmdBuildChain   = "bldc"
	//	cmdAddPrefDot   = "adpd"
	//cmdAddPrefChain = "adpc"
	//	cmdDelPrefDot   = "dlpd"
	//	cmdDelPrefChain = "dlpc"
	cmdSetEntity = "sete"

	//New for 2.1.x
	cmdPutDot                = "putd"
	cmdPutEntity             = "pute"
	cmdPutChain              = "putc"
	cmdEntityBalances        = "ebal"
	cmdAddressBalance        = "abal"
	cmdBCInteractionParams   = "bcip"
	cmdTransfer              = "xfer"
	cmdMakeShortAlias        = "mksa"
	cmdMakeLongAlias         = "mkla"
	cmdResolveAlias          = "resa"
	cmdNewDROffer            = "ndro"
	cmdAcceptDROffer         = "adro"
	cmdResolveRegistryObject = "rsro"
	cmdUpdateSRVRecord       = "usrv"
	cmdListDROffers          = "ldro"
	cmdMakeView              = "mkvw"
	cmdSubscribeView         = "vsub"
	cmdPublishView           = "vpub"
	cmdListView              = "vlst"
	cmdUnsubscribe           = "usub"
	cmdRevokeDROffer         = "rdro"
	cmdRevokeDRAccept        = "rdra"
	cmdRevokeRO              = "revk"
	cmdPutRevocation         = "prvk"
	cmdFindDots              = "fdot"

	cmdResponse = "resp"
	cmdResult   = "rslt"
)

type header struct {
	Content []byte
	Key     string
	Length  string
	ILength int
}
type roEntry struct {
	RO     objects.RoutingObject
	RONum  string
	Length string
}
type poEntry struct {
	PO           []byte
	PONum        int
	StrPONum     string
	StrLen       string
	StrPODotForm string
}
type frame struct {
	SeqNo   int
	Headers []header
	Cmd     string
	ROs     []roEntry
	POs     []poEntry
	Length  int
}

func createFrame(cmd string, seqno int) *frame {
	return &frame{Cmd: cmd,
		SeqNo:   seqno,
		Headers: make([]header, 0),
		POs:     make([]poEntry, 0),
		ROs:     make([]roEntry, 0),
		Length:  4, //"end\n"
	}
}
func (f *frame) AddHeaderB(k string, v []byte) {
	h := header{Key: k, Content: v, Length: strconv.Itoa(len(v))}
	f.Headers = append(f.Headers, h)
	//6 = 3 for "kv " 1 for space, 1 for newline before content and 1 for newline after
	f.Length += len(k) + len(h.Length) + 6 + len(v)
}
func (f *frame) AddHeader(k string, v string) {
	f.AddHeaderB(k, []byte(v))
}

/*
func (f *frame) GetAllPOs() []PayloadObject {
	rv := make([]PayloadObject, len(f.POs))
	for i, v := range f.POs {
		po := LoadPayloadObject(v.PONum, v.PO)
		rv[i] = po
	}
	return rv
}*/
func (f *frame) NumPOs() int {
	return len(f.POs)
}
func (f *frame) GetPO(num int) (PayloadObject, error) {
	return LoadPayloadObject(f.POs[num].PONum, f.POs[num].PO)
}
func (f *frame) GetAllROs() []objects.RoutingObject {
	rv := make([]objects.RoutingObject, len(f.ROs))
	for i, v := range f.ROs {
		rv[i] = v.RO
	}
	return rv
}
func (f *frame) GetFirstHeaderB(k string) ([]byte, bool) {
	for _, h := range f.Headers {
		if h.Key == k {
			return h.Content, true
		}
	}
	return nil, false
}
func (f *frame) GetFirstHeader(k string) (string, bool) {
	r, ok := f.GetFirstHeaderB(k)
	return string(r), ok
}
func (f *frame) GetAllHeaders(k string) []string {
	var rv []string
	for _, h := range f.Headers {
		if h.Key == k {
			rv = append(rv, string(h.Content))
		}
	}
	return rv
}
func (f *frame) GetAllHeadersB(k string) [][]byte {
	var rv [][]byte
	for _, h := range f.Headers {
		if h.Key == k {
			rv = append(rv, h.Content)
		}
	}
	return rv
}
func (f *frame) AddRoutingObject(ro objects.RoutingObject) {
	re := roEntry{
		RO:     ro,
		RONum:  strconv.Itoa(ro.GetRONum()),
		Length: strconv.Itoa(len(ro.GetContent())),
	}
	f.ROs = append(f.ROs, re)
	//3 for "ro ", 2 for newlines before and after 1 for space
	f.Length += 3 + len(re.RONum) + 1 + len(re.Length) + 1 + len(ro.GetContent()) + 1
}
func (f *frame) AddPayloadObject(po PayloadObject) {
	pe := poEntry{
		PO:           po.GetContents(),
		PONum:        po.GetPONum(),
		StrPONum:     strconv.Itoa(po.GetPONum()),
		StrPODotForm: PONumDotForm(po.GetPONum()),
		StrLen:       strconv.Itoa(len(po.GetContents())),
	}
	f.POs = append(f.POs, pe)
	//3 for "po ",                  colon                space                newline                   newline
	f.Length += 3 + len(pe.StrPONum) + 1 + len(pe.StrPODotForm) + 1 + len(pe.StrLen) + 1 + len(po.GetContents()) + 1
}
func (f *frame) MustResponse() error {
	response, err := f.IsResponse()
	if !response {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Expecting RESP frame %s\n", msg)
	}
	return err
}
func (f *frame) IsResponse() (bool, error) {
	if f == nil {
		return true, fmt.Errorf("channel closed early")
	}
	if f.Cmd == cmdResponse {
		st, stok := f.GetFirstHeader("status")
		if !stok {
			panic("Should not hit this")
		}
		if st == "okay" {
			return true, nil
		}
		code, _ := f.GetFirstHeader("code")
		msg, _ := f.GetFirstHeader("reason")
		return true, fmt.Errorf("[%s] %s", code, msg)
	}
	return false, nil
}
func (f *frame) WriteToStream(s *bufio.Writer) {
	saveError := func(n int, err error) {
		if err != nil {
			fmt.Println("save", err)
		}
	}

	s.WriteString(fmt.Sprintf("%4s %010d %010d\n", f.Cmd, f.Length, f.SeqNo))
	for _, v := range f.Headers {
		saveError(s.WriteString(fmt.Sprintf("kv %s %s\n", v.Key, v.Length)))
		saveError(s.Write(v.Content))
		saveError(s.WriteRune('\n'))
	}
	for _, re := range f.ROs {
		saveError(s.WriteString(fmt.Sprintf("ro %s %s\n",
			re.RONum, re.Length)))
		saveError(s.Write(re.RO.GetContent()))
		saveError(s.WriteRune('\n'))
	}
	for _, pe := range f.POs {
		saveError(s.WriteString(fmt.Sprintf("po %s:%s %s\n",
			pe.StrPODotForm, pe.StrPONum, pe.StrLen)))
		saveError(s.Write(pe.PO))
		saveError(s.WriteRune('\n'))
	}
	saveError(s.WriteString("end\n"))
	if err := s.Flush(); err != nil {
		fmt.Println("flush", err)
	}
}

func loadFrameFromStream(s *bufio.Reader) (f *frame, e error) {
	defer func() {
		if r := recover(); r != nil {
			f = nil
			fmt.Println(r)
			e = errors.New("Malformed frame")
			return
		}
	}()
	hdr := make([]byte, 27)
	if _, e := io.ReadFull(s, hdr); e != nil {
		return nil, e
	}
	//Remember header is
	//    4          15         26
	//CMMD 10DIGITLEN 10DIGITSEQ\n
	f = &frame{}
	f.Cmd = string(hdr[0:4])
	cx, err := strconv.ParseUint(string(hdr[5:15]), 10, 32)
	if err != nil {
		return nil, err
	}
	f.Length = int(cx)
	cx, err = strconv.ParseUint(string(hdr[16:26]), 10, 32)
	if err != nil {
		return nil, err
	}
	f.SeqNo = int(cx)
	for {
		l, err := s.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if string(l) == "end\n" {
			return f, nil
		}
		tok := strings.Split(string(l), " ")
		if len(tok) != 3 {
			return nil, errors.New("Bad line")
		}
		//Strip newline
		tok[2] = tok[2][:len(tok[2])-1]
		switch tok[0] {
		case "kv":
			h := header{}
			h.Key = tok[1]
			cx, err := strconv.ParseUint(tok[2], 10, 32)
			if err != nil {
				return nil, err
			}
			h.ILength = int(cx)
			body := make([]byte, h.ILength)
			if _, e := io.ReadFull(s, body); e != nil {
				return nil, e
			}
			//Strip newline
			if _, e := s.ReadByte(); e != nil {
				return nil, e
			}
			h.Content = body
			f.Headers = append(f.Headers, h)
		case "ro":
			cx, err := strconv.ParseUint(tok[1], 10, 32)
			if err != nil {
				return nil, err
			}
			ronum := int(cx)
			cx, err = strconv.ParseUint(tok[2], 10, 32)
			if err != nil {
				return nil, err
			}
			length := int(cx)
			body := make([]byte, length)
			if _, e := io.ReadFull(s, body); e != nil {
				return nil, e
			}
			//Strip newline
			if _, e := s.ReadByte(); e != nil {
				return nil, e
			}
			ro, err := objects.LoadRoutingObject(ronum, body)
			if err != nil {
				return nil, e
			}
			f.ROs = append(f.ROs, roEntry{ro, strconv.Itoa(ronum), strconv.Itoa(length)})
		case "po":
			ponums := strings.Split(tok[1], ":")
			var ponum int
			if len(ponums[1]) != 0 {
				cx, err := strconv.ParseUint(ponums[1], 10, 32)
				if err != nil {
					return nil, err
				}
				ponum = int(cx)
			} else {
				cx, err := PONumFromDotForm(ponums[0])
				if err != nil {
					return nil, err
				}
				ponum = cx
			}
			cx, err = strconv.ParseUint(tok[2], 10, 32)
			if err != nil {
				return nil, err
			}
			length := int(cx)
			body := make([]byte, length)
			if _, e := io.ReadFull(s, body); e != nil {
				return nil, e
			}
			//Strip newline
			if _, e := s.ReadByte(); e != nil {
				return nil, e
			}
			poe := poEntry{
				PO:    body,
				PONum: ponum,
			}
			f.POs = append(f.POs, poe)
		case "end":
			return f, nil
		}
	}
}
