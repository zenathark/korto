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

const ShortSize = 11

type State int

const (
	Exists = 1
	OK
)

type ShortUrlRecord struct {
	ID        string `json:"_id"`
	LongUrl   string `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
}

func hash(longUrl []byte) chan []byte {
	yield := make(chan []byte)
	md := new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)
	h := make([]byte, 8)
	binary.LittleEndian.PutUint64(h, crc64.Checksum(longUrl, crc64.MakeTable(crc64.ISO)))
	go func() {
		var count int64 = 0
		for {
			n := new(big.Int)
			n.SetBytes(h[:])
			n.Add(n, big.NewInt(count))
			n.Mod(n, md)
			yield <- n.Bytes()
			count++
		}
	}()
	return yield
}

func exists(db *kivik.DB, key string) (*ShortUrlRecord, error) {
	var r ShortUrlRecord
	err := db.Get(context.TODO(), key).ScanDoc(&r)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[USE-000-02E01] Error retrieving database code: %d",
			kivik.StatusCode(err)))
	} else {
		return &r, nil
	}
}

func saveUrl(db *kivik.DB, data ShortUrlRecord) error {
	_, err := db.Put(context.TODO(), data.ID, data)
	if err != nil {
		return err
	}
	return nil
}

func replaceIllegalChars(illegalShortUrl string) string {
	noSlash := strings.ReplaceAll(illegalShortUrl, "/", "_")
	noPlus := strings.ReplaceAll(noSlash, "+", "-")
	return noPlus
}

// TODO
func commitURL(shortUrl string) State {
	return OK
}

func genShortUrl(k []byte) string {
	return replaceIllegalChars(base64.RawURLEncoding.EncodeToString(k))
}

func GenerateURL(db *kivik.DB, longUrl []byte) (*string, error) {
	key := hash(longUrl)
	defer close(key)
	for e := range key {
		url := genShortUrl(e)
		if _, err := exists(db, url); err != nil {
			return &url, nil
		}
	}
	return nil, errors.New("[USE-000-02E02] Unexpected error when generating url")
}

func RegisterURL(longUrl []byte) {

}
