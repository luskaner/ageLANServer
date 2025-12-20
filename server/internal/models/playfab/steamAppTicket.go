package playfab

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
)

func readUint32(b []byte, off *int) (uint32, bool) {
	if *off+4 > len(b) {
		return 0, false
	}
	v := binary.LittleEndian.Uint32(b[*off : *off+4])
	*off += 4
	return v, true
}

func readUint64(b []byte, off *int) (uint64, bool) {
	if *off+8 > len(b) {
		return 0, false
	}
	v := binary.LittleEndian.Uint64(b[*off : *off+8])
	*off += 8
	return v, true
}

// ParseSteamIDHex decodes a ticket provided as a hex string and returns the SteamID (uint64).
func ParseSteamIDHex(ticketHex string) (uint64, error) {
	data, err := hex.DecodeString(ticketHex)
	if err != nil {
		return 0, err
	}

	off := 0

	initialLen, ok := readUint32(data, &off)
	if !ok {
		return 0, errors.New("ticket too short")
	}

	if initialLen == 20 {
		// wrapper case: read wrapper fields like node parser
		if _, ok := readUint64(data, &off); !ok {
			return 0, errors.New("unexpected end (gcToken)")
		}
		// skip 8
		off += 8
		if _, ok := readUint32(data, &off); !ok {
			return 0, errors.New("unexpected end (tokenGenerated)")
		}
		if _, ok := readUint32(data, &off); !ok {
			return 0, errors.New("unexpected end (sessionheader)")
		}
		off += 8
		if _, ok := readUint32(data, &off); !ok {
			return 0, errors.New("unexpected end (sessionExternalIP)")
		}
		off += 4
		if _, ok := readUint32(data, &off); !ok {
			return 0, errors.New("unexpected end (clientConnectionTime)")
		}
		if _, ok := readUint32(data, &off); !ok {
			return 0, errors.New("unexpected end (clientConnectionCount)")
		}
		if _, ok := readUint32(data, &off); !ok {
			return 0, errors.New("unexpected end (ownership section length)")
		}
	} else {
		// rewind the 4 bytes we read
		off -= 4
	}

	// ownership ticket length
	if _, ok := readUint32(data, &off); !ok {
		return 0, errors.New("unexpected end (ownershipLength)")
	}

	// version
	if _, ok := readUint32(data, &off); !ok {
		return 0, errors.New("unexpected end (version)")
	}

	// steamID
	steamID, ok := readUint64(data, &off)
	if !ok {
		return 0, errors.New("unexpected end (steamID)")
	}

	return steamID, nil
}
