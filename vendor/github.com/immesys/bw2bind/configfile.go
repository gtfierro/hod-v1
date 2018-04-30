package bw2bind

import (
	"bytes"
	"text/template"
)

// LoadConfigFile allows you to preprocess an augmented
// file into processed form. This is not normally used.
func LoadConfigFile(contents string) ([]byte, error) {
	tmp := template.New("root")
	buf := &bytes.Buffer{}
	data := struct{}{}
	rv, err := tmp.Parse(contents)
	if err != nil {
		return nil, err
	}
	err = rv.ExecuteTemplate(buf, "root", data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
