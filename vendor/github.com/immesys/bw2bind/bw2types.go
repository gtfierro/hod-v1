package bw2bind

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/immesys/bw2/objects"
	"gopkg.in/vmihailenco/msgpack.v3"
)

// ElaborateDefault is the same as ElaboratePartial
const ElaborateDefault = ""

// ElaborateFull will copy the entire contents of every DOT in the dot chain
// into the message. This used to be useful before BW 2.1.x because it guaranteed
// that the receiver could verify the message. We now always have that guarantee
// so this is wasteful
const ElaborateFull = "full"

// ElaboratePartial will include the DOT hashes of all DOTs in the primary access
// chain in the message. This is required if the DOT chain is not published to
// the registry. This is the default elaboration option.
const ElaboratePartial = "partial"

// ElaborateNone only sends the DChain hash in the message. This will only work
// if the DChain is published to the registry.
const ElaborateNone = "none"

// PublishParams is used for Publish and Persist messages
type PublishParams struct {
	// The URI you wish to publish to
	URI string
	// The PrimaryAccessChain hash, if you are manually specifying the chain.
	// as of 2.1.x this is not recommended as it will not work unless the chain
	// is published to the registry
	PrimaryAccessChain string
	// Tell the local router to build the chain for you, if you always set
	// this value, consider using BW2Client.OverrideAutoChainTo()
	AutoChain bool
	// The routing objects to include in the message, this is not commonly used
	RoutingObjects []objects.RoutingObject
	// The payload objects to include in the message.
	PayloadObjects []PayloadObject
	// The expiry date of this message. Note that routers will reject messages
	// that arrive after this time
	Expiry *time.Time
	// Same as expiry but expressed from now
	ExpiryDelta *time.Duration
	// The PAC elaboration level to use, defaults to ElaboratePartial
	ElaboratePAC string
	// By default the local router will verify the message before sending, setting
	// this to true will disable this stage
	DoNotVerify bool
	// Do you want the message to be persist on the designated router
	Persist bool
	// Do you want the message delivery to the designated router to be guaranteed
	// This gets persisted in a local WAL. Setting this to true automatically sets
	// the Persist flag to false.
	EnsureDelivery bool
}

func (pp PublishParams) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.Encode(pp.URI)
	if err != nil {
		return err
	}
	err = enc.Encode(len(pp.PayloadObjects))
	if err != nil {
		return err
	}
	for _, po := range pp.PayloadObjects {
		err = enc.Encode(po.GetPONum())
		if err != nil {
			return err
		}
		err = enc.Encode(po.GetContents())
		if err != nil {
			return err
		}
	}
	return err
}

func (pp *PublishParams) DecodeMsgpack(dec *msgpack.Decoder) error {
	err := dec.Decode(&pp.URI)
	if err != nil {
		return err
	}
	var num int
	var ponum int
	var contents []byte
	err = dec.Decode(&num)
	if err != nil {
		return err
	}
	for i := 0; i < num; i++ {
		if err := dec.Decode(&ponum, &contents); err != nil {
			return err
		}
		po, err := LoadPayloadObject(ponum, contents)
		if err != nil {
			return err
		}
		pp.PayloadObjects = append(pp.PayloadObjects, po)
	}

	return nil
}

func (po *PayloadObjectImpl) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(po.GetPONum(), po.GetContents())
}
func (po *PayloadObjectImpl) DecodeMsgpack(dec *msgpack.Decoder) error {
	var ponum int
	var contents []byte
	if err := dec.Decode(&ponum, &contents); err != nil {
		return err
	}
	var err error
	po, err = LoadBasePayloadObject(ponum, contents)
	return err
}

