//go:build !windows

package text

import (
	"fmt"

	"golang.org/x/text/encoding"
)

func GetEncoding(encodingType int) (enc encoding.Encoding, err error) {
	if encodingType != EncodingUTF8 {
		err = fmt.Errorf("only UTF-8 encoding is supported")
	}
	enc = encoding.Nop
	return
}
