package amalgam

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"math"
	"strconv"
	"strings"

	"github.com/juju/errors"
)

func EncodeID(id uint64, model string) string {
	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, id)
	crc := crc32.ChecksumIEEE(idBytes) & 0xffffffff

	message := make([]byte, 0)
	mm := make([]byte, 4)
	binary.LittleEndian.PutUint32(mm, crc)
	message = append(message, mm...)

	message = append(message, idBytes...)

	mm = make([]byte, 4)
	message = append(message, mm...)

	block, err := aes.NewCipher([]byte(Secret[:32]))
	if err != nil {
		panic(err)
	}

	t := sha256.Sum256([]byte(Secret + model))
	iv := t[:16]

	blockmode := cipher.NewCBCEncrypter(block, iv)

	blockmode.CryptBlocks(message, message)

	tt := base64.URLEncoding.EncodeToString(message)

	eid := strings.Replace(string(tt), "=", "", -1)

	return eid

}

func DecodeEID(eid string, model string) (string, error) {
	if len(eid)%3 != 0 {
		rem := len(eid) % 3
		for i := 3; i > rem; i-- {
			eid = eid + "="
		}
	}

	e, err := base64.URLEncoding.DecodeString(eid)
	if err != nil {
		return "", errors.Trace(err)
	}

	block, err := aes.NewCipher([]byte(Secret[:32]))
	if err != nil {
		return "", errors.Trace(err)
	}

	t := sha256.Sum256([]byte(Secret + model))
	iv := t[:16]

	blockmode := cipher.NewCBCDecrypter(block, iv)

	blockmode.CryptBlocks(e, e)

	b := int(e[4])
	var exp = 1
	for i := 5; i < len(e); i++ {
		if int(e[i]) != 0 {
			b += int(math.Pow(float64(256), float64(exp))) * int(e[i])
		}
		exp++
	}

	tracker := strconv.Itoa(b)

	return tracker, nil
}
