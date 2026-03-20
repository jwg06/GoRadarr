package downloadclient

import "context"

// Item represents a single download managed by a download client.
type Item struct {
	ID       string
	Name     string
	Size     int64
	SizeLeft int64
	Status   string
	Hash     string
	Category string
}

// Client is the common interface all download client implementations must satisfy.
type Client interface {
	Name() string
	Protocol() string // "torrent" | "usenet"
	TestConnection(ctx context.Context) error
	AddTorrent(ctx context.Context, magnetOrURL string, savePath string) error
	AddNZB(ctx context.Context, nzbURL string, category string) error
	GetItems(ctx context.Context) ([]Item, error)
	RemoveItem(ctx context.Context, id string, deleteFiles bool) error
}
