package turtle

// #cgo CFLAGS: -I ../raptor/src
// #cgo LDFLAGS: -lraptor2
// #include <stdio.h>
// #include <raptor2.h>
import "C"
import (
	"fmt"
	"strings"
	"time"
)

type URI struct {
	Namespace string
	Value     string
}

func ParseURI(uri string) URI {
	uri = strings.TrimLeft(uri, "<")
	uri = strings.TrimRight(uri, ">")
	parts := strings.Split(uri, "#")
	if len(parts) != 2 {
		return URI{Value: uri}
	}
	return URI{Namespace: parts[0], Value: parts[1]}
}

type Triple struct {
	Subject   URI
	Predicate URI
	Object    URI
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

var p = &Parser{
	count:      0,
	Namespaces: make(map[string]string),
}

//export Transform
func Transform(_subject, _predicate, _object *C.char, sub_len, pred_len, obj_len C.int) {
	subject := C.GoStringN(_subject, sub_len)
	predicate := C.GoStringN(_predicate, pred_len)
	object := C.GoStringN(_object, obj_len)
	p.addTriple(subject, predicate, object)
}

//export RegisterNamespace
func RegisterNamespace(_namespace, _prefix *C.char, ns_len, pfx_len C.int) {
	namespace := C.GoStringN(_namespace, ns_len)
	prefix := C.GoStringN(_prefix, pfx_len)
	p.Namespaces[prefix] = namespace
}

type Parser struct {
	count      int
	Namespaces map[string]string
	Triples    []Triple
}

func NewParser(filename string) *Parser {
	start := time.Now()
	p.ParseFile(filename)
	took := time.Since(start)
	fmt.Printf("Parsed %d records in %s\n", p.count, took)
	return p
}

func (p *Parser) addTriple(subject, predicate, object string) {
	p.count += 1
	p.Triples = append(p.Triples, MakeTriple(subject, predicate, object))
}
