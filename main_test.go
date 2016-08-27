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
			conf:     config{debug: false, cols: []int{1, 3}, padded: true, outDelimiter: []byte(" ")},
			expected: "TAG CREATED",
		},
		{
			name: "trailing newline",
			in: `REPOSITORY                               TAG                 IMAGE ID            CREATED             SIZE
`,
			conf: config{debug: false, cols: []int{1, 3}, padded: true, outDelimiter: []byte(" ")},
			expected: `TAG CREATED
`,
		},
		{
			name: "multi-line",
			in: `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
example.com/hans    v1.2.3              049ca1a8121c        13 hours ago        18.35 MB
example.com/gerd    v1.6.12             8c77bb6fe931        15 hours ago        14.84 MB`,
			conf: config{debug: false, cols: []int{1, 3}, padded: true, outDelimiter: []byte(" ")},
			expected: `TAG CREATED
v1.2.3 13 hours ago
v1.6.12 15 hours ago`,
		},
		{
			name:     "leading whitespace",
			in:       " a b",
			conf:     config{debug: false, cols: []int{0, 1}, padded: false, outDelimiter: []byte(" ")},
			expected: "a b",
		},
		{
			name:     "re-ordering",
			in:       " a b",
			conf:     config{debug: false, cols: []int{1, 0}, padded: false, outDelimiter: []byte(" ")},
			expected: "b a",
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
