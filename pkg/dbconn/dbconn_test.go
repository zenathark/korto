package dbconn

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivikmock"
	"github.com/stretchr/testify/assert"
	"hash/crc64"
	"net/http"
	"testing"
	"time"
)

func generateTestData() *ShortUrlRecord {
	rowTest := ShortUrlRecord{
		ID:        "",
		LongUrl:   "www.google.com",
		CreatedAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	}
	rowTest.ID = string(genShortUrl(hashAsByteArray([]byte(rowTest.LongUrl))))
	return &rowTest
}

func genMockDB() (*ShortUrlDB, *kivikmock.DB) {
	client, mock, err := kivikmock.New()
	if err != nil {
		panic(err)
	}
	testDB := mock.NewDB()
	mock.ExpectDB().WithName("_korto").WillReturn(testDB)
	db := client.DB(context.TODO(), "_korto")
	refDb := WithDatabase(db)
	return refDb, testDB
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
	refDb, testDB := genMockDB()
	testDB.ExpectGet().WithDocID("short_url").WillReturn(kivikmock.DocumentT(t,
		`{"_id":"short_url", "long_url":"long_url", "created_at":"2018-09-22T16:35:08+07:00"}`))
	_, err := refDb.GetLongUrl("short_url")
	assert.Nil(t, err, "[USE-000-02T03] Unable to retrieve expected key: test-url")
}

//T04
func TestSaveUrl(t *testing.T) {
	data := ShortUrlRecord{
		ID:        "test_url",
		LongUrl:   "loong_url",
		CreatedAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)}
	refDb, testDB := genMockDB()
	testDB.ExpectPut().WithDocID(data.ID).WillReturn("ok")
	err := refDb.saveUrl(data)
	assert.Nilf(t, err, "[USE-000-02T04] Unable to save new data with msg: %s", err)
	testDB.ExpectPut().WithDocID(data.ID).WillReturnError(errors.New("test error"))
	err = refDb.saveUrl(data)
	assert.NotNilf(t, err, "[USE-000-02T04] Incorrect handling error on db save", err)
}

//T05
func TestCommitURLWhenDoesntExists(t *testing.T) {
	rowTest := generateTestData()
	refDb, testDB := genMockDB()
	testDB.ExpectGet().WithDocID(rowTest.ID).WillReturnError(&kivik.Error{HTTPStatus: http.StatusNotFound, Err: nil})
	testDB.ExpectPut().WithDocID(rowTest.ID).WillReturn("ok")
	s, url, err := refDb.CommitURL([]byte(rowTest.LongUrl))
	if err != nil {
		panic(fmt.Sprintf("[USE-000-02T05] Malformed mock data %s", err))
	}
	assert.Nil(t, s, "[USE-000-02T05] Unexpected short url found")
	assert.NotNil(t, url, "[USE-000-02T05] Short url not generated")
}

//T06
func TestCommitURLWhenExists(t *testing.T) {
	rowTest := generateTestData()
	rowTestAsJson, _ := json.Marshal(rowTest)
	refDb, testDB := genMockDB()
	testDB.ExpectGet().WithDocID(rowTest.ID).WillReturn(kivikmock.DocumentT(t, rowTestAsJson))
	s, url, err := refDb.CommitURL([]byte(rowTest.LongUrl))
	if err != nil {
		panic(fmt.Sprintf("[USE-000-02T06] Malformed mock data %s", err))
	}
	assert.NotNil(t, s, "[USE-000-02T06] Short url not found")
	assert.Nil(t, url, "[USE-000-02T06] Short url generated not expected")
}

//T07
func TestCommitURLWhenColliding(t *testing.T) {
	firstRowTest := generateTestData()
	secondRowTest := generateTestData()
	secondRowTest.ID = ""
	firstRowTest.LongUrl = "www.google.test.com" //Fake url that simulates collision
	firstRowTestAsJson, _ := json.Marshal(firstRowTest)
	hashGenerator := hash([]byte(secondRowTest.LongUrl))
	_ = <-hashGenerator
	followingHash := genShortUrl(<-hashGenerator)
	refDb, testDB := genMockDB()
	testDB.ExpectGet().WithDocID(firstRowTest.ID).WillReturn(kivikmock.DocumentT(t, firstRowTestAsJson))
	testDB.ExpectGet().WithDocID(followingHash).WillReturnError(&kivik.Error{HTTPStatus: http.StatusNotFound, Err: nil})
	testDB.ExpectPut().WithDocID(followingHash).WillReturn("ok")
	s, url, err := refDb.CommitURL([]byte(secondRowTest.LongUrl))
	if err != nil {
		panic(fmt.Sprintf("[USE-000-02T07] Malformed mock data %s", err))
	}
	assert.Nil(t, s, "[USE-000-02T07] Short url found, expected not found")
	assert.NotNil(t, url, "[USE-000-02T07] Short url generated found, should return the following key")
	assert.Equal(t, followingHash, *url, "[USE-00-02T07] Generated url doesn't match expected")
}

