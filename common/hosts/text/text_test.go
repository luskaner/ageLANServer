package text

import "testing"

func TestEncoding(t *testing.T) {
	cases := []struct {
		name string
		buf  []byte
		want int
	}{
		{"empty", []byte{}, EncodingUTF8},
		{"single ascii byte", []byte{'A'}, EncodingUTF8},
		{"single invalid byte", []byte{0xFF}, EncodingOther},
		{"ascii text", []byte("127.0.0.1 example.com"), EncodingUTF8},
		{"utf16le bom", []byte{0xFF, 0xFE, 'h', 0x00}, EncodingUTF16LEBOM},
		{"utf16be bom", []byte{0xFE, 0xFF, 0x00, 'h'}, EncodingUTF16BEBOM},
		{"utf8 bom", []byte{0xEF, 0xBB, 0xBF, 'h', 'i'}, EncodingUTF8BOM},
		{"utf16le no bom", []byte{'h', 0x00, 'i', 0x00}, EncodingUTF16LE},
		{"utf16be no bom", []byte{0x00, 'h', 0x00, 'i'}, EncodingUTF16BE},
		{"ansi high bytes", []byte{'A', 0xE9}, EncodingANSI},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Encoding(tc.buf); got != tc.want {
				t.Fatalf("Encoding(%v) = %d, want %d", tc.buf, got, tc.want)
			}
		})
	}
}

func TestDecodeUTF8(t *testing.T) {
	text, encType, err := Decode([]byte("127.0.0.1 example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if encType != EncodingUTF8 {
		t.Fatalf("encType = %d, want %d", encType, EncodingUTF8)
	}
	if text != "127.0.0.1 example.com" {
		t.Fatalf("text = %q, want %q", text, "127.0.0.1 example.com")
	}
}

func TestDecodeEmpty(t *testing.T) {
	text, encType, err := Decode([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if encType != EncodingUTF8 {
		t.Fatalf("encType = %d, want %d", encType, EncodingUTF8)
	}
	if text != "" {
		t.Fatalf("text = %q, want empty", text)
	}
}
