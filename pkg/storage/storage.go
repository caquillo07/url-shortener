package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrURLNotFound = errors.New("url not found")
)

type ShortURL struct {
	ID        string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Visit struct {
	URLID     string
	IP        string
	Referer   string
	UserAgent string
	CreatedAt time.Time
}

type Pagination struct {
	Page int
	PerPage int
}

type Storage interface {
	// CreateURL creates the given ShortURL. This method should/will also
	// update the ID of the newly created record.
	CreateURL(ctx context.Context, url *ShortURL) error

	// GetURL returns the ShortURL for the given ID, or ErrURLNotFound error
	// if not found.
	GetURL(ctx context.Context, id string) (*ShortURL, error)

	// RegisterVisit registers information about a visitor to a given URL
	RegisterVisit(ctx context.Context, urlID string, visit *Visit) error
}
