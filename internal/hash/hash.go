package hash

import (
	"crypto/hmac"
	"crypto/sha256"
)

func GetHashSHA256(data string, key string) []byte {
	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := hmac.New(sha256.New, []byte(key))
	// передаём байты для хеширования
	h.Write([]byte(data))
	// вычисляем хеш
	return h.Sum(nil)
}
