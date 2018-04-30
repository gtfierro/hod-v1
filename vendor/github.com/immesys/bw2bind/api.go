package bw2bind

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	log "github.com/cihub/seelog"
	"github.com/immesys/bw2/objects"
)

// SilenceLog will redirect the log output typically emitted by bw2bind to
// a file (.bw2bind.log) in the working directory. It is useful for interactive
// applications that do not wish log output to interfere.
func SilenceLog() {
	cfg := `
	<seelog>
    <outputs>
        <splitter formatid="common">
            <file path=".bw2bind.log"/>
        </splitter>
    </outputs>
		<formats>
				<format id="common" format="[%LEV] %Time %Date %File:%Line %Msg%n"/>
		</formats>
	</seelog>`
	nlogger, err := log.LoggerFromConfigAsString(cfg)
	if err == nil {
		log.ReplaceLogger(nlogger)
	} else {
		fmt.Printf("Bad log config: %v\n", err)
		os.Exit(1)
	}
}

// OverrideAutoChainTo(v) will set the value of the AutoChain parameter to v
// for any subsequent Publish or Consume operations, it exists purely as a
// convenience. Do note that even if AutoChain is specified for these operations,
// this setting will always override it.
func (cl *BW2Client) OverrideAutoChainTo(v bool) {
	cl.defAutoChain = &v
}

// This enables the WAL for persistent publishing. The WAL is created at the given
// directory path. To use the WAL, set the EnsureDelivery flag in PublishParams to true.
// Calling EnableWAL also starts asynchronously re-publishing pending messages to the
// designated router
func (cl *BW2Client) EnableWAL(dir string) error {
	var err error
	cl.wal, err = newWal(dir)
	if err != nil {
		log.Critical(err)
	}

	// replay the pending messages
	go func() {
		topublish, err := cl.wal.pending()
		if err != nil {
			log.Error("error replay", err)
		}
		for _, pp := range topublish {
			pp.Persist = false
			if err := cl.Publish(pp); err != nil {
				log.Error("error replay", err)
			}
		}
	}()
	return nil
}

// ClearAutoChainOverride will remove the AutoChain override, allowing the
// per-call AutoChain setting to take effect.
func (cl *BW2Client) ClearAutoChainOverride() {
	cl.defAutoChain = nil
}

// Connect will connect to a BOSSWAVE local router. If "to" is the empty
// string, it will default to $BW2_AGENT if set else localhost:28589
func Connect(to string) (*BW2Client, error) {
	if to == "" {
		to = os.Getenv("BW2_AGENT")
		if to == "" {
			to = "localhost:28589"
		}
	}
	_, _, err := net.SplitHostPort(to)
	if err != nil && err.Error() == "missing port in address" {
		to = to + ":28589"
		_, _, err = net.SplitHostPort(to)
	}
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", to)
	if err != nil {
		return nil, err
	}
	rv := &BW2Client{c: conn,
		out:    bufio.NewWriter(conn),
		in:     bufio.NewReader(conn),
		seqnos: make(map[int]chan *frame),
		rHost:  to,
	}

	//As a bit of a sanity check, we read the first frame, which is the
	//server HELO message
	ok := make(chan bool, 1)
	go func() {
		helo, err := loadFrameFromStream(rv.in)
		if err != nil {
			log.Error("Malformed HELO frame: ", err)
			ok <- false
			return
		}
		if helo.Cmd != cmdHello {
			log.Error("frame not HELO")
			ok <- false
			return
		}
		rver, hok := helo.GetFirstHeader("version")
		if !hok {
			log.Error("frame has no version")
			ok <- false
			return
		}
		rv.remotever = rver
		log.Info("Connected to BOSSWAVE router version ", rver)
		ok <- true
	}()
	go func() {
		time.Sleep(10 * time.Second)
		for {
			time.Sleep(10 * time.Second)
			rv.olock.Lock()
			/*
				fmt.Printf("DEBUG IN BW2BIND: OPEN SEQNOS: ")
				for k := range rv.seqnos {
					fmt.Printf(" - %d\n", k)
				}
			*/
			rv.olock.Unlock()
		}
	}()

	select {
	case okv := <-ok:
		if okv {
			//Reader:
			go func() {
				for {
					frame, err := loadFrameFromStream(rv.in)
					if err != nil {
						log.Error("Invalid frame")
						log.Flush()
						if rv.errorHandler != nil {
							rv.errorHandler(errors.New("Invalid frame"))
							return
						}
						os.Exit(1)
					}
					rv.olock.Lock()
					dest, ok := rv.seqnos[frame.SeqNo]
					rv.olock.Unlock()
					if ok {
						dest <- frame
					}
				}
			}()
			return rv, nil
		}
		return nil, errors.New("Bad router")
	case _ = <-time.After(5 * time.Second):
		log.Error("Timeout on router HELO")
		conn.Close()
		return nil, errors.New("Timeout on HELO")
	}
}

