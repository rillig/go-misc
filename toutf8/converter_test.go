package toutf8

import (
	"bytes"
	"io"
	"testing"
)

func TestUtf8Writer_Write_empty(t *testing.T) {
	testWrite(t, "", "")
}

func TestUtf8Writer_Write_ASCII(t *testing.T) {
	testWrite(t, "hello", "hello")
}

func TestUtf8Writer_Write_Latin1(t *testing.T) {
	testWrite(t, "\xFC", "ü")
}

func TestUtf8Writer_Write_UTF8_2(t *testing.T) {
	testWrite(t, "\xC3\xBC", "ü")
}

func TestUtf8Writer_Write_UTF8_3(t *testing.T) {
	testWrite(t, "\xE2\x82\xAC", "€")
}

func TestUtf8Writer_Write_UTF8_4(t *testing.T) {
	testWrite(t, "\xF0\x9D\x84\x9E", "𝄞")
}

func TestUtf8Writer_Write_mixed(t *testing.T) {
	testWrite(t, "Ä\xC4ö\xF6ü\xFC", "ÄÄööüü")
}

func TestUtf8Writer_Write_incomplete_2(t *testing.T) {
	testWrite(t, "\xC3", "\u00C3")
}

func TestUtf8Writer_Write_incomplete_3_1(t *testing.T) {
	testWrite(t, "\xE2", "\u00E2")
}

func TestUtf8Writer_Write_invalid_3_1(t *testing.T) {
	testWrite(t, "\xE2\xE2", "\u00E2\u00E2")
}

func TestUtf8Writer_Write_invalid_3_2(t *testing.T) {
	testWrite(t, "\xE2\xBF\xE2", "\u00E2\u00BF\u00E2")
}

func TestUtf8Writer_Write_incomplete_3_2(t *testing.T) {
	testWrite(t, "\xE2\x82", "\u00E2\u0082")
}

func TestUtf8Writer_Write_incomplete_4_1(t *testing.T) {
	testWrite(t, "\xF0", "\u00F0")
}

func TestUtf8Writer_Write_incomplete_4_2(t *testing.T) {
	testWrite(t, "\xF0\x9D", "\u00F0\u009D")
}

func TestUtf8Writer_Write_incomplete_4_3(t *testing.T) {
	testWrite(t, "\xF0\x9D\x84", "\u00F0\u009D\u0084")
	// XXX: U+0084 is not assigned; but there is no unicode.IsAssigned(rune)
}

func TestUtf8Writer_Write_invalid_4_1(t *testing.T) {
	testWrite(t, "\xF0\xFF\x84\x9E", "\u00F0\u00FF\u0084\u009E")
}

func TestUtf8Writer_Write_invalid_4_2(t *testing.T) {
	testWrite(t, "\xF0\x9D\xFF\x9E", "\u00F0\u009D\u00FF\u009E")
}

func TestUtf8Writer_Write_invalid_4_3(t *testing.T) {
	testWrite(t, "\xF0\x9D\x84\xFF", "\u00F0\u009D\u0084\u00FF")
}

func TestUtf8Writer_Write_FF(t *testing.T) {
	testWrite(t, "\xFF\xFF\xFF\xFF\xFF", "\u00FF\u00FF\u00FF\u00FF\u00FF")
}

var _ io.Writer = &noSpaceWriter{}

type noSpaceWriter struct {
	wr io.Writer
	n  int
}

func (wr *noSpaceWriter) Write(b []byte) (int, error) {
	if len(b) <= wr.n {
		n, err := wr.wr.Write(b)
		wr.n -= n
		return n, err
	}

	n, err := wr.wr.Write(b[:wr.n])
	wr.n -= n
	if err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

func TestUtf8Writer_Write_error(t *testing.T) {
	input := "hello"
	output := "hel" // truncated to 3 bytes because of the noSpaceWriter

	in := bytes.NewBufferString(input)
	out := bytes.NewBufferString("")
	wr := NewUtf8Writer(&noSpaceWriter{out, 3})

	n, err := io.Copy(wr, in)
	if err != io.ErrShortWrite {
		t.Errorf("Copy: error = %v", err)
	}
	if n != 0 {
		t.Errorf("Copy: n = %v", n)
	}

	err = wr.Close()
	if err != nil {
		t.Errorf("Close: error = %v", err)
	}

	wantOutput := output
	gotOutput := out.String()
	if gotOutput != wantOutput {
		t.Errorf("output = %q, want %q", gotOutput, wantOutput)
	}
}

func testWrite(t *testing.T, input string, output string) {
	in := bytes.NewBufferString(input)
	out := bytes.NewBufferString("")
	wr := NewUtf8Writer(out)

	gotN, err := io.Copy(wr, in)
	if err != nil {
		t.Errorf("converter.Write() error = %v", err)
		return
	}

	err = wr.Close()
	if err != nil {
		t.Errorf("converter.Close() error = %v", err)
		return
	}

	wantN := int64(len(input))
	if gotN != wantN {
		t.Errorf("converter.Write() = %v, want %v", gotN, wantN)
	}

	wantOutput := output
	gotOutput := out.String()
	if gotOutput != wantOutput {
		t.Errorf("output = %q, want %q", gotOutput, wantOutput)
	}
}
