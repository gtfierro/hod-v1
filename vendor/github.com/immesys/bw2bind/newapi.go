package bw2bind

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"

	"github.com/immesys/bw2/crypto"
	"github.com/immesys/bw2/objects"
	"github.com/immesys/bw2bind/expr"
	"github.com/mgutz/ansi"
)

// PublishDOTWithAcc is like PublishDOT but allows you to specify the
// account you want to bankroll the operation
func (cl *BW2Client) PublishDOTWithAcc(blob []byte, account int) (string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdPutDot, seqno)
	//Strip first byte of blob, assuming it came from a file
	po := CreateBasePayloadObject(PONumROAccessDOT, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	hash, _ := fr.GetFirstHeader("hash")
	return hash, nil
}

// Publish the given DOT to the registry
func (cl *BW2Client) PublishDOT(blob []byte) (string, error) {
	return cl.PublishDOTWithAcc(blob, 0)
}

// Same as PublishEntity, but specify the account to use
func (cl *BW2Client) PublishEntityWithAcc(blob []byte, account int) (string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdPutEntity, seqno)
	po := CreateBasePayloadObject(PONumROEntity, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	vk, _ := fr.GetFirstHeader("vk")
	return vk, nil
}

func (cl *BW2Client) SetMetadata(uri, key, val string) error {
	po := CreateMetadataPayloadObject(&MetadataTuple{
		Value:     val,
		Timestamp: time.Now().UnixNano(),
	})
	uri = strings.TrimSuffix(uri, "/")
	uri += "/!meta/" + key
	return cl.Publish(&PublishParams{
		AutoChain:      true,
		PayloadObjects: []PayloadObject{po},
		URI:            uri,
		Persist:        true,
	})
}

func (cl *BW2Client) DevelopTrigger() {
	seqno := cl.GetSeqNo()
	req := createFrame("devl", seqno)
	<-cl.transact(req)
}

func (cl *BW2Client) DelMetadata(uri, key string) error {
	uri = strings.TrimSuffix(uri, "/")
	uri += "/!meta/" + key
	return cl.Publish(&PublishParams{
		AutoChain:      true,
		PayloadObjects: []PayloadObject{},
		URI:            uri,
		Persist:        true,
	})
}

func (cl *BW2Client) GetMetadata(uri string) (data map[string]*MetadataTuple,
	from map[string]string,
	err error) {
	uri = strings.TrimSuffix(uri, "/")
	type de struct {
		K string
		M *MetadataTuple
		O string
	}
	parts := strings.Split(uri, "/")
	chans := make([]chan de, len(parts))
	for i := 0; i < len(parts); i++ {
		chans[i] = make(chan de, 10)
	}
	var ge error
	for i := 0; i < len(parts); i++ {
		li := i
		go func() {
			turi := strings.Join(parts[:li+1], "/")
			smc, err := cl.Query(&QueryParams{
				AutoChain: true,
				URI:       turi + "/!meta/+",
			})
			if err != nil {
				close(chans[li])
				// Ignore permission errors on prefixes of the original URI
				if !(strings.HasPrefix(err.Error(), "[401]") && li < len(parts)-1) {
					ge = err
				}
				return
			}
			for sm := range smc {
				uriparts := strings.Split(sm.URI, "/")
				meta, ok := sm.GetOnePODF(PODFSMetadata).(MetadataPayloadObject)
				if ok {
					chans[li] <- de{uriparts[len(uriparts)-1], meta.Value(), turi}
				}
			}
			close(chans[li])
		}()
	}

	//		key -> de
	rvO := make(map[string]string)
	rvM := make(map[string]*MetadataTuple)

	//otherwise, iterate in forward order to prefer more specified metadata
	for i := 0; i < len(parts); i++ {
		for res := range chans[i] {
			rvO[res.K] = res.O
			rvM[res.K] = res.M
		}
	}

	//check error
	if ge != nil {
		return nil, nil, ge
	}
	return rvM, rvO, nil
}
func (cl *BW2Client) GetMetadataKey(uri, key string) (v *MetadataTuple, from string, err error) {
	uri = strings.TrimSuffix(uri, "/")
	parts := strings.Split(uri, "/")
	type de struct {
		K string
		M *MetadataTuple
		O string
	}
	chans := make([]chan *de, len(parts))
	for i := 0; i < len(parts); i++ {
		chans[i] = make(chan *de, 1)
	}
	var ge error
	wg := sync.WaitGroup{}
	wg.Add(len(parts))
	for i := 0; i < len(parts); i++ {
		li := i
		go func() {
			turi := strings.Join(parts[:li+1], "/")
			sm, err := cl.QueryOne(&QueryParams{
				AutoChain: true,
				URI:       turi + "/!meta/" + key,
			})
			if err != nil {
				ge = err
				wg.Done()
				return
			}
			if sm == nil {
				chans[li] <- nil
			} else {
				meta, ok := sm.GetOnePODF(PODFSMetadata).(MetadataPayloadObject)
				if ok {
					chans[li] <- &de{key, meta.Value(), turi}
				} else {
					chans[li] <- nil
				}
			}
			wg.Done()
		}()
	}
	//wait for queries to finish
	wg.Wait()
	//check error
	if ge != nil {
		return nil, "", ge
	}
	//otherwise, iterate in reverse order to prefer more specified metadata
	for i := len(parts) - 1; i >= 0; i-- {
		v := <-chans[i]
		if v != nil {
			return v.M, v.O, nil
		}
	}
	return nil, "", nil
}

