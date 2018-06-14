package turtle

// #cgo CFLAGS: -I. -L.
// #cgo LDFLAGS: -lraptor2
// #include <stdio.h>
// #include <raptor2.h>
// extern void transform();
// extern void registerNamespace();
//
// static void
// handle_triple(void* user_data, raptor_statement* triple)
// {
//   size_t sub_len, pred_len, obj_len;
//   char *subject = raptor_term_to_counted_string(triple->subject, &sub_len);
//   char *predicate = raptor_term_to_counted_string(triple->predicate, &pred_len);
//   char *object = raptor_term_to_counted_string(triple->object, &obj_len);
//   transform(subject, predicate, object, sub_len, pred_len, obj_len);
// }
// static void
// handle_namespace(void *user_data, raptor_namespace* namespace)
// {
//   size_t ns_len, pfx_len;
//   char *ns = raptor_uri_to_counted_string(raptor_namespace_get_uri(namespace), &ns_len);
//   const char *pfx = raptor_namespace_get_counted_prefix(namespace, &pfx_len);
//   registerNamespace(ns, pfx, ns_len, pfx_len);
// }
// void parse_file_turtle(char *filename) {
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
// void parse_file_ntriples(char *filename) {
//   raptor_world *world = NULL;
//   raptor_parser* rdf_parser = NULL;
//   unsigned char *uri_string;
//   raptor_uri *uri, *base_uri;
//
//   world = raptor_new_world();
//
//   rdf_parser = raptor_new_parser(world, "ntriples");
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
// void parse_file_guess(char *filename) {
//   raptor_world *world = NULL;
//   raptor_parser* rdf_parser = NULL;
//   unsigned char *uri_string;
//   raptor_uri *uri, *base_uri;
//
//   world = raptor_new_world();
//
//   rdf_parser = raptor_new_parser(world, "guess");
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

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var linecutoff = int64(5e6)
var epsilon = int64(1000)

// TODO: if the file is over a certain length, chunk it!
func (p *Parser) parseFile(filename string) {
	extension := filepath.Ext(filename)
	files := chunkFile(filename)
	log.Println(files)
	switch extension {
	case "ttl", "turtle":
		for _, filename := range files {
			C.parse_file_turtle(C.CString(filename))
		}
	case "n3", "ntriples":
		for _, filename := range files {
			C.parse_file_ntriples(C.CString(filename))
		}
	default:
		for _, filename := range files {
			C.parse_file_guess(C.CString(filename))
		}
	}
}

// Given a large source file, we need to chunk it into several smaller files.
func chunkFile(filename string) []string {
	numlines, numbytes := getFileSize(filename)
	log.Println("Lines:", numlines, "Bytes:", numbytes)
	// if more than 10 million lines, then we need to chunk
	var numFiles = int64(-1)
	if numlines > linecutoff+epsilon {
		numFiles = numlines / linecutoff
		log.Printf("Splitting %s into %d files", filename, numFiles)
	} else {
		// if not too many lines, then we just return the current filename
		return []string{filename}
	}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	header := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@prefix") || strings.HasPrefix(line, "@PREFIX") {
			header += line + "\n"
		} else {
			break
		}
	}
	// now have the header. Need to read approximately linecutoff lines into 'numFiles' files
	chunkedfiles := make([]string, numFiles+1)
	suffix := []byte{'.'}
	eol := []byte{'\n'}
	for idx := int64(0); idx < numFiles+int64(1); idx++ {
		chunkfile, err := ioutil.TempFile(".", "chunk")
		defer chunkfile.Close()
		if err != nil {
			log.Fatal(err)
		}
		// write the header
		if _, err := chunkfile.Write([]byte(header)); err != nil {
			log.Fatal(err)
		}

		numread := int64(0)
		for scanner.Scan() {
			line := scanner.Bytes()
			numread += 1
			if _, err := chunkfile.Write(line); err != nil {
				log.Fatal(err)
			}
			if _, err := chunkfile.Write(eol); err != nil {
				log.Fatal(err)
			}
			if numread > linecutoff && bytes.HasSuffix(line, suffix) {
				break
			}
		}
		chunkedfiles[idx] = chunkfile.Name()
	}
	return chunkedfiles
}

func getFileSize(filename string) (numlines, numbytes int64) {
	// get bytes
	stats, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	numbytes = stats.Size()

	// get lines
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 32*1024)
	sep := []byte{'\n'}
	for {
		n, err := file.Read(buf)
		numlines += int64(bytes.Count(buf[:n], sep))
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
		}
	}
	return

}
