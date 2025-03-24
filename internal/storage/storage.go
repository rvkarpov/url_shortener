package storage

type URLStorage interface {
	StoreURL(shortURL, longURL string) error
	TryGetLongURL(shortURL string) (string, error)
	Finalize()
}
