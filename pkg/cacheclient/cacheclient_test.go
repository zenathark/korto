package cacheclient

import (
	"github.com/alicebob/miniredis"
	"github.com/mediocregopher/radix/v3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestRedis() (*radix.Pool, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	c, err := radix.NewPool("tcp", mr.Addr(), 1)
	if err != nil {
		panic(err)
	}
	return c, mr
}

func TestGetLongUrl(t *testing.T) {
	input := "shorturl"
	expected := "longurl"
	c, s := newTestRedis()
	url, err := GetLongUrl(c, input)
	assert.Nil(t, err, "[USE-001-01T01] Error should not be returned if url is not found")
	assert.NotNil(t, *url, "[USE-001-01T1] Url should not be nil when url not found")
	assert.Empty(t, *url, "[USE-001-01T1] Url should be empty if not found")


	err = s.Set(input, expected)
	if err != nil {
		panic(err)
	}
	url, err = GetLongUrl(c, input)
	assert.Nil(t, err, "[USE-001-01T01] Error should not be returned if url is not found")
	assert.NotNil(t, *url, "[USE-001-01T1] Url should not be nil when url not found")
	assert.Equal(t, expected, *url, "[USE-001-01T1]")
}

//T01
func TestPutLongUrl(t *testing.T) {
	input := "shorturl"
	expected := "longurl"
	notExpectedRewrite := "weirdurl"
	c, s := newTestRedis()
	err := PutLongUrl(c, input, expected)
	assert.Nil(t, err, "[USE-006-00T01]")
	got, err := s.Get(input)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expected, got, "[USE-006-00T01]")

	err = PutLongUrl(c, input, notExpectedRewrite)
	assert.Nil(t, err, "[USE-006-00T01] Should not return error when saving key/value")
	got, err = s.Get(input)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expected, got, "[USE-006-00T01] Should not rewrite")

}

//T02
func TestForcePutLongUrl(t *testing.T) {
	input := "shorturl"
	expected := "longurl"
	expectedRewrite := "weirdurl"
	c, s := newTestRedis()
	err := ForcePutLongUrl(c, input, expected)
	assert.Nil(t, err, "[USE-006-00T02]")
	got, err := s.Get(input)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expected, got, "[USE-006-00T02]")

	err = ForcePutLongUrl(c, input, expectedRewrite)
	assert.Nil(t, err, "[USE-006-00T02] Should not return error when saving key/value")
	got, err = s.Get(input)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedRewrite, got, "[USE-006-00T02] Should rewrite")

}
