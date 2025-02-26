package utils

import (
	"crypto/md5"
	"encoding/base64"
)

func GenerateShortURL(longUrl string) string {
	hash := md5.Sum([]byte(longUrl))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
