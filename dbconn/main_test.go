package dbconn

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivikmock"
	"github.com/stretchr/testify/assert"
	"hash/crc64"
	"testing"
	"time"
)

func loadToDB(t *testing.T, db *kivikmock.DB, data map[string]interface{}) {
	for k, v := range data {
		row, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		db.ExpectGet().WithDocID(k).WillReturn(kivikmock.DocumentT(t, row))
	}
}

func genMockDB() (*kivik.Client, *kivikmock.Client) {
	client, mock, err := kivikmock.New()
	if err != nil {
		panic(err)
	}
	return client, mock
}

//T00
func TestReplaceIllegalChars(t *testing.T) {
	or := "STtshu/oae+nsh+/a"
	expected := "STtshu_oae-nsh-_a"
	got := replaceIllegalChars(or)
	assert.Equalf(t, expected, got, "[USE-000-01T00] Illegal characters where not replaced")
}

//T01
func TestConsecutiveHashing(t *testing.T) {
	key := "1234567890789456123"
	generator := hash([]byte(key))
	genKeys := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		genKeys[i] = <-generator
		for j := 0; j < i; j++ {
			assert.NotEqualf(t, genKeys[j], genKeys[i],
				"[USE-000-02T01] Found two equal generated keys, got %d different keys before error", i)
		}
	}
}

//T02
func TestGenShortUrl(t *testing.T) {
	testUrl := "www.google.com"
	testHash := make([]byte, 8)
	binary.LittleEndian.PutUint64(testHash, crc64.Checksum([]byte(testUrl), crc64.MakeTable(crc64.ISO)))
	got := genShortUrl(testHash[:])
	assert.Equal(t, len(got), ShortSize,
		"[USE-000-02T02] Unexpected length of generated key with %s %d", ShortSize, len(testHash))
}

//T03
func TestExists(t *testing.T) {
	client, mock := genMockDB()
	testDB := mock.NewDB()
	mock.ExpectDB().WithName("_korto").WillReturn(testDB)
	testDB.ExpectGet().WithDocID("short_url").WillReturn(kivikmock.DocumentT(t,
		`{"_id":"short_url", "long_url":"long_url", "created_at":"2018-09-22T16:35:08+07:00"}`))
	db := client.DB(context.TODO(), "_korto")
	_, err := exists(db ,"short_url")
	assert.Nil(t, err, "[USE-000-02T03] Unable to retrieve expected key: test-url")
}

//T04
func TestSaveUrl(t *testing.T) {
	data := ShortUrlRecord{
		ID:        "test_url",
		LongUrl:   "loong_url",
		CreatedAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)}
	client, mock := genMockDB()
	testDB := mock.NewDB()
	mock.ExpectDB().WithName("_korto").WillReturn(testDB)
	testDB.ExpectPut().WithDocID(data.ID).WillReturn("ok")
	db := client.DB(context.TODO(), "_korto")
	err := saveUrl(db, data)
	assert.Nilf(t, err, "[USE-000-02T04] Unable to save new data with msg: %s", err)
	testDB.ExpectPut().WithDocID(data.ID).WillReturnError(errors.New("test error"))
	err = saveUrl(db , data)
	assert.NotNilf(t, err, "[USE-000-02T04] Incorrect handling error on db save", err)
}

//func TestGenerateURL(t *testing.T) {
//	data := map[string]interface{}{
//		"test_url": ShortUrlRecord{
//			ID: "test_url",
//			LongUrl: "loong_url",
//			CreatedAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
//	}
//	client, mock, err := kivikmock.New()
//	if err != nil {
//		panic(err)
//	}
//	testDB := mock.NewDB()
//	mock.ExpectDB().WithName("_korto").WillReturn(testDB)
//}
