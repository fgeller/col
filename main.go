package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	chNUL     = 0x00
	chSpace   = 0x20
	chTab     = 0x09
	chNewline = 0x0a
)

type config struct {
	fields       []int // TODO s/fields/cols
	padded       bool
	outDelimiter []byte

	debug bool
}

func bufs() ([]byte, [][]byte) {
	in := make([]byte, 1024)
	cols := make([][]byte, 32)
	for i := range cols {
		cols[i] = make([]byte, 1024)
	}
	return in, cols
}

func debug(conf config, msg string, values ...interface{}) {
	if conf.debug {
		fmt.Fprintf(os.Stderr, msg, values...)
	}
}

func col(in io.Reader, out io.Writer, conf config) {
	inBuf, cols := bufs()
	var col, colIdx, spaceCount int
	flush := func() { // TODO: label?
		for i, f := range conf.fields {
			var length int
			col := cols[f]
			for j, c := range col {
				if c == 0 {
					length = j
					break
				}
			}
			debug(conf, "WRITING COL %v [%s]\n", f, col[:length])
			if _, err := out.Write(col[:length]); err != nil {
				log.Fatal(err)
			}
			if len(conf.fields)-1 > i {
				debug(conf, "WRITING OUT DELIM [%s]\n", conf.outDelimiter)
				if _, err := out.Write(conf.outDelimiter); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	for {
		n, err := in.Read(inBuf)
		debug(conf, "READ N=%v err=%v.\n", n, err)
		if err == io.EOF {
			flush()
			return
		}
		if err != nil {
			log.Fatal(err)
		}

		col = 0
		colIdx = 0
		spaceCount = 0
		for i := 0; i < n; i++ {
			ch := inBuf[i]
			switch {
			case ch == chSpace:
				spaceCount++
				endCol := colIdx > 0 && (!conf.padded || spaceCount > 1)
				betweenCol := colIdx == 0
				switch {
				case endCol:
					if spaceCount > 1 {
						debug(conf, "CLEARING TRAILING SPACE FOR COL %v IDX %v\n", col, colIdx)
						cols[col][colIdx-1] = 0
					}
					colIdx = 0
					col++
					continue
				case betweenCol:
					continue
				}

			default:
				spaceCount = 0
			}

			debug(conf, "ADD %x TO COL %v IDX %v\n", ch, col, colIdx)
			cols[col][colIdx] = ch
			colIdx++
		}
	}
}

func main() {}
