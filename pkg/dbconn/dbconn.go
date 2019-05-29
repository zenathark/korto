package dbconn

import (
	"context"
	"errors"
	"time"

	//"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/go-kivik/kivik"
	"hash/crc64"
	"math/big"
	"strings"
)

// Size of the short url
const ShortSize = 11
const commitTries = 3

// Short url data model
type ShortUrlRecord struct {
	ID        string    `json:"_id"`
	LongUrl   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
}

type ShortUrlDB struct {
	db *kivik.DB
}

// NewShortUrlDB returns an instance of ShortUrlDB that holds a reference to a couch `database`. If the server
// is unreachable or the database does not exists, the function returns an error
func NewShortUrlDB(address string, database string) (*ShortUrlDB, error) {
	c, err := kivik.New("couch", address)
	if err != nil {
		return nil, err
	}

	db := c.DB(context.TODO(), database)
	if db.Err() != nil {
		return nil, db.Err()
	}

	return &ShortUrlDB{db}, nil
}

// WithDatabase returns an instance of ShortUrlDB using `db` as its database. Mostly for DB
func WithDatabase(db *kivik.DB) *ShortUrlDB {
	return &ShortUrlDB{db}
}

// exists checks for the `key` to exists on the database `db`. If key exists on the
// database, the row associated is returned, otherwise, the function returns an error.
func (c *ShortUrlDB) exists(key string) (*ShortUrlRecord, error) {
	var r ShortUrlRecord
	err := c.db.Get(context.TODO(), key).ScanDoc(&r)
	if err != nil {
		if kivik.StatusCode(err) == kivik.StatusNotFound {
			return nil, nil
		} else {
			return nil,
				errors.New(fmt.Sprintf("[USE-000-02E01] Error reading the database %s %d", key, kivik.StatusCode(err)))
		}
	} else {
		return &r, nil
	}
}

// existsLongUrl checks if the longUrl exists on the database `db`. If longUrl exists on the
// database, the row associated is returned otherwise, an shortUrl pointing to an empty block
// is returned.
func (c *ShortUrlDB) existsLongUrl(longUrl []byte) (*ShortUrlRecord, *string, error) {
	hashGenerator := hash(longUrl)
	for h := range hashGenerator {
		url := genShortUrl(h)
		r, err := c.exists(url)
		if err != nil {
			return nil, nil, err
		}
		if r != nil {
			if strings.Compare(r.LongUrl, string(longUrl)) == 0{
				return r, nil, nil
			}
		} else {
			return nil, &url, nil
		}
	}
	panic("[USE-000-02E5] Error generating keys, code must not reach here")
}

// saveUrl inserts a new short url record onto the database, returns an error if the operation fails.
func (c *ShortUrlDB) saveUrl(url ShortUrlRecord) error {
	_, err := c.db.Put(context.TODO(), url.ID, url)
	if err != nil {
		return err
	}
	return nil
}

// GenerateURL returns a new short url that doesn't exists on the database at the time of
// the call.
func (c *ShortUrlDB) GenerateURL(longUrl []byte) (*string, error) {
	key := hash(longUrl)
	defer close(key)
	for e := range key {
		url := genShortUrl(e)
		if _, err := c.exists(url); err != nil {
			return &url, nil
		}
	}
	return nil, errors.New("[USE-000-02E02] Unexpected error when generating url")
}

// CommitURL tries to save the new short url, if the short url already exists, the function
// returns the row associated, otherwise returns the assigned short Url.
// IF saveUrl fails, the function will retry up to 3 times before returning an error.
func (c *ShortUrlDB) CommitURL(longUrl []byte) (*ShortUrlRecord, *string, error) {
	for i := 0; i < commitTries; i++ {
		r, url, err := c.existsLongUrl(longUrl)
		if err != nil {
			return nil, nil, err
		}
		if r != nil {
			return r, nil, nil
		}
		newRecord := ShortUrlRecord{
			ID:        *url,
			LongUrl:   string(longUrl),
			CreatedAt: time.Now(),
		}
		if err = c.saveUrl(newRecord); err == nil {
			return nil, url, nil
		}
	}
	return nil, nil, errors.New("[USE-000-02E03] Unable to save into the database")
}

// ---- Pure functions ------

// calcBase returns the base for the key 2^64 as a BigInt
func calcBase() *big.Int {
	return new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)
}

// hashAsByteArray generates a new 64bit hash using crc64 and returns it as a byte array of size 8
func hashAsByteArray(longUrl []byte) []byte {
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, crc64.Checksum(longUrl, crc64.MakeTable(crc64.ISO)))
	return buffer
}

func nextHash(currentHash []byte, offset int64, base *big.Int) []byte {
	n := new(big.Int)
	n.SetBytes(currentHash)
	n.Add(n, big.NewInt(offset))
	n.Mod(n, base)
	return n.Bytes()
}

// hash generates a series of hash signatures with the given longUrl as seed.
// The signatures (s_i) are defined as follow:
// s_0 = crc64(longUrl)
// s_i = s_{i-1} + 1
// The hash function is interpreted as a BigInt on the addition operation
// returns a channel with all generated keys lazily
func hash(longUrl []byte) chan []byte {
	yield := make(chan []byte)
	base := calcBase()
	h := hashAsByteArray(longUrl)
	go func() {
		var count int64 = 0
		for {
			yield <- nextHash(h, count, base)
			count++
		}
	}()
	return yield
}


// Generates a new short url
func genShortUrl(k []byte) string {
	return replaceIllegalChars(base64.RawURLEncoding.EncodeToString(k))
}

// replace illegal base64 characters.
// Slash (/) is replaced by underscore (_)
// Plus (+) is replaced by dash (-)
func replaceIllegalChars(illegalShortUrl string) string {
	noSlash := strings.ReplaceAll(illegalShortUrl, "/", "_")
	noPlus := strings.ReplaceAll(noSlash, "+", "-")
	return noPlus
}