// Print a line to stdout that depicts the local router status, typically
// used at the start of an interactive command
func (cl *BW2Client) StatLine() {
	cip, err := cl.GetBCInteractionParams()
	if err != nil {
		fmt.Printf("<statline err: %s>\n", err.Error())
		return
	}
	fmt.Printf("%s%s ╔╡%s%s %s\n%s ╚╡peers=%s%d%s block=%s%d%s age=%s%s%s\n",
		ansi.ColorCode("reset"),
		ansi.ColorCode("white"),
		cl.rHost,
		ansi.ColorCode("green+b"),
		cl.remotever,
		ansi.ColorCode("reset")+ansi.ColorCode("white"),
		ansi.ColorCode("blue+b"),
		cip.Peers,
		ansi.ColorCode("reset")+ansi.ColorCode("white"),
		ansi.ColorCode("blue+b"),
		cip.CurrentBlock,
		ansi.ColorCode("reset")+ansi.ColorCode("white"),
		ansi.ColorCode("blue+b"),
		cip.CurrentAge.String(),
		ansi.ColorCode("reset"))
}

func (cl *BW2Client) PublishEntity(blob []byte) (string, error) {
	return cl.PublishEntityWithAcc(blob, 0)
}
func (cl *BW2Client) PublishChainWithAcc(blob []byte, account int) (string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdPutChain, seqno)
	//TODO it might not be with a key...
	po := CreateBasePayloadObject(PONumROAccessDChain, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	hash, _ := fr.GetFirstHeader("hash")
	return hash, nil

}
func (cl *BW2Client) PublishChain(blob []byte) (string, error) {
	return cl.PublishChainWithAcc(blob, 0)
}
func (cl *BW2Client) UnresolveAlias(blob []byte) (string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdResolveAlias, seqno)
	req.AddHeaderB("unresolve", blob)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	v, _ := fr.GetFirstHeader("value")
	return v, nil
}
func (cl *BW2Client) ResolveLongAlias(al string) (data []byte, zero bool, err error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdResolveAlias, seqno)
	req.AddHeader("longkey", al)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return nil, false, err
	}
	v, _ := fr.GetFirstHeaderB("value")
	return v, bytes.Equal(v, make([]byte, 32)), nil
}
func (cl *BW2Client) ResolveShortAlias(al string) (data []byte, zero bool, err error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdResolveAlias, seqno)
	req.AddHeader("shortkey", al)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return nil, false, err
	}
	v, _ := fr.GetFirstHeaderB("value")
	return v, bytes.Equal(v, make([]byte, 32)), nil
}
func (cl *BW2Client) ResolveEmbeddedAlias(al string) (data string, err error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdResolveAlias, seqno)
	req.AddHeader("embedded", al)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	v, _ := fr.GetFirstHeader("value")
	return v, nil
}

type RegistryValidity int

const (
	StateUnknown = iota
	StateValid
	StateExpired
	StateRevoked
	StateError
)

