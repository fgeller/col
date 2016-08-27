package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

const (
	chNUL     = 0x00
	chSpace   = 0x20
	chTab     = 0x09
	chNewline = 0x0a
)

type config struct {
	cols         []int
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
		fmt.Fprintf(os.Stderr, msg+"\n", values...)
	}
}

func col(in io.Reader, out io.Writer, conf config) {
	inBuf, cols := bufs()
	var (
		bufIdx, readLen, col, colIdx, spcs int
		err                                error
		emptyLine                          = []byte{chNewline}
	)

read:
	bufIdx = 0
	readLen, err = in.Read(inBuf)
	debug(conf, "readLen=%v err=%v", readLen, err)
	if err != nil {
		debug(conf, "err=%v  flushing!", err)
		goto flush
	}

loop:
	for {
		if bufIdx >= readLen {
			goto read
		}

		ch := inBuf[bufIdx]
		bufIdx++
		switch ch {
		case chNewline:
			debug(conf, "newline - flushing!")
			goto flush

		case chSpace:
			spcs++
			endCol := colIdx > 0 && (!conf.padded || spcs > 1)
			betweenCol := colIdx == 0
			switch {
			case endCol:
				if spcs > 1 {
					cols[col][colIdx-1] = 0
				}
				colIdx = 0
				col++
				continue loop
			case betweenCol:
				continue loop
			}

		default:
			spcs = 0
		}

		debug(conf, "add %x to col %v idx %v", ch, col, colIdx)
		cols[col][colIdx] = ch
		colIdx++
	}

flush:
	if err != nil && (col|colIdx == 0) {
		return
	}
	col, colIdx, spcs = 0, 0, 0
	debug(conf, "Flushing")

	for i, c := range conf.cols {
		var length int
		col := cols[c]
		for j, ch := range col {
			if ch == 0 {
				length = j
				break
			}
		}
		debug(conf, "Writing col %v [%s]", c, col[:length])
		if _, err := out.Write(col[:length]); err != nil {
			log.Fatal(err)
		}

		debug(conf, "Nulling col %v len %v", c, length)
		for j := 0; j <= length; j++ {
			col[j] = 0
		}

		if len(conf.cols)-1 > i {
			debug(conf, "Writing out delim [%s]", conf.outDelimiter)
			if _, err := out.Write(conf.outDelimiter); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err != nil {
		return
	}

	debug(conf, "Writing empty line")
	if _, err := out.Write(emptyLine); err != nil {
		log.Fatal(err)
	}

	goto loop
}

func main() {
	pd := flag.Bool("padded", true, "Expect padded input.")
	dbg := flag.Bool("debug", false, "Print debug output.")
	oDelim := flag.String("out-delimiter", " ", "Output delimiter.")

	flag.Parse()

	conf := config{
		debug:        *dbg,
		padded:       *pd,
		outDelimiter: []byte(*oDelim),
	}

	for _, c := range flag.Args() {
		i, err := strconv.Atoi(c)
		if err != nil {
			log.Fatalf("Failed to interpret %#v as a number, err=%v", c, err)
		}
		conf.cols = append(conf.cols, i-1)
	}

	debug(conf, "config %#v", conf)

	col(os.Stdin, os.Stdout, conf)
}
