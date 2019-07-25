//go:generate msgp
//msgp:ignore Parser
package turtle

// #cgo CFLAGS: -I ../raptor/src
// #cgo LDFLAGS: -lraptor2
// #include <stdio.h>
// #include <raptor2.h>
import "C"
import (
	"io"
	"io/ioutil"
	//"log"
	"os"
	"strings"
	"sync"
	"time"
)

var p *Parser

type Parser struct {
	dataset *DataSet
	sync.Mutex
}

type URI struct {
	Namespace string `msg:"n"`
	Value     string `msg:"v"`
}

func (u URI) String() string {
	if u.Namespace != "" {
		return u.Namespace + "#" + u.Value
	}
	return u.Value
}

func (u URI) Bytes() []byte {
	if u.Namespace != "" {
		return []byte(u.Namespace + "#" + u.Value)
	}
	return []byte(u.Value)
}

func (u URI) IsVariable() bool {
	return strings.HasPrefix(u.Value, "?")
}

func (u URI) IsEmpty() bool {
	return len(u.Namespace) == 0 && len(u.Value) == 0
}

func ParseURI(uri string) URI {
	uri = strings.TrimLeft(uri, "<")
	uri = strings.TrimRight(uri, ">")
	parts := strings.Split(uri, "#")
	parts[0] = strings.TrimRight(parts[0], "#")
	if len(parts) != 2 {
		if strings.HasPrefix(uri, "\"") {
			uri = strings.Trim(uri, "\"")
			uri = strings.TrimSuffix(uri, "@en")
			return URI{Value: uri}
		}
		// try to parse ":"
		parts = strings.SplitN(uri, ":", 2)
		if len(parts) > 1 {
			return URI{Namespace: parts[0], Value: parts[1]}
		}
		uri = strings.Trim(uri, "\"")
		uri = strings.TrimSuffix(uri, "@en")
		return URI{Value: uri}
	}
	return URI{Namespace: parts[0], Value: parts[1]}
}

type Triple struct {
	Subject   URI `msg:"s"`
	Predicate URI `msg:"p"`
	Object    URI `msg:"o"`
}

func MakeTriple(sub, pred, obj string) Triple {
	s := ParseURI(sub)
	p := ParseURI(pred)
	o := ParseURI(obj)
	return Triple{
		Subject:   s,
		Predicate: p,
		Object:    o,
	}
}

func init() {
	p = &Parser{}
}

//export transform
func transform(_subject, _predicate, _object *C.char, sub_len, pred_len, obj_len C.int) {
	subject := C.GoStringN(_subject, sub_len)
	predicate := C.GoStringN(_predicate, pred_len)
	object := C.GoStringN(_object, obj_len)
	p.dataset.AddTripleStrings(subject, predicate, object)
}

//export registerNamespace
func registerNamespace(_namespace, _prefix *C.char, ns_len, pfx_len C.int) {
	namespace := C.GoStringN(_namespace, ns_len)
	prefix := C.GoStringN(_prefix, pfx_len)
	p.dataset.addNamespace(prefix, namespace)
}

// Return Parser instance
func GetParser() *Parser {
	return p
}

// Parses the given filename using the turtle format.
// Returns the dataset, and the time elapsed in parsing
func (p *Parser) Parse(filename string) (DataSet, time.Duration) {
	p.Lock()
	defer p.Unlock()
	start := time.Now()
	p.dataset = newDataSet()
	p.parseFile(filename)
	took := time.Since(start)
	return *p.dataset, took
}

// Writes the contents of the reader to a temporary file, and then reads in that file
func (p *Parser) ParseReader(r io.Reader) (DataSet, time.Duration, error) {
	p.Lock()
	defer p.Unlock()
	start := time.Now()
	p.dataset = newDataSet()
	f, err := ioutil.TempFile(".", "_raptor")
	defer f.Close()
	if err != nil {
		return *p.dataset, time.Since(start), err
	}
	defer func() {
		os.Remove(f.Name())
	}()
	_, err = io.Copy(f, r)
	if err != nil {
		return *p.dataset, time.Since(start), err
	}
	//log.Printf("Wrote %d bytes", n)
	p.parseFile(f.Name())
	took := time.Since(start)
	return *p.dataset, took, nil
}