type SubscribeParams struct {
	// The URI you wish to subscribe to
	URI string
	// The PrimaryAccessChain hash, if you are manually specifying the chain.
	// as of 2.1.x this is not recommended as it will not work unless the chain
	// is published to the registry
	PrimaryAccessChain string
	// Tell the local router to build the chain for you, if you always set
	// this value, consider using BW2Client.OverrideAutoChainTo()
	AutoChain bool
	// The routing objects to include in the message, this is not commonly used
	RoutingObjects []objects.RoutingObject
	// The expiry date of this message. Note that routers will reject messages
	// that arrive after this time. For subscribe this is essentially used
	// for replay-protection in very corner-case attack vectors
	Expiry *time.Time
	// Same as expiry but expressed from now
	ExpiryDelta *time.Duration
	// The PAC elaboration level to use, defaults to ElaboratePartial
	ElaboratePAC string
	// By default the local router will verify the message before sending, setting
	// this to true will disable this stage
	DoNotVerify bool
	// By default, the local router will take incoming messages and decompose
	// them into PO's and RO's. If you want the message to remain packed in
	// signed bosswave format, set this to true
	LeavePacked bool
}
type ListParams struct {
	// The URI you wish to list the children of
	URI string
	// The PrimaryAccessChain hash, if you are manually specifying the chain.
	// as of 2.1.x this is not recommended as it will not work unless the chain
	// is published to the registry
	PrimaryAccessChain string
	// Tell the local router to build the chain for you, if you always set
	// this value, consider using BW2Client.OverrideAutoChainTo()
	AutoChain bool
	// The routing objects to include in the message, this is not commonly used
	RoutingObjects []objects.RoutingObject
	// The expiry date of this message. Note that routers will reject messages
	// that arrive after this time. For subscribe this is essentially used
	// for replay-protection in very corner-case attack vectors
	Expiry *time.Time
	// Same as expiry but expressed from now
	ExpiryDelta *time.Duration
	// The PAC elaboration level to use, defaults to ElaboratePartial
	ElaboratePAC string
	// By default the local router will verify the message before sending, setting
	// this to true will disable this stage
	DoNotVerify bool
}
type QueryParams struct {
	// The URI you wish to query
	URI string
	// The PrimaryAccessChain hash, if you are manually specifying the chain.
	// as of 2.1.x this is not recommended as it will not work unless the chain
	// is published to the registry
	PrimaryAccessChain string
	// Tell the local router to build the chain for you, if you always set
	// this value, consider using BW2Client.OverrideAutoChainTo()
	AutoChain bool
	// The routing objects to include in the message, this is not commonly used
	RoutingObjects []objects.RoutingObject
	// The expiry date of this message. Note that routers will reject messages
	// that arrive after this time. For subscribe this is essentially used
	// for replay-protection in very corner-case attack vectors
	Expiry *time.Time
	// Same as expiry but expressed from now
	ExpiryDelta *time.Duration
	// The PAC elaboration level to use, defaults to ElaboratePartial
	ElaboratePAC string
	// By default the local router will verify the message before sending, setting
	// this to true will disable this stage
	DoNotVerify bool
	// By default, the local router will take incoming messages and decompose
	// them into PO's and RO's. If you want the message to remain packed in
	// signed bosswave format, set this to true
	LeavePacked bool
}
type CreateDOTParams struct {
	// Is this a permission DOT (hope not, they are not supported yet)
	IsPermission bool
	// The VK to grant the DOT to (from comes from BW2Client.SetEntity)
	To string
	// The time to live
	TTL uint8
	// The expiry time
	Expiry *time.Time
	// Same as Expiry but specified from now
	ExpiryDelta *time.Duration
	// The contact information of the DOT, typically "Name Surname <email@address.net>"
	Contact string
	// The comment information in the dot
	Comment string
	// The entities that are allowed to revoke this DOT
	Revokers []string
	// Leave the creation date out of the DOT (this is not normally used)
	OmitCreationDate bool

	// For Access DOTs, the URI to grant on
	URI string
	// For Access DOTs, the ADPS permissions e.g LPC*
	AccessPermissions string

	//For Permissions DOTs, don't use this yet
	AppPermissions map[string]string
}
type CreateDotChainParams struct {
	DOTs         []string
	IsPermission bool
	UnElaborate  bool
}
type CreateEntityParams struct {
	Expiry      *time.Time
	ExpiryDelta *time.Duration
	Contact     string
	Comment     string
	// The entities that will be allowed to revoke this entity
	Revokers         []string
	OmitCreationDate bool
}
type BuildChainParams struct {
	// The URI you wish to build a chain for
	URI string
	// The ADPS permissions you need (e.g. "LPC*")
	Permissions string
	// The VK to grant to
	To string
}

type SimpleMessage struct {
	From      string
	URI       string
	POs       []PayloadObject
	ROs       []objects.RoutingObject
	POErrors  []error
	Signature []byte
}
type SimpleChain struct {
	Hash        string
	Permissions string
	URI         string
	To          string
	Content     []byte
}

// Dump a given message to the console, deconstructing it as much as possible
func (sm *SimpleMessage) Dump() {
	fmt.Printf("Message from %s on %s:\n", sm.From, sm.URI)
	for _, po := range sm.POs {
		fmt.Println(po.TextRepresentation())
	}
}

// PONumDotForm turns an integer Payload Object number into dotted quad form
func PONumDotForm(ponum int) string {
	return fmt.Sprintf("%d.%d.%d.%d", ponum>>24, (ponum>>16)&0xFF, (ponum>>8)&0xFF, ponum&0xFF)
}

// PONumFromDotForm turns a dotted quad form into an integer Payload Object number
func PONumFromDotForm(dotform string) (int, error) {
	parts := strings.Split(dotform, ".")
	if len(parts) != 4 {
		return 0, errors.New("Bad dotform")
	}
	rv := 0
	for i := 0; i < 4; i++ {
		cx, err := strconv.ParseUint(parts[i], 10, 8)
		if err != nil {
			return 0, err
		}
		rv += (int(cx)) << uint(((3 - i) * 8))
	}
	return rv, nil
}

// FromDotForm is a shortcut for PONumFromDotForm that panics
// if there is an error
func FromDotForm(dotform string) int {
	rv, err := PONumFromDotForm(dotform)
	if err != nil {
		panic(err)
	}
	return rv
}

// GetOnePODF -Get a single Payload Object of the given Dot Form
// returns nil if there are none that match
func (sm *SimpleMessage) GetOnePODF(df string) PayloadObject {
	for _, p := range sm.POs {
		if p.IsTypeDF(df) {
			return p
		}
	}
	return nil
}

type View struct {
	vid  int
	cl   *BW2Client
	cbz  []func()
	cbmu sync.Mutex
}

type InterfaceDescriptor struct {
	URI       string            `msgpack:"uri"`
	Interface string            `msgpack:"iface"`
	Service   string            `msgpack:"svc"`
	Namespace string            `msgpack:"namespace"`
	Prefix    string            `msgpack:"prefix"`
	Suffix    string            `msgpack:"suffix"`
	Metadata  map[string]string `msgpack:"metadata"`
	v         *View
}
