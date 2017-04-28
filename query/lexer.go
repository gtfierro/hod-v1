//go:generate goyacc -o lang.go lang.y
package query

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type Token uint32

const (
	EOF Token = 1<<32 - 1 - iota
	Error
)

type Result struct {
	Token Token
	Value []byte
	Line  int
}

type Definition struct {
	Token   Token
	Pattern string
	regexp  *regexp.Regexp
}

type Scanner struct {
	quote      string
	inQuotes   bool
	tokenizer  *bufio.Scanner
	defs       []Definition
	lineNumber int
	leftover   []byte
}

func NewScanner(defs []Definition) *Scanner {
	for i := range defs {
		defs[i].regexp = regexp.MustCompile("^" + defs[i].Pattern)
	}
	return &Scanner{
		quote:      "",
		inQuotes:   false,
		defs:       defs,
		lineNumber: 1,
		leftover:   []byte{},
	}
}

func (s *Scanner) SetInput(input io.Reader) {
	s.tokenizer = bufio.NewScanner(input)
	s.tokenizer.Split(s.ScanWords)
}

func (s *Scanner) matchBytesToToken(bytes []byte) *Result {
	for _, def := range s.defs {
		if result := def.regexp.Find(bytes); result != nil {
			if len(result) != len(bytes) { // stuff leftover!
				if len(s.leftover) == 0 {
					s.leftover = append(s.leftover, bytes[len(result):]...)
				}
			}
			return &Result{Token: def.Token, Value: result, Line: s.lineNumber}
		}
	}
	return &Result{Token: Error, Line: s.lineNumber, Value: []byte("No match for '" + string(bytes) + "'")}
}

func (s *Scanner) Next() *Result {
	if len(s.leftover) > 0 {
		res := s.matchBytesToToken(s.leftover)
		if res.Token != Error {
			s.leftover = s.leftover[len(res.Value):]
			return res
		}
	}
	if !s.tokenizer.Scan() {
		err := s.tokenizer.Err()
		var ev []byte
		if err != nil {
			ev = []byte(err.Error())
		}
		return &Result{Token: Error, Line: s.lineNumber, Value: ev}
	}
	bytes := s.tokenizer.Bytes()
	return s.matchBytesToToken(bytes)
}

func (s *Scanner) Tokenize() []*Result {
	var results []*Result
	r := s.Next()
	for r != nil && r.Token != Error {
		results = append(results, r)
		r = s.Next()
	}
	return results
}

// slightly altered version of https://golang.org/src/bufio/scan.go?s=12794:12872 to keep track of line numbers
func (s *Scanner) ScanWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
		s.lineNumber += bytes.Count(data[start:], []byte{'\n'})
		s.lineNumber += bytes.Count(data[start:], []byte{'\r'})
	}
	// Scan until space, marking end of word, unless we are in a string
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		quote := strconv.QuoteRuneToASCII(r)
		if len(quote) == 3 && (quote[1] == '"' || quote[1] == '\'') {
			if s.quote == quote {
				s.quote = ""
				s.inQuotes = false
			} else {
				s.quote = quote
				s.inQuotes = true
			}
		}
		if unicode.IsSpace(r) && !s.inQuotes {
			s.lineNumber += bytes.Count(data[i:], []byte{'\n'})
			s.lineNumber += bytes.Count(data[i:], []byte{'\r'})
			return i + width, data[start:i], nil
		} else if unicode.IsSpace(r) {
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}