// ConnectOrExit is the same as Connect but will
// print an error message to stderr and exit the program if the connection
// fails
func ConnectOrExit(to string) *BW2Client {
	bw, err := Connect(to)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to local BW2 router:", err.Error())
		os.Exit(1)
	}
	return bw
}

// CreateEntity will create a new entity and return the verifying key and the
// binary representation
func (cl *BW2Client) CreateEntity(p *CreateEntityParams) (string, []byte, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdMakeEntity, seqno)
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("contact", p.Contact)
	req.AddHeader("comment", p.Comment)
	for _, rvk := range p.Revokers {
		req.AddHeader("revoker", rvk)
	}
	if p.OmitCreationDate {
		req.AddHeader("omitcreationdate", "true")
	}
	rsp := cl.transact(req)
	fr, _ := <-rsp
	err := fr.MustResponse()
	if err != nil {
		return "", nil, err
	}
	if len(fr.POs) != 1 {
		return "", nil, errors.New("bad response")
	}
	vk, _ := fr.GetFirstHeader("vk")
	po := fr.POs[0].PO

	return vk, po, nil
}

// CreateDOT will create a new Declaration of Trust and return the
// DOT Hash and the binary representation
func (cl *BW2Client) CreateDOT(p *CreateDOTParams) (string, []byte, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdMakeDot, seqno)
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("contact", p.Contact)
	req.AddHeader("comment", p.Comment)
	for _, rvk := range p.Revokers {
		req.AddHeader("revoker", rvk)
	}
	if p.OmitCreationDate {
		req.AddHeader("omitcreationdate", "true")
	}
	req.AddHeader("ttl", strconv.Itoa(int(p.TTL)))
	req.AddHeader("to", p.To)
	req.AddHeader("ispermission", strconv.FormatBool(p.IsPermission))
	if !p.IsPermission {
		req.AddHeader("uri", p.URI)
		req.AddHeader("accesspermissions", p.AccessPermissions)
	} else {
		panic("Not supported yet")
	}
	rsp := cl.transact(req)
	fr, _ := <-rsp
	err := fr.MustResponse()
	if err != nil {
		return "", nil, err
	}
	if len(fr.POs) != 1 {
		return "", nil, errors.New("bad response")
	}
	hash, _ := fr.GetFirstHeader("hash")
	po := fr.POs[0].PO

	return hash, po, nil
}

// CreateDOTChain will manually create a DOT chain. This is not a commonly
// used method, as typically you will rely on the local router to create the
// the chain, either via specifying AutoChain = true, or by using BuildChain
func (cl *BW2Client) CreateDOTChain(p *CreateDotChainParams) (string, *objects.DChain, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdMakeChain, seqno)
	req.AddHeader("ispermission", strconv.FormatBool(p.IsPermission))
	req.AddHeader("unelaborate", strconv.FormatBool(p.UnElaborate))
	for _, dot := range p.DOTs {
		req.AddHeader("dot", dot)
	}
	rsp := cl.transact(req)
	fr, _ := <-rsp
	err := fr.MustResponse()
	if err != nil {
		return "", nil, err
	}
	if len(fr.ROs) != 1 {
		return "", nil, errors.New("bad response")
	}
	hash, _ := fr.GetFirstHeader("hash")
	ro := fr.ROs[0].RO

	return hash, ro.(*objects.DChain), nil
}

