package storage

import "fmt"

type DuplicateURLError struct {
	URL string
}

func (e *DuplicateURLError) Error() string {
	return fmt.Sprintf("duplicate URL violation: %s", e.URL)
}

func NewDuplicateURLError(url string) error {
	return &DuplicateURLError{URL: url}
}
