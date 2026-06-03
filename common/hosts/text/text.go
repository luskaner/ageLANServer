package text

import (
	"fmt"
	"unicode/utf8"

	"golang.org/x/text/encoding"
)

const (
	EncodingUTF16LEBOM = iota
	EncodingUTF16LE
	EncodingUTF16BEBOM
	EncodingUTF16BE
	EncodingUTF8
	EncodingUTF8BOM
	EncodingANSI
	EncodingOther
)

func Decode(buf []byte) (text string, encType int, err error) {
	encType = Encoding(buf)
	var enc encoding.Encoding
	enc, err = GetEncoding(encType)
	if err != nil {
		return
	}
	var utf8Bytes []byte
	utf8Bytes, err = enc.NewDecoder().Bytes(buf)
	if err != nil {
		return
	}
	text = string(utf8Bytes)
	if !utf8.ValidString(text) {
		err = fmt.Errorf("decoded text is not valid UTF-8")
	}
	return
}

/*func Encode(text string, encType int) (buf []byte, err error) {
	if !utf8.ValidString(text) {
		err = fmt.Errorf("encoded text is not valid UTF-8")
	}
	var enc encoding.Encoding
	enc, err = GetEncoding(encType)
	if err != nil {
		return
	}
	buf, err = enc.NewEncoder().Bytes([]byte(text))
	return
}*/

func Encoding(buf []byte) (enc int) {
	enc = EncodingOther
	n := len(buf)
	if n == 0 {
		// Empty input is valid UTF-8 text.
		enc = EncodingUTF8
		return
	}
	if n < 2 {
		if utf8.Valid(buf) {
			enc = EncodingUTF8
		}
		return
	}
	// 1. Check BOM
	if buf[0] == 0xFF && buf[1] == 0xFE {
		enc = EncodingUTF16LEBOM
		return
	}
	if buf[0] == 0xFE && buf[1] == 0xFF {
		enc = EncodingUTF16BEBOM
		return
	}
	if n >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		enc = EncodingUTF8BOM
		return
	}
	// 2. Heuristic for UTF-16 without BOM based on null byte distribution
	nullInEven := 0
	nullInOdd := 0
	for i := 0; i < n-1; i += 2 {
		if buf[i] == 0x00 {
			nullInEven++
		}
		if buf[i+1] == 0x00 {
			nullInOdd++
		}
	}
	// Many null bytes in odd positions and none in even positions suggest UTF-16 LE
	if nullInOdd > n/4 && nullInEven == 0 {
		enc = EncodingUTF16LE
		return
	}
	// Many null bytes in even positions and none in odd positions suggest UTF-16 BE
	if nullInEven > n/4 && nullInOdd == 0 {
		enc = EncodingUTF16BE
		return
	}

	// 3. UTF-8 first: includes ASCII-only content.
	if utf8.Valid(buf) {
		enc = EncodingUTF8
		return
	}

	// 4. Non-UTF8 bytes with high bits are best-effort ANSI on Windows.
	for _, b := range buf {
		if b >= 0x80 {
			enc = EncodingANSI
			return
		}
	}
	return
}