func (cl *BW2Client) ValidityToString(i RegistryValidity, err error) string {
	switch i {
	case StateUnknown:
		return "UNKNOWN"
	case StateValid:
		return "valid"
	case StateExpired:
		return "EXPIRED"
	case StateRevoked:
		return "REVOKED"
	case StateError:
		return "ERROR: " + err.Error()
	}
	return "<WTF?>"
}

func (cl *BW2Client) ResolveRegistry(key string) (ro objects.RoutingObject, validity RegistryValidity, err error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdResolveRegistryObject, seqno)
	req.AddHeader("key", key)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, StateError, er
	}
	if len(fr.GetAllROs()) == 0 {
		return nil, StateUnknown, nil
	}
	ro = fr.GetAllROs()[0]
	err = nil
	valid, _ := fr.GetFirstHeader("validity")
	switch valid {
	case "valid":
		validity = StateValid
		return
	case "expired":
		validity = StateExpired
		return
	case "revoked":
		validity = StateRevoked
		return
	case "unknown":
		validity = StateUnknown
		return
	default:
		panic(valid)
	}
}
func (cl *BW2Client) FindDOTsFromVK(vk string) ([]*objects.DOT, []RegistryValidity, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdFindDots, seqno)
	req.AddHeader("vk", vk)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, nil, er
	}
	rvd := []*objects.DOT{}
	rvv := []RegistryValidity{}
	for _, po := range fr.POs {
		rpo, err := LoadPayloadObject(po.PONum, po.PO)
		if err != nil {
			return nil, nil, err
		}
		if po.PONum == PONumString {
			switch strings.ToLower(rpo.(TextPayloadObject).Value()) {
			case "valid":
				rvv = append(rvv, StateValid)
			case "expired":
				rvv = append(rvv, StateExpired)
			case "revoked":
				rvv = append(rvv, StateRevoked)
			case "unknown":
				rvv = append(rvv, StateUnknown)
			default:
				panic("unknown validity string: " + rpo.(TextPayloadObject).Value())
			}
		}
		if po.PONum == PONumROAccessDOT {
			doti, err := objects.NewDOT(po.PONum, rpo.(PayloadObject).GetContents())
			if err != nil {
				return nil, nil, err
			}
			rvd = append(rvd, doti.(*objects.DOT))
		}
	}
	return rvd, rvv, nil
}

type BalanceInfo struct {
	Addr    string
	Human   string
	Decimal string
	Int     *big.Int
}

func (cl *BW2Client) EntityBalances() ([]*BalanceInfo, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdEntityBalances, seqno)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, er
	}
	rv := make([]*BalanceInfo, 0, 16)
	for _, poe := range fr.POs {
		if poe.PONum == PONumAccountBalance {
			parts := strings.Split(string(poe.PO), ",")
			i := big.NewInt(0)
			i, _ = i.SetString(parts[1], 10)
			rv = append(rv, &BalanceInfo{
				Addr:    parts[0],
				Decimal: parts[1],
				Human:   parts[2],
				Int:     i,
			})
		}
	}
	return rv, nil
}
func (cl *BW2Client) AddressBalance(addr string) (*BalanceInfo, error) {
	if addr[0:2] == "0x" {
		addr = addr[2:]
	}
	if len(addr) != 40 {
		return nil, fmt.Errorf("Address must be 40 hex characters")
	}
	seqno := cl.GetSeqNo()
	req := createFrame(cmdAddressBalance, seqno)
	req.AddHeader("address", addr)
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, er
	}
	poe := fr.POs[0]
	parts := strings.Split(string(poe.PO), ",")
	i := big.NewInt(0)
	i, _ = i.SetString(parts[1], 10)
	rv := &BalanceInfo{
		Addr:    parts[0],
		Decimal: parts[1],
		Human:   parts[2],
		Int:     i,
	}
	return rv, nil
}

type BCIP struct {
	Confirmations *int64
	Timeout       *int64
	Maxage        *int64
}

type CurrentBCIP struct {
	Confirmations int64
	Timeout       int64
	Maxage        int64
	CurrentAge    time.Duration
	CurrentBlock  uint64
	Peers         int64
	HighestBlock  int64
	Difficulty    int64
}

