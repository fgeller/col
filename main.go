package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
)

const (
	chSpace   = 0x20
	chNewline = 0x0a
)

type config struct {
	cols         []int
	padded       bool
	outDelimiter []byte
	debug        bool
}

func bufs() ([]byte, [][]byte) {
	in := make([]byte, 4*1024)
	cols := make([][]byte, 32)
	for i := range cols {
		cols[i] = make([]byte, 1024)
	}
	return in, cols
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
	if err != nil {
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

		cols[col][colIdx] = ch
		colIdx++
	}

flush:
	if err != nil && (col|colIdx == 0) {
		return
	}
	col, colIdx, spcs = 0, 0, 0

	for i, c := range conf.cols {
		var length int
		col := cols[c]
		for j, ch := range col {
			if ch == 0 {
				length = j
				break
			}
		}

		if _, err := out.Write(col[:length]); err != nil {
			log.Fatal(err)
		}

		for j := 0; j <= length; j++ {
			col[j] = 0
		}

		if len(conf.cols)-1 > i {
			if _, err := out.Write(conf.outDelimiter); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err != nil {
		return
	}

	if _, err := out.Write(emptyLine); err != nil {
		log.Fatal(err)
	}

	goto loop
}

func main() {
	pd := flag.Bool("padded", true, "Expect padded input.")
	oDelim := flag.String("out-delimiter", " ", "Output delimiter.")
	inFile := flag.String("in", "", "Input file - default is stdin.")

	flag.Parse()

	conf := config{
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

	var err error
	in := os.Stdin
	if *inFile != "" {
		if in, err = os.Open(*inFile); err != nil {
			log.Fatal(err)
		}
	}

	out := bufio.NewWriter(os.Stdout)
	col(in, out, conf)
	out.Flush()
}
