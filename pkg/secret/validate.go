package secret

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const (
	hashLen = 6
)

func Hash(token, data string) string {
	secret := hmac.New(sha256.New, nil)
	secret.Write([]byte(token))

	hHash := hmac.New(sha256.New, secret.Sum(nil))
	hHash.Write([]byte(data))

	hash := hex.EncodeToString(hHash.Sum(nil))
	if len(hash) > hashLen {
		return hash[:hashLen]
	}

	return hash
}

func IsValid(token, data, hash string) error {
	secret := hmac.New(sha256.New, nil)
	secret.Write([]byte(token))

	hHash := hmac.New(sha256.New, secret.Sum(nil))
	hHash.Write([]byte(data))

	dataHash := hex.EncodeToString(hHash.Sum(nil))
	if len(dataHash) > hashLen {
		dataHash = dataHash[:hashLen]
	}

	if hash != dataHash {
		return ErrHashNotEqual
	}

	return nil
}