// PublishOrExit is the same as Publish but will print an error message and
// exit the program if the operation does not succeed.
func (cl *BW2Client) PublishOrExit(p *PublishParams) {
	e := cl.Publish(p)
	if e != nil {
		fmt.Println("Could not publish:", e)
		os.Exit(1)
	}
}

// Publish sends a message with the given PublishParams, for example:
//  client.Publish(&bw2bind.PublishParams{
//			URI : name.space/my/path,
//			AutoChain : true,
//			PayloadObjects : myPoSlice,
//	})
func (cl *BW2Client) Publish(p *PublishParams) error {
	seqno := cl.GetSeqNo()
	cmd := cmdPublish
	if p.Persist {
		cmd = cmdPersist
	}
	var (
		hash []byte
		err  error
	)

	// add this message to the WAL
	if p.EnsureDelivery && cl.wal == nil {
		return errors.New("Need to call client.EnableWAL(<directory>) to ensure delivery")
	}
	if p.EnsureDelivery {
		hash, err = cl.wal.add(*p)
		if err != nil {
			return err
		}
	}

	req := createFrame(cmd, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}

	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	for _, po := range p.PayloadObjects {
		req.AddPayloadObject(po)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaboratePartial
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	req.AddHeader("persist", strconv.FormatBool(p.Persist))
	rsp := cl.transact(req)
	fr, _ := <-rsp

	// mark this message as sent in the WAL
	if p.EnsureDelivery {
		if err := cl.wal.done(hash); err != nil {
			return err
		}
	}
	err = fr.MustResponse()
	return err
}

// SubscribeOrExit is just like subscribe but will print an error message
// and exit the program if the operation does not succeed
func (cl *BW2Client) SubscribeOrExit(p *SubscribeParams) chan *SimpleMessage {
	rv, err := cl.Subscribe(p)
	if err == nil {
		return rv
	}
	fmt.Println("Could not subscribe:", err)
	os.Exit(1)
	return nil
}

func (cl *BW2Client) Subscribe(p *SubscribeParams) (chan *SimpleMessage, error) {
	ch, _, e := cl.SubscribeH(p)
	return ch, e
}

// Subscribe will consume a URI specified by SubscribeParams, it returns a
// channel that received messages will be written to, a handle that can be
// passed to unsubscribe, and an error
func (cl *BW2Client) SubscribeH(p *SubscribeParams) (chan *SimpleMessage, string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdSubscribe, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}
	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaboratePartial
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	if !p.LeavePacked {
		req.AddHeader("unpack", "true")
	}
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	rsp := cl.transact(req)
	//First response is the RESP frame
	fr, _ := <-rsp
	err := fr.MustResponse()

	if err != nil {
		return nil, "", err
	}
	handle, _ := fr.GetFirstHeader("handle")
	//Generate converted output channel
	rv := make(chan *SimpleMessage, 10)
	go func() {
		for f := range rsp {
			sm := SimpleMessage{}
			sm.From, _ = f.GetFirstHeader("from")
			sm.URI, _ = f.GetFirstHeader("uri")
			sigh, ok := f.GetFirstHeader("signature")
			if ok {
				rv, err := base64.URLEncoding.DecodeString(sigh)
				if err == nil && len(rv) == 64 {
					sm.Signature = rv
				}
			}
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
	return rv, handle, nil
}

// SetEntity will tell your local router "who you are". This is the
// entity that will be used to sign messages. It takes a binary
// representation of an Entity, as obtained from objects.Entity.GetSigningBlob.
// Note that if you have read an on-disk entity file (as made by bw2 mke),
// those routing object files have a one byte type header that must be stripped.
// Consider using SetEntityFile. This operation returns the entity's VK
func (cl *BW2Client) SetEntity(keyfile []byte) (vk string, err error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdSetEntity, seqno)
	po := CreateBasePayloadObject(objects.ROEntityWKey, keyfile)
	req.AddPayloadObject(po)
	rsp := cl.transact(req)
	fr, _ := <-rsp
	err = fr.MustResponse()
	if err != nil {
		return "", err
	}
	vk, _ = fr.GetFirstHeader("vk")
	return vk, nil
}

// SetEntityOrExit is the same as SetEntity but will print an error message
// and exit if the operation fails.
func (cl *BW2Client) SetEntityOrExit(keyfile []byte) (vk string) {
	rv, e := cl.SetEntity(keyfile)
	if e != nil {
		fmt.Fprintln(os.Stderr, "Could not set entity :", e.Error())
		os.Exit(1)
	}
	return rv
}

// SetEntityFileOrExit is the same as SetEntityOrExit but reads the entity
// contents from the given file
func (cl *BW2Client) SetEntityFileOrExit(filename string) (vk string) {
	rv, e := cl.SetEntityFile(filename)
	if e != nil {
		fmt.Fprintln(os.Stderr, "Could not set entity file:", e.Error())
		os.Exit(1)
	}
	return rv
}

// SetEntityFromEnvironOrExit is the same as SetEntityFileOrExit
// but loads the file name from the BW2_DEFAULT_ENTITY environment
// variable
func (cl *BW2Client) SetEntityFromEnvironOrExit() (vk string) {
	fname := os.Getenv("BW2_DEFAULT_ENTITY")
	if fname == "" {
		fmt.Fprintln(os.Stderr, "$BW2_DEFAULT_ENTITY not set")
		os.Exit(1)
	}
	return cl.SetEntityFileOrExit(fname)
}

// SetEntityFile is the same as SetEntity but reads the entity contents
// from the given file
func (cl *BW2Client) SetEntityFile(filename string) (vk string, err error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return cl.SetEntity(contents[1:])
}

// BuildChain will ask the local router to find all chains granting permissions
// to the given VK. It returns a channel that the chains will be written to.
// This is a poweruser method, consider using BuildAnyChain or simply AutoChain
func (cl *BW2Client) BuildChain(uri, permissions, to string) (chan *SimpleChain, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdBuildChain, seqno)
	req.AddHeader("uri", uri)
	req.AddHeader("to", to)
	req.AddHeader("accesspermissions", permissions)
	rv := make(chan *SimpleChain, 2)
	rsp := cl.transact(req)
	proc := func() {
		for fr := range rsp {
			hash, _ := fr.GetFirstHeader("hash")
			if hash != "" {
				permissions, _ := fr.GetFirstHeader("permissions")
				to, _ := fr.GetFirstHeader("to")
				uri, _ := fr.GetFirstHeader("uri")
				sc := SimpleChain{
					Hash:        hash,
					Permissions: permissions,
					To:          to,
					URI:         uri,
					Content:     fr.POs[0].PO,
				}
				rv <- &sc
			}

		}
		close(rv)
	}
	fr, _ := <-rsp
	err := fr.MustResponse()
	if err != nil {
		return nil, err
	}
	go proc()
	return rv, nil
}

