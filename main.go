package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	start := "."
	if len(os.Args) > 1 {
		start = os.Args[1]
	}
	for msg := range update(calculate(start)) {
		fmt.Println(msg)
	}
}

func update(in <-chan crcRecord) <-chan string {
	out := make(chan string)
	go func() {
		for c := range in {
			out <- updateFile(c)
		}
		close(out)
	}()
	return out
}

func updateFile(c crcRecord) string {
	in, err := os.Open(c.Path)
	if err != nil {
		return fmt.Sprintf("error opening %s: %v", c.Path, err)
	}
	tmp := c.Path + ".new"
	out, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, c.Mode)
	if err != nil {
		return fmt.Sprintf("error opening %s: %v", tmp, multiClose(err, in))
	}
	r := bufio.NewReader(in)
	done := false
	header := true
	for !done {
		buf, e := r.ReadBytes('\n')
		switch e {
		case nil:
		case io.EOF:
			done = true
		default:
			return fmt.Sprintf("error reading %s: %v", c.Path, multiClose(e, in, out))
		}
		if header {
			if isComment(buf) {
				if isCRC(buf) {
					crc := []byte(fmt.Sprintf("%x", c.CRC))
					if len(crc) != 8 {
						err = fmt.Errorf("invalid crc '%x'", c.CRC)
						return fmt.Sprintf("error updating crc: %v", multiClose(err, in, out))
					}
					for x, b := range crc {
						buf[8+x] = b
					}
					header = false
				}
			}
		}
		if _, err = out.Write(buf); err != nil {
			return fmt.Sprintf("error writing %s: %v", c.Path, multiClose(err, in, out))
		}
	}
	if err = multiClose(nil, in, out); err != nil {
		return fmt.Sprintf("error closing: %v", err)
	}
	if err = os.Rename(tmp, c.Path); err != nil {
		return fmt.Sprintf("error renaming %s to %s: %v", tmp, c.Path, err)
	}
	msg := "OK"
	if c.CRC != c.OldCRC {
		msg = "Updated"
	}
	return fmt.Sprintf("%s %s", c.Path, msg)
}

func isComment(l []byte) bool { return len(l) > 1 && l[0] == '#' }
func isCRC(l []byte) bool     { return len(l) > 16 && string(l[:8]) == "# CRC32 " }
func isLDIF(i os.FileInfo) bool {
	if i.IsDir() {
		return false
	}
	if strings.HasSuffix(i.Name(), ".ldif") {
		return true
	}
	return false
}