func (cl *BW2Client) GetBCInteractionParams() (*CurrentBCIP, error) {
	return cl.SetBCInteractionParams(nil)
}
func (cl *BW2Client) SetBCInteractionParams(to *BCIP) (*CurrentBCIP, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdBCInteractionParams, seqno)
	if to != nil {
		if to.Confirmations != nil {
			req.AddHeader("confirmations", strconv.FormatInt(*to.Confirmations, 10))
		}
		if to.Timeout != nil {
			req.AddHeader("timeout", strconv.FormatInt(*to.Timeout, 10))
		}
		if to.Maxage != nil {
			req.AddHeader("maxage", strconv.FormatInt(*to.Maxage, 10))
		}
	}
	fr := <-cl.transact(req)
	if er := fr.MustResponse(); er != nil {
		return nil, er
	}
	rv := &CurrentBCIP{}
	v, _ := fr.GetFirstHeader("confirmations")
	iv, _ := strconv.ParseInt(v, 10, 64)
	rv.Confirmations = iv
	v, _ = fr.GetFirstHeader("timeout")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Timeout = iv
	v, _ = fr.GetFirstHeader("maxage")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Maxage = iv
	v, _ = fr.GetFirstHeader("currentblock")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.CurrentBlock = uint64(iv)
	v, _ = fr.GetFirstHeader("currentage")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.CurrentAge = time.Duration(iv) * time.Second
	v, _ = fr.GetFirstHeader("peers")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Peers = iv
	v, _ = fr.GetFirstHeader("highest")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.HighestBlock = iv
	v, _ = fr.GetFirstHeader("difficulty")
	iv, _ = strconv.ParseInt(v, 10, 64)
	rv.Difficulty = iv
	return rv, nil
}

type Currency int64

const KiloEther Currency = 1000 * Ether
const Ether Currency = 1000 * MilliEther
const MilliEther Currency = 1000 * MicroEther
const Finney Currency = 1000 * MicroEther
const MicroEther Currency = 1000 * NanoEther
const Szabo Currency = 1000 * NanoEther
const NanoEther Currency = 1
const GigaWei Currency = 1

func CurrencyToWei(v Currency) *big.Int {
	rv := big.NewInt(int64(v))
	rv = rv.Mul(rv, big.NewInt(1000000000))
	return rv
}

func (cl *BW2Client) TransferWei(from int, to string, wei *big.Int) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdTransfer, seqno)
	req.AddHeader("account", strconv.Itoa(from))
	req.AddHeader("address", to)
	req.AddHeader("valuewei", wei.Text(10))
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) TransferFrom(from int, to string, value Currency) error {
	return cl.TransferWei(from, to, CurrencyToWei(value))
}
func (cl *BW2Client) Transfer(to string, value Currency) error {
	return cl.TransferFrom(0, to, value)
}
func (cl *BW2Client) NewDesignatedRouterOffer(account int, nsvk string, dr *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdNewDROffer, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("nsvk", nsvk)
	if dr != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, dr.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) RevokeDesignatedRouterOffer(account int, nsvk string, dr *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdRevokeDROffer, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("nsvk", nsvk)
	if dr != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, dr.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) RevokeAcceptanceOfDesignatedRouterOffer(account int, drvk string, ns *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdRevokeDRAccept, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("drvk", drvk)
	if ns != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, ns.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}

