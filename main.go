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
	var read, col, colIdx, spaceCount int
	var err error
	emptyLine := []byte{chNewline}

	flush := func() { // TODO: label?
		if col == 0 {
			debug(conf, "Nothing to flush")
			return
		}
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

		if err != io.EOF {
			debug(conf, "Writing empty line")
			if _, err := out.Write(emptyLine); err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	for {
		read, err = in.Read(inBuf)
		debug(conf, "Read=%v err=%v", read, err)
		if err == io.EOF {
			flush()
			return
		}
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < read; i++ {
			ch := inBuf[i]
			switch {
			case ch == chNewline:
				flush()
				col, colIdx, spaceCount = 0, 0, 0
				continue

			case ch == chSpace:
				spaceCount++
				endCol := colIdx > 0 && (!conf.padded || spaceCount > 1)
				betweenCol := colIdx == 0
				switch {
				case endCol:
					if spaceCount > 1 {
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

			debug(conf, "add %x to col %v idx %v", ch, col, colIdx)
			cols[col][colIdx] = ch
			colIdx++
		}
	}
}

func main() {}
