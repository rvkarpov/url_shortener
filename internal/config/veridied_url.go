package config

import (
	"net/url"
)

type VerifiedURL string

func (addr VerifiedURL) String() string {
	return string(addr)
}

func (addr *VerifiedURL) Set(value string) error {
	*addr = VerifiedURL(value)
	_, err := url.ParseRequestURI(value)
	return err
}

func (addr *VerifiedURL) UnmarshalText(text []byte) error {
	return addr.Set(string(text))
}
