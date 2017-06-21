package main

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type crcRecord struct {
	CRC    uint32
	OldCRC uint32
	Path   string
	Mode   os.FileMode
}

func calculate(start string) <-chan crcRecord {
	out := make(chan crcRecord)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if isLDIF(info) {
			doCalc(path, info.Mode(), out)
		}
		return nil
	}
	go func() {
		if err := filepath.Walk(start, walkFn); err != nil {
			log.Fatalf("error: %v", err)
		}
		close(out)
	}()
	return out
}

func doCalc(path string, mode os.FileMode, res chan<- crcRecord) {
	crc, err := calculateCRC(path, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem parsing %s: %v\n", path, err)
		return
	}
	res <- crc
}

func calculateCRC(path string, mode os.FileMode) (crcRecord, error) {
	record := crcRecord{Path: path, Mode: mode}
	fd, err := os.Open(path)
	if err != nil {
		return record, err
	}

	crc, old, err := extractCRC(fd)
	if err == nil {
		var parsed uint64
		parsed, err = strconv.ParseUint(old, 16, 32)
		record.CRC = crc
		record.OldCRC = uint32(parsed)
	}

	if e2 := fd.Close(); e2 != nil {
		merr := newErr(err)
		merr.Related = append(merr.Related, e2)
		err = merr
	}
	return record, err
}

func extractCRC(in io.Reader) (uint32, string, error) {
	h := crc32.NewIEEE()
	r := bufio.NewReader(in)
	foundCRC := "0"
	pastHeader := false
	for {
		line, err := r.ReadBytes('\n')
		switch err {
		case nil:
		case io.EOF:
			return h.Sum32(), foundCRC, nil
		default:
			return 0, foundCRC, err
		}

		if !pastHeader {
			if isComment(line) {
				if isCRC(line) {
					foundCRC = string(line[8:16])
					pastHeader = true
				}
				continue
			}
		}
		_, err = h.Write(line)
		if err != nil {
			return 0, foundCRC, err
		}
	}
}
