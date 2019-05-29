package cacheclient

import (
	"github.com/mediocregopher/radix/v3"
	"strings"
)

type MemoryCache struct {
	pool *radix.Pool
}

// NewMemoryCache creates a MemoryCache that keeps a connection pool to the cache implemented in
// redis. For argument details, check radix.NewPool
func NewMemoryCache(network string, address string, size int, opts ...radix.PoolOpt) (*MemoryCache, error) {
	c, err := radix.NewPool(network, address, size, opts...)
	if err != nil {
		return nil, err
	}
	return &MemoryCache{c}, nil
}

func WithPool(c *radix.Pool) *MemoryCache {
	return &MemoryCache{c}
}

// GetLongUrl returns the original url associated to the short url from the
// cache on redis
func (m *MemoryCache) GetLongUrl(shortUrl string) (*string, error) {
	var longUrl string
	err := m.pool.Do(radix.Cmd(&longUrl, "GET", shortUrl))
	if err != nil {
		return nil, err
	}
	return &longUrl, nil
}

// PutLongUrl saves a key/value type record with the shorturl as the key
// and the long url as the value. The data is saveod onto the cache on redis.
// If the key already exists, the functions ends without updating the record.
func (m *MemoryCache) PutLongUrl(shortUrl string, longUrl string) error {
	dbUrl, err := m.GetLongUrl(shortUrl)
	if err != nil {
		return err
	}
	if strings.Compare(*dbUrl, "") != 0 {
		return nil
	}
	err = m.pool.Do(radix.Cmd(nil, "SET", shortUrl, longUrl))
	return err
}

// ForceLongUrl saves a key/value type record with the shorturl as the key
// and the long url as the value. The data is saveod onto the cache on redis.
// The main difference with `PutLongUrl` is that if the key already exists,
// the value is overwritten with the `longUrl` value.
func (m *MemoryCache) ForcePutLongUrl(shortUrl string, longUrl string) error {
	err := m.pool.Do(radix.Cmd(nil, "SET", shortUrl, longUrl))
	return err
}