// func (cl *BW2Client) RevokeDOT(account int, dothash string) (*objects.Revocation, error) {
//
// }
func (cl *BW2Client) RevokeEntity(vk string, comment string) (string, []byte, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdRevokeRO, seqno)
	req.AddHeader("entity", vk)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", nil, err
	}
	hash, _ := fr.GetFirstHeader("hash")
	po := fr.POs[0].PO
	return hash, po, nil
}
func (cl *BW2Client) RevokeDOT(hash string, comment string) (string, []byte, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdRevokeRO, seqno)
	req.AddHeader("dot", hash)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", nil, err
	}
	rhash, _ := fr.GetFirstHeader("hash")
	po := fr.POs[0].PO
	return rhash, po, nil
}
func (cl *BW2Client) PublishRevocation(account int, blob []byte) (string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdPutRevocation, seqno)
	//Strip first byte of blob, assuming it came from a file
	po := CreateBasePayloadObject(PONumRORevocation, blob)
	req.AddPayloadObject(po)
	req.AddHeader("account", strconv.Itoa(account))
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", err
	}
	hash, _ := fr.GetFirstHeader("hash")
	return hash, nil
}
func (cl *BW2Client) GetDesignatedRouterOffers(nsvk string) (active string, activesrv string, drvks []string, err error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdListDROffers, seqno)
	req.AddHeader("nsvk", nsvk)
	fr := <-cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return "", "", nil, err
	}
	rv := make([]string, 0)
	for _, po := range fr.POs {
		if po.PONum == objects.RODesignatedRouterVK {
			rv = append(rv, crypto.FmtKey(po.PO))
		}
	}
	act, _ := fr.GetFirstHeader("active")
	srv, _ := fr.GetFirstHeader("srv")
	return act, srv, rv, nil
}
func (cl *BW2Client) AcceptDesignatedRouterOffer(account int, drvk string, ns *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdAcceptDROffer, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("drvk", drvk)
	if ns != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, ns.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) SetDesignatedRouterSRVRecord(account int, srv string, dr *objects.Entity) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdUpdateSRVRecord, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeader("srv", srv)
	if dr != nil {
		po := CreateBasePayloadObject(objects.ROEntityWKey, dr.GetSigningBlob())
		req.AddPayloadObject(po)
	}
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) CreateLongAlias(account int, key []byte, val []byte) error {
	if len(key) > 32 || len(val) > 32 {
		return fmt.Errorf("Key and value must be shorter than 32 bytes")
	}
	seqno := cl.GetSeqNo()
	req := createFrame(cmdMakeLongAlias, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeaderB("content", val)
	req.AddHeaderB("key", key)
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) CreateShortAlias(account int, val []byte) (string, error) {
	if len(val) > 32 {
		return "", fmt.Errorf("Value must be shorter than 32 bytes")
	}
	seqno := cl.GetSeqNo()
	req := createFrame(cmdMakeShortAlias, seqno)
	req.AddHeader("account", strconv.Itoa(account))
	req.AddHeaderB("content", val)
	fe := <-cl.transact(req)
	if err := fe.MustResponse(); err != nil {
		return "", err
	}
	k, _ := fe.GetFirstHeader("hexkey")
	return k, nil
}
func (cl *BW2Client) Unsubscribe(handle string) error {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdUnsubscribe, seqno)
	req.AddHeader("handle", handle)
	return (<-cl.transact(req)).MustResponse()
}
func (cl *BW2Client) CreateView(expression expr.M) (*View, error) {
	mp, err := msgpack.Marshal(expression)
	seqno := cl.GetSeqNo()
	req := createFrame(cmdMakeView, seqno)
	req.AddHeaderB("msgpack", mp)
	rc := cl.transact(req)
	fr := <-rc
	if err := fr.MustResponse(); err != nil {
		return nil, err
	}
	vids, _ := fr.GetFirstHeader("id")
	vid, err := strconv.ParseUint(vids, 10, 64)
	if err != nil {
		return nil, err
	}
	rv := &View{vid: int(vid), cl: cl}
	go func() {
		for _ = range rc {
			rv.cbmu.Lock()
			for _, cb := range rv.cbz {
				cb()
			}
		}
	}()
	return rv, nil
}
func (v *View) OnChange(f func()) {
	v.cbmu.Lock()
	v.cbz = append(v.cbz, f)
	v.cbmu.Unlock()
}
func (v *View) List() ([]*InterfaceDescriptor, error) {
	rv := []*InterfaceDescriptor{}
	seqno := v.cl.GetSeqNo()
	req := createFrame(cmdListView, seqno)
	req.AddHeader("id", strconv.Itoa(v.vid))
	fr := <-v.cl.transact(req)
	if err := fr.MustResponse(); err != nil {
		return nil, err
	}
	for i := 0; i < fr.NumPOs(); i++ {
		po, _ := fr.GetPO(i)
		poz := po.(MsgPackPayloadObject)
		ifd := InterfaceDescriptor{}
		err := poz.ValueInto(&ifd)
		if err != nil {
			return nil, err
		}
		rv = append(rv, &ifd)
	}
	return rv, nil
}
func chToCB(ch chan *SimpleMessage, cb func(sm *SimpleMessage)) {
	go func() {
		for m := range ch {
			cb(m)
		}
	}()
}
func (v *View) PubSlot(iface, slot string, poz []PayloadObject) error {
	return v.pubSigSlot(iface, "slot", slot, poz)
}
func (v *View) PubSignal(iface, signal string, poz []PayloadObject) error {
	return v.pubSigSlot(iface, "signal", signal, poz)
}
func (v *View) pubSigSlot(iface, t, sigslot string, poz []PayloadObject) error {
	seqno := v.cl.GetSeqNo()
	req := createFrame(cmdPublishView, seqno)
	req.AddHeader("id", strconv.Itoa(v.vid))
	req.AddHeader(t, sigslot)
	req.AddHeader("iface", iface)
	for _, po := range poz {
		req.AddPayloadObject(po)
	}
	return (<-v.cl.transact(req)).MustResponse()
}
func (v *View) SubSlot(iface, slot string) (chan *SimpleMessage, error) {
	return v.subSigSlot(iface, "slot", slot)
}
func (v *View) SubSlotOrExit(iface, slot string) chan *SimpleMessage {
	rv, err := v.SubSlot(iface, slot)
	if err != nil {
		fmt.Println("View error in SubSlotOrExit:", err)
		os.Exit(1)
	}
	return rv
}
func (v *View) SubSlotF(iface, slot string, cb func(sm *SimpleMessage)) error {
	rv, err := v.SubSlot(iface, slot)
	chToCB(rv, cb)
	return err
}
func (v *View) SubSlotFOrExit(iface, slot string, cb func(sm *SimpleMessage)) {
	rv, err := v.SubSlot(iface, slot)
	if err != nil {
		fmt.Println("View error in SubSlotFOrExit:", err)
		os.Exit(1)
	}
	chToCB(rv, cb)
}
func (v *View) SubSignal(iface, signal string) (chan *SimpleMessage, error) {
	return v.subSigSlot(iface, "signal", signal)
}
func (v *View) SubSignalOrExit(iface, signal string) chan *SimpleMessage {
	rv, err := v.SubSignal(iface, signal)
	if err != nil {
		fmt.Println("View error in SubSignalOrExit:", err)
		os.Exit(1)
	}
	return rv
}
func (v *View) SubSignalF(iface, signal string, cb func(sm *SimpleMessage)) error {
	rv, err := v.SubSignal(iface, signal)
	chToCB(rv, cb)
	return err
}
func (v *View) SubSignalFOrExit(iface, signal string, cb func(sm *SimpleMessage)) {
	rv, err := v.SubSignal(iface, signal)
	chToCB(rv, cb)
	if err != nil {
		fmt.Println("View error in SubSignalFOrExit:", err)
		os.Exit(1)
	}
}
func (v *View) subSigSlot(iface, t, sigslot string) (chan *SimpleMessage, error) {
	seqno := v.cl.GetSeqNo()
	req := createFrame(cmdSubscribeView, seqno)
	req.AddHeader("id", strconv.Itoa(v.vid))
	req.AddHeader(t, sigslot)
	req.AddHeader("iface", iface)
	rsp := v.cl.transact(req)
	//First response is the RESP frame
	fr, _ := <-rsp
	err := fr.MustResponse()
	if err != nil {
		return nil, err
	}
	//Generate converted output channel
	rv := make(chan *SimpleMessage, 10)
	go func() {
		for f := range rsp {
			sm := SimpleMessage{}
			sm.From, _ = f.GetFirstHeader("from")
			sm.URI, _ = f.GetFirstHeader("uri")
			sm.ROs = f.GetAllROs()
			poslice := make([]PayloadObject, f.NumPOs())
			errslice := make([]error, 0)
			for i := 0; i < f.NumPOs(); i++ {
				var err error
				poslice[i], err = f.GetPO(i)
				if err != nil {
					errslice = append(errslice, err)
				}
			}
			sm.POs = poslice
			sm.POErrors = errslice
			rv <- &sm
		}
		close(rv)
	}()
	return rv, nil
}
