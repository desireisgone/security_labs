package scanner

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const HASH_SUM_SIZE = 2

type Result struct {
	CalculatedHash uint16
	SavedHash      uint16
	MarkerOffset   int64
}

type Scanner struct {
	buffer                []byte
	bufferForSearchMarker []byte
	savedHashSumOffset    int64
	fileIndex             int64

	marker []byte
}

func NewScanner(marker []byte, bufferSize int) *Scanner {
	if bufferSize%HASH_SUM_SIZE != 0 {
		fmt.Printf("Buffer size must be a multiple of %d\n", HASH_SUM_SIZE)
		return nil
	}

	return &Scanner{
		buffer:                make([]byte, bufferSize),
		bufferForSearchMarker: make([]byte, 0),
		marker:                marker,
	}
}

func (h *Scanner) ScanFile(filePath string) (Result, error) {
	var res Result
	var hashSum uint16 = 0

	res.MarkerOffset = -1

	file, err := os.Open(filePath)
	if err != nil {
		return res, err
	}
	defer file.Close()

	for {
		read, err := file.Read(h.buffer)
		if read == 0 {
			if err == io.EOF {
				break
			}
			return res, err
		}

		h.searchMarker(&res)

		h.calculateHashCode(&res, &hashSum, read)

		h.fileIndex += int64(read)
		if err == io.EOF {
			break
		}
	}

	res.CalculatedHash = hashSum
	return res, nil
}

func (h *Scanner) searchMarker(res *Result) {
	if res.MarkerOffset == -1 {
		searchBuf := append(h.bufferForSearchMarker, h.buffer...)
		idx := bytes.Index(searchBuf, h.marker)

		if idx != -1 {
			res.MarkerOffset = h.fileIndex - int64(len(h.bufferForSearchMarker)) + int64(idx)
			h.savedHashSumOffset = res.MarkerOffset + int64(len(h.marker))
		}

		if len(h.buffer) >= len(h.marker) {
			h.bufferForSearchMarker = h.buffer[len(h.buffer)-len(h.marker):]
		} else {
			h.bufferForSearchMarker = searchBuf
			if len(h.bufferForSearchMarker) > len(h.marker) {
				h.bufferForSearchMarker = h.bufferForSearchMarker[len(h.bufferForSearchMarker)-len(h.marker):]
			}
		}
	}
}

func (h *Scanner) calculateHashCode(res *Result, hashSum *uint16, bytesToReadCount int) {
	if bytesToReadCount%HASH_SUM_SIZE != 0 {
		h.buffer[bytesToReadCount] = 0
		bytesToReadCount++
	}

	for i := 0; i < bytesToReadCount; i += HASH_SUM_SIZE {
		blockBytes := uint16(h.buffer[i])<<8 | uint16(h.buffer[i+1])

		currentPairOffset := h.fileIndex + int64(i)

		if res.MarkerOffset != -1 {
			if currentPairOffset == h.savedHashSumOffset {
				res.SavedHash = blockBytes
				blockBytes = 0
			} else if currentPairOffset == h.savedHashSumOffset+1 {
				res.SavedHash |= blockBytes >> 8
				blockBytes &= 0x00FF
			} else if currentPairOffset == h.savedHashSumOffset-1 {
				res.SavedHash |= blockBytes << 8
				blockBytes &= 0xFF00
			}
		}

		*hashSum ^= blockBytes
	}
}
