package oauth

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/memcache"
)

func getVerifier(ctx context.Context, serviceName string, key string) ([]byte, error) {
	newKey := strings.Join([]string{oauthPrefix, serviceName, key}, verifierSeparator)
	item, err := memcache.Get(ctx, newKey)
	if err != nil {
		log.Printf("memcache get error: %+v", err)
		return nil, err
	}

	return item.Value, nil
}

func putVerifier(ctx context.Context, serviceName string, key string, ver Verifier) error {
	verBytes, err := getBytes(ver)
	if err != nil {
		return err
	}

	newKey := strings.Join([]string{oauthPrefix, serviceName, key}, verifierSeparator)
	item := &memcache.Item{
		Key:        newKey,
		Value:      verBytes,
		Expiration: verifierExpiration,
	}

	return memcache.Add(ctx, item)
}

func deleteVerifier(ctx context.Context, serviceName string, key string) error {
	newKey := strings.Join([]string{oauthPrefix, serviceName, key}, verifierSeparator)
	return memcache.Delete(ctx, newKey)
}

// getBytes converts any interface variables registered to gob into byte array
func getBytes(val interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/*
Random string generator.
Taken from: http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
*/
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
