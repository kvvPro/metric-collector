package hash

import "crypto/sha256"

func GetHashSHA256(data string) string {
	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := sha256.New()
	// передаём байты для хеширования
	h.Write([]byte(data))
	// вычисляем хеш
	return string(h.Sum(nil))
}
