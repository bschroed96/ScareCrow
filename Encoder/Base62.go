package Base62

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Utility Functions

var ErrLength = errors.New("encoding/hex: odd length hex string")

type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("encoding/hex: invalid byte: %#U", rune(e))
}

func minOf(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func padLeft(str, pad string, length int) string {
	for len(str) < length {
		str = pad + str
	}
	return str
}

// End Utility Functions

// DecodedLen returns the length of a decoding of x source bytes.
// Specifically, it returns x / 2.
func DecodedLen(x int) int { return x / 2 }

// Decode decodes src into DecodedLen(len(src)) bytes,
// returning the actual number of bytes written to dst.
//
// Decode expects that src contains only hexadecimal
// characters and that src has even length.
// If the input is malformed, Decode returns the number
// of bytes decoded before the error.
func DecodeToByteArray(dst, src []byte) (int, error) {
	i, j := 0, 1
	for ; j < len(src); j += 2 {
		a, ok := fromHexChar(src[j-1])
		if !ok {
			return i, InvalidByteError(src[j-1])
		}
		b, ok := fromHexChar(src[j])
		if !ok {
			return i, InvalidByteError(src[j])
		}
		dst[i] = (a << 4) | b
		i++
	}
	if len(src)%2 == 1 {
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if _, ok := fromHexChar(src[j-1]); !ok {
			return i, InvalidByteError(src[j-1])
		}
		return i, ErrLength
	}
	return i, nil
}

// fromHexChar converts a hex character into its value and a success flag.
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}

const (
	base         uint64 = 62
	characterSet        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func toBase62(num uint64) string {
	encoded := ""
	for num > 0 {
		r := num % base
		num /= base
		encoded = string(characterSet[r]) + encoded
	}
	return encoded
}

func fromBase62(encoded string) (uint64, error) {
	var val uint64
	for index, char := range encoded {
		pow := len(encoded) - (index + 1)
		pos := strings.IndexRune(characterSet, char)
		if pos == -1 {
			return 0, errors.New("invalid character: " + string(char))
		}

		val += uint64(pos) * uint64(math.Pow(float64(base), float64(pow)))
	}

	return val, nil
}

const encodingChunkSize = 2

// no of bytes required in base62 to represent hex encoded string value of length encodingChunkSize
// given by formula :: int(math.Ceil(math.Log(math.Pow(16, 2*encodingChunkSize)-1) / math.Log(62)))
const decodingChunkSize = 3

func Encode(str string) string {
	var encoded strings.Builder

	inBytes := []byte(str)
	byteLength := len(inBytes)

	for i := 0; i < byteLength; i += encodingChunkSize {
		chunk := inBytes[i:minOf(i+encodingChunkSize, byteLength)]
		s := hex.EncodeToString(chunk)
		val, _ := strconv.ParseUint(s, 16, 64)
		w := padLeft(toBase62(val), "0", decodingChunkSize)
		encoded.WriteString(w)
	}
	return encoded.String()
}

func Decode(encoded string) (string, error) {
	decodedBytes := []byte{}
	for i := 0; i < len(encoded); i += decodingChunkSize {
		chunk := encoded[i:minOf(i+decodingChunkSize, len(encoded))]

		fmt.Println("chunk" + chunk)

		val, err := fromBase62(chunk)
		if err != nil {
			return "", err
		}

		fmt.Println(val)

		chunkHex := strconv.FormatUint(val, 16)
		fmt.Println("chunkHex: " + chunkHex)
		// we convert the chunk hex into byte array. Each number represented by the ascii value
		// 5 => 53 , 3 => 51 etc.

		fmt.Println([]byte(chunkHex))

		// generates a byte array which is the size of our decoded chunkHex
		dst := make([]byte, hex.DecodedLen(len([]byte(chunkHex))))

		_, err = DecodeToByteArray(dst, []byte(chunkHex))
		fmt.Println(dst)
		if err != nil {
			return "", errors.Wrap(err, "malformed input")
		}
		decodedBytes = append(decodedBytes, dst...)
		fmt.Println(decodedBytes)
	}
	s := string(decodedBytes)
	return s, nil
}
