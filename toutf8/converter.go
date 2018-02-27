package toutf8

import (
	"io"
	"unicode/utf8"
)

type converter struct {
	writer io.Writer
	buf    [4]byte
	buflen int
	outbuf []byte
}

// NewUtf8Writer creates a writer that produces well-formed UTF-8,
// preserving the input as much as possible.
// Bytes that are not well-formed UTF-8 are assumed to be encoded in
// ISO 8859-1 (Latin-1) and are converted.
func NewUtf8Writer(wr io.Writer) io.WriteCloser {
	return &converter{writer: wr}
}

func (wr *converter) Write(bs []byte) (int, error) {
	for _, b := range bs {
		wr.buf[wr.buflen] = b
		wr.buflen++
		wr.outputBeginning()
	}

	_, err := wr.writer.Write(wr.outbuf)
	wr.outbuf = wr.outbuf[:0]
	if err != nil {
		return 0, err
	}
	return len(bs), nil
}

func (wr *converter) Close() error {
	for wr.buflen != 0 {
		buflenBefore := wr.buflen
		wr.outputBeginning()
		if wr.buflen == buflenBefore {
			wr.outputAsCodePoint()
		}
	}
	_, err := wr.writer.Write(wr.outbuf)
	wr.outbuf = wr.outbuf[:0]
	return err
}

func (wr *converter) outputBeginning() {
	buf := wr.buf
	buflen := wr.buflen

	switch {
	case buf[0] < 0x80:
		wr.outputBuf(1)
		return

	case buf[0] < 0xC0:
		break

	case buf[0] < 0xE0:
		if 1 < buflen && !cont(buf[1]) {
			break
		}
		if buflen < 2 {
			return
		}
		codePoint := codePoint(0, 0, buf[0]&0x1F, buf[1])
		if codePoint >= 0x0080 {
			wr.outputBuf(2)
			return
		}

	case buf[0] < 0xF0:
		if 1 < buflen && !cont(buf[1]) {
			break
		}
		if 2 < buflen && !cont(buf[2]) {
			break
		}
		if buflen < 3 {
			return
		}
		codePoint := codePoint(0, buf[0]&0x0F, buf[1], buf[2])
		if codePoint >= 0x0800 && utf8.ValidRune(codePoint) {
			wr.outputBuf(3)
			return
		}

	case buf[0] < 0xF8:
		if 1 < buflen && !cont(buf[1]) {
			break
		}
		if 2 < buflen && !cont(buf[2]) {
			break
		}
		if 3 < buflen && !cont(buf[3]) {
			break
		}
		if buflen < 4 {
			return
		}
		codePoint := codePoint(buf[0]&0x07, buf[1], buf[2], buf[3])
		if codePoint >= 0x010000 && codePoint < 0x110000 {
			wr.outputBuf(4)
			return
		}
	}
	wr.outputAsCodePoint()
}

func (wr *converter) outputBuf(n int) {
	buf := wr.buf
	for i := 0; i < n; i++ {
		wr.outbuf = append(wr.outbuf, buf[0])
		buf = [4]byte{buf[1], buf[2], buf[3], 0}
	}
	wr.buf = buf
	wr.buflen -= n
}

func (wr *converter) outputAsCodePoint() {
	b0 := 0xC0 + ((wr.buf[0] >> 6) & 0x1F)
	b1 := 0x80 + ((wr.buf[0] >> 0) & 0x3F)
	wr.outbuf = append(wr.outbuf, b0, b1)
	wr.buf = [4]byte{wr.buf[1], wr.buf[2], wr.buf[3], 0}
	wr.buflen--
}

func codePoint(bits18 byte, bits12 byte, bits06 byte, bits00 byte) rune {
	return (rune(bits18&0x3F) << 18) |
		(rune(bits12&0x3F) << 12) |
		(rune(bits06&0x3F) << 6) |
		rune(bits00&0x3F)
}

// cont tests whether the byte is a UTF-8 continuation byte
func cont(b byte) bool {
	return 0x80 <= b && b < 0xC0
}
