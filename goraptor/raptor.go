package turtle

// #cgo CFLAGS: -I. -L.
// #cgo LDFLAGS: -lraptor2
// #include <stdio.h>
// #include <raptor2.h>
// extern void Transform();
// extern void RegisterNamespace();
//
// static void
// handle_triple(void* user_data, raptor_statement* triple)
// {
//   size_t sub_len, pred_len, obj_len;
//   char *subject = raptor_term_to_counted_string(triple->subject, &sub_len);
//   char *predicate = raptor_term_to_counted_string(triple->predicate, &pred_len);
//   char *object = raptor_term_to_counted_string(triple->object, &obj_len);
//   Transform(subject, predicate, object, sub_len, pred_len, obj_len);
// }
// static void
// handle_namespace(void *user_data, raptor_namespace* namespace)
// {
//   size_t ns_len, pfx_len;
//   char *ns = raptor_uri_to_counted_string(raptor_namespace_get_uri(namespace), &ns_len);
//   const char *pfx = raptor_namespace_get_counted_prefix(namespace, &pfx_len);
//   RegisterNamespace(ns, pfx, ns_len, pfx_len);
// }
// void parse_file(char *filename) {
//   raptor_world *world = NULL;
//   raptor_parser* rdf_parser = NULL;
//   unsigned char *uri_string;
//   raptor_uri *uri, *base_uri;
//
//   world = raptor_new_world();
//
//   rdf_parser = raptor_new_parser(world, "turtle");
//
//   raptor_parser_set_statement_handler(rdf_parser, NULL, handle_triple);
//   raptor_parser_set_namespace_handler(rdf_parser, NULL, handle_namespace);
//
//   uri_string = raptor_uri_filename_to_uri_string(filename);
//   uri = raptor_new_uri(world, uri_string);
//   base_uri = raptor_uri_copy(uri);
//
//   raptor_parser_parse_file(rdf_parser, uri, base_uri);
//
//   raptor_free_parser(rdf_parser);
//
//   raptor_free_uri(base_uri);
//   raptor_free_uri(uri);
//   raptor_free_memory(uri_string);
//
//   raptor_free_world(world);
// }
import "C"

func (p *Parser) ParseFile(filename string) {
	C.parse_file(C.CString(filename))
}
