package turtle

// #cgo CFLAGS: -I ../raptor/src
// #cgo LDFLAGS: -lraptor2
// #include <stdio.h>
// #include <raptor2.h>
import "C"
import "fmt"

//export Transform
func Transform(_subject, _predicate, _object *C.char, sub_len, pred_len, obj_len C.int) {
	subject := C.GoStringN(_subject, sub_len)
	predicate := C.GoStringN(_predicate, pred_len)
	object := C.GoStringN(_object, obj_len)
	addTriple(subject, predicate, object)
}

func addTriple(subject, predicate, object string) {
	fmt.Println(subject, predicate, object)
}
