package synchro

import (
	"context"
	"github.com/go-kivik/kivik"
	"github.com/mediocregopher/radix/v3"
	"korto/pkg/cacheclient"
	"korto/pkg/dbconn"
)

const SliceSice = 100

func migrateDatabase(cache *radix.Pool, db *kivik.DB) error {
	rows, err := db.AllDocs(context.TODO())
	if err != nil {
		return err
	}
	for rows.Next() {
		var doc dbconn.ShortUrlRecord
		if err := rows.ScanValue(doc); err != nil {
			return err
		}
		if err = cacheclient.PutLongUrl(cache, doc.ID, doc.LongUrl); err != nil {
			return err
		}
	}
	return nil
}