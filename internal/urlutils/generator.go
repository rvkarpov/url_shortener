package urlutils

import (
	"crypto/md5"
	"encoding/base64"
	"strings"
)

func GenerateShortURL(longURL string) string {
	hash := md5.Sum([]byte(longURL))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(encoded, "=")[:8]
}
