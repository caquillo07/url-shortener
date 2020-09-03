package storage

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	maxIDTries = 5
	idLength   = 4
	letters    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
)

// memoryStore in memory storage for development and testing
type memoryStore struct {
	sync.RWMutex
	urlData map[string]*ShortURL

	// this would not scale very well, but in a read prod instance this would
	// be hooked up to a faster datastore like Redis or something like it.
	visitorData map[string][]*Visit
}

func NewMemoryStore() Storage {
	return &memoryStore{
		urlData:     make(map[string]*ShortURL),
		visitorData: make(map[string][]*Visit),
	}
}

func (ms *memoryStore) CreateURL(_ context.Context, url *ShortURL) error {
	id, err := ms.generateID()
	if err != nil {
		return err
	}

	url.UpdatedAt, url.CreatedAt = time.Now(), time.Now()
	url.ID = id
	ms.Lock()
	ms.urlData[id] = url
	ms.Unlock()
	return nil
}

func (ms *memoryStore) GetURL(_ context.Context, id string) (*ShortURL, error) {
	// Will not use defer as it adds a few ms of overhead, and we trying to be
	// as fast as possible.
	ms.RLock()
	url, ok := ms.urlData[id]
	if !ok {
		ms.RUnlock()
		return nil, ErrURLNotFound
	}
	ms.RUnlock()
	return url, nil
}

func (ms *memoryStore) RegisterVisit(_ context.Context, urlID string, visit *Visit) error {
	visit.CreatedAt = time.Now()
	ms.Lock()
	ms.visitorData[urlID] = append(ms.visitorData[urlID], visit)
	ms.Unlock()
	return nil
}

// func (ms *memoryStore) GetURLVisits(ctx context.Context, urlID string, pagination *Pagination) ([]*Visit, error) {
// 	url, err := ms.GetURL(ctx, urlID)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ms.RLock()
// 	visits, ok := ms.visitorData[url.ID]
// 	ms.RUnlock()
// 	if !ok {
// 		return []*Visit{}, nil
// 	}
//
// 	// calculate the slice
// 	offset := pagination.PerPage * pagination.Page
// 	if len(visits) < offset {
// 		return []*Visit{}, nil
// 	}
// }

func (ms *memoryStore) generateID() (string, error) {
	ms.RLock()

	for i := 0; i < maxIDTries; i++ {
		id, err := generateRandomString(idLength)
		if err != nil {
			ms.RUnlock()
			return "", err
		}
		if _, ok := ms.urlData[id]; ok {
			// we hit a collision, try again
			continue
		}
		ms.RUnlock()
		return id, nil
	}
	ms.RUnlock()
	return "", errors.New("could not generate URL ID")
}

func generateRandomString(n int) (string, error) {
	bi := big.NewInt(int64(len(letters)))
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, bi)
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
