package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCol(t *testing.T) {
	data := []struct {
		name     string
		in       string
		conf     config
		expected string
	}{
		{
			name:     "baseline",
			in:       "REPOSITORY                               TAG                 IMAGE ID            CREATED             SIZE",
			conf:     config{debug: true, fields: []int{1, 3}, padded: true, outDelimiter: []byte(" ")},
			expected: "TAG CREATED",
		},
	}

	for _, d := range data {
		buf := new(bytes.Buffer)
		reader := strings.NewReader(d.in)

		col(reader, buf, d.conf)
		actual := buf.String()

		if d.expected != actual {
			t.Errorf(`%v failed:
expected: %#v
actual:   %#v
`, d.name, d.expected, actual)
		}
	}
}
