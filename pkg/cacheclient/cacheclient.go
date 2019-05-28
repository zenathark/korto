package cacheclient

import (
	"github.com/mediocregopher/radix/v3"
	"strings"
)

func GetLongUrl(c *radix.Pool, shortUrl string) (*string, error) {
	var longUrl string
	err := (*c).Do(radix.Cmd(&longUrl, "GET", shortUrl))
	if err != nil {
		return nil, err
	}
	return &longUrl, nil
}

func PutLongUrl(c *radix.Pool, shortUrl string, longUrl string) error {
	dbUrl, err := GetLongUrl(c, shortUrl)
	if err != nil {
		return err
	}
	if strings.Compare(*dbUrl, "") != 0 {
		return nil
	}
	err = (*c).Do(radix.Cmd(nil, "SET", shortUrl, longUrl))
	return err
}

func ForcePutLongUrl(c *radix.Pool, shortUrl string, longUrl string) error {
	err := (*c).Do(radix.Cmd(nil, "SET", shortUrl, longUrl))
	return err
}