// BuildAnyChainOrExit is like BuildAnyChain but will print an error message and
// exit the program if the operation fails
func (cl *BW2Client) BuildAnyChainOrExit(uri, permissions, to string) *SimpleChain {
	rv, e := cl.BuildAnyChain(uri, permissions, to)
	if e != nil || rv == nil {
		fmt.Fprintf(os.Stderr, "Could not build chain to %s granting %s: %s", uri, permissions, e.Error())
		os.Exit(1)
	}
	return rv
}

// BuildAnyChain is a convenience function that calls BuildChain and only returns
// the first result, or nil if no chains were found
func (cl *BW2Client) BuildAnyChain(uri, permissions, to string) (*SimpleChain, error) {
	rc, err := cl.BuildChain(uri, permissions, to)
	if err != nil {
		return nil, err
	}
	rv, ok := <-rc
	if ok {
		go func() {
			for _ = range rc {
			}
		}()
		return rv, nil
	}
	return nil, errors.New("No result")
}

// QueryOneOrExit is like QueryOne but prints an error message and exits if
// the operation does not succeed
func (cl *BW2Client) QueryOneOrExit(p *QueryParams) *SimpleMessage {
	rv, err := cl.QueryOne(p)
	if err != nil {
		fmt.Printf("Could not query: %v\n", err)
		os.Exit(1)
	}
	return rv
}

