package utils

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func tryParseUrl(urlRaw string) (string, error) {
	trimmedUrl := strings.TrimSpace(urlRaw)
	if len(trimmedUrl) == 0 {
		return "", errors.New("empty URL")
	}

	normUrl, err := url.Parse(trimmedUrl)
	if err != nil {
		return "", errors.New("invalid URL")
	}

	return normUrl.String(), nil
}

func TryGetUrlFromPostRqs(rqs *http.Request) (string, error) {
	body, err := io.ReadAll(rqs.Body)
	if err != nil {
		return "", errors.New("failed to read request body")
	}
	defer rqs.Body.Close()

	return tryParseUrl(string(body))
}

func TryGetUrlFromGetRqs(rqs *http.Request) (string, error) {
	path := strings.TrimPrefix(rqs.URL.Path, "/")

	if len(path) == 0 {
		return "", errors.New("a non-empty path is expected")
	}

	return path, nil
}
