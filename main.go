package main

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	start := "."
	if len(os.Args) > 1 {
		start = os.Args[1]
	}
	//fmt.Printf("Start: %s\n", start)
	if err := filepath.Walk(start, walkFn); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func walkFn(path string, info os.FileInfo, err error) error {
	//fmt.Fprintf(os.Stderr, "Path: %s\n", path)
	if !info.IsDir() {
		rr, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem with %s: %v\n", path, err)
			return nil
		}
		n, o := getCRC(rr)
		if err = rr.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Problem closing %s: %v\n", path, err)
		}

		fmt.Printf("File: %s\nFOUND CRC %s\n CALC CRC %x\n\n", path, o, n)
	}
	return nil
}

func getCRC(in io.Reader) (uint32, string) {
	h := crc32.NewIEEE()
	r := bufio.NewReader(in)
	old := ""
	found := false
	for {
		b, err := r.ReadBytes('\n')
		switch err {
		case nil:
		case io.EOF:
			return h.Sum32(), old
		default:
			fmt.Fprintf(os.Stdout, "Error %v\n", err)
			return 0, old
		}

		if !found {
			if len(b) > 1 && b[0] == '#' {
				if len(b) > 16 && string(b[:8]) == "# CRC32 " {
					old = string(b[8:16])
					found = true
				}
				continue
			}
		}
		_, err = h.Write(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 0, old
		}
	}
}