// QueryOne calls Query but only returns the first result
func (cl *BW2Client) QueryOne(p *QueryParams) (*SimpleMessage, error) {
	rvc, err := cl.Query(p)
	if err != nil {
		return nil, err
	}
	v, ok := <-rvc
	if !ok {
		return nil, nil
	}
	go func() {
		for _ = range rvc {
		}
	}()
	return v, nil
}

// QueryOrExit is like Query but will print an error message and exit if the
// operation fails
func (cl *BW2Client) QueryOrExit(p *QueryParams) chan *SimpleMessage {
	rv, e := cl.Query(p)
	if e != nil {
		fmt.Println("Could not query:", e)
		os.Exit(1)
	}
	return rv
}

// Query will return all persisted messages matching the QueryParams e.g.
//  client.Query(&bw2bind.QueryParams{
//		URI : "name.space/my/uri",
//		AutoChain: true,
//	})
func (cl *BW2Client) Query(p *QueryParams) (chan *SimpleMessage, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdQuery, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}
	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaboratePartial
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	if !p.LeavePacked {
		req.AddHeader("unpack", "true")
	}
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	rsp := cl.transact(req)
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
			var ok bool
			sm.From, ok = f.GetFirstHeader("from")
			if !ok {
				continue
			}
			sm.URI, _ = f.GetFirstHeader("uri")
			sigh, ok := f.GetFirstHeader("signature")
			if ok {
				rv, err := base64.URLEncoding.DecodeString(sigh)
				if err == nil && len(rv) == 64 {
					sm.Signature = rv
				}
			}
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

// List will list all immediate children of the URI specified in ListParams,
// as long as one of the URIs under that child has a persisted message
func (cl *BW2Client) List(p *ListParams) (chan string, error) {
	seqno := cl.GetSeqNo()
	req := createFrame(cmdQuery, seqno)
	if cl.defAutoChain != nil {
		p.AutoChain = *cl.defAutoChain
	}
	if p.AutoChain {
		req.AddHeader("autochain", "true")
	}
	if p.Expiry != nil {
		req.AddHeader("expiry", p.Expiry.Format(time.RFC3339))
	}
	if p.ExpiryDelta != nil {
		req.AddHeader("expirydelta", p.ExpiryDelta.String())
	}
	req.AddHeader("uri", p.URI)
	if len(p.PrimaryAccessChain) != 0 {
		req.AddHeader("primary_access_chain", p.PrimaryAccessChain)
	}
	for _, ro := range p.RoutingObjects {
		req.AddRoutingObject(ro)
	}
	if p.ElaboratePAC == "" {
		p.ElaboratePAC = ElaboratePartial
	}
	req.AddHeader("elaborate_pac", p.ElaboratePAC)
	req.AddHeader("doverify", strconv.FormatBool(!p.DoNotVerify))
	rsp := cl.transact(req)
	//First response is the RESP frame
	fr, ok := <-rsp
	if ok {
		status, _ := fr.GetFirstHeader("status")
		if status != "okay" {
			msg, _ := fr.GetFirstHeader("reason")
			return nil, errors.New(msg)
		}
	} else {
		return nil, errors.New("receive channel closed")
	}
	//Generate converted output channel
	rv := make(chan string, 10)
	go func() {
		for f := range rsp {
			child, _ := f.GetFirstHeader("child")
			rv <- child
		}
		close(rv)
	}()
	return rv, nil
}

func (cl *BW2Client) GetSeqNo() int {
	newseqno := atomic.AddUint32(&cl.curseqno, 1)
	return int(newseqno)
}

// ToBase64 is a utility function to convert a binary representation of a VK or
// a hash into the 44-character base64 representation
func ToBase64(key []byte) string {
	return base64.URLEncoding.EncodeToString(key)
}

// FromBase64 is a utility function to convert a 44-character representation of
// a hash or VK into the binary representation. It returns an error if the
// result is not 32 bytes or the key is not base64
func FromBase64(key string) ([]byte, error) {
	rv, err := base64.URLEncoding.DecodeString(key)
	if len(rv) != 32 {
		return nil, errors.New("Invalid length")
	}
	return rv, err
}
