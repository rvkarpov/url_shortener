package urlutils

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func tryParseURL(urlRaw string) (string, error) {
	trimmedURL := strings.TrimSpace(urlRaw)
	if len(trimmedURL) == 0 {
		return "", errors.New("empty URL")
	}

	normURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return "", errors.New("invalid URL")
	}

	return normURL.String(), nil
}

func TryGetURLFromPostRqs(rqs *http.Request) (string, error) {
	body, err := io.ReadAll(rqs.Body)
	if err != nil {
		return "", errors.New("failed to read request body")
	}
	defer rqs.Body.Close()

	return tryParseURL(string(body))
}
