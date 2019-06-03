package cacheclient

import (
	"github.com/mediocregopher/radix/v3"
	"korto/pkg/dbconn"
	"strings"
)

type Cache struct {
	cache *MemoryCache
	serialized *dbconn.ShortUrlDB
}

type MemoryCache struct {
	pool *radix.Pool
}

func NewCache(cache *MemoryCache, serialized *dbconn.ShortUrlDB) *Cache {
	return &Cache{ cache, serialized }
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

// GetLongUrlCached returns the original url associated to the short url from the
// cache on redis
func (m *MemoryCache) GetLongUrlCached(shortUrl string) (*string, error) {
	var longUrl string
	err := m.pool.Do(radix.Cmd(&longUrl, "GET", shortUrl))
	if err != nil {
		return nil, err
	}
	return &longUrl, nil
}

// PutLongUrlCached saves a key/value type record with the shorturl as the key
// and the long url as the value. The data is saveod onto the cache on redis.
// If the key already exists, the functions ends without updating the record.
func (m *MemoryCache) PutLongUrlCached(shortUrl string, longUrl string) error {
	dbUrl, err := m.GetLongUrlCached(shortUrl)
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
// The main difference with `PutLongUrlCached` is that if the key already exists,
// the value is overwritten with the `longUrl` value.
func (m *MemoryCache) ForcePutLongUrlCache(shortUrl string, longUrl string) error {
	err := m.pool.Do(radix.Cmd(nil, "SET", shortUrl, longUrl))
	return err
}

// GetLongUrl returns the original url associated to the short url. When requested, the redis cache is first queried
// for the shortUrl, if it doesn't exists, then the coachdb is queried.
// If the shorturl is found in the couchdb, the shortUrl/longUrl pair is saved in the redis cache and the long url is
// returned. Otherwise, a nil is returned.
// If there is an error when trying to save the shortUrl/longUrl pair, the error is ignored
func (c *Cache) GetLongUrl(shortUrl string) (*string, error) {
	cachedUrl, err := c.cache.GetLongUrlCached(shortUrl)
	if err != nil {
		serializedUrl, err := c.serialized.GetLongUrl(shortUrl)
		if err != nil {
			return nil, err
		}
		if serializedUrl != nil {
			_ = c.cache.PutLongUrlCached(serializedUrl.ID, serializedUrl.LongUrl)
			return &serializedUrl.LongUrl, nil
		} else {
			return nil, nil
		}
	}
	return cachedUrl, nil
}
