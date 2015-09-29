package gcs

import (
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// GCS provides access to Google Cloud Storage (GCS)
type GCS struct {
	c context.Context
	Bucket string
}

// NewGCS returns a new NewGCS bound to the context and the bucket.
func NewGCS(context context.Context, bucketName string) *GCS {
	return &GCS{
		c:      context,
		Bucket: bucketName,
	}
}

// Create returns a io.WriteCloser of a file with the name.
func (gcs *GCS) Create(fileName string) (io.WriteCloser, error) {
	hc := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(gcs.c, storage.ScopeReadWrite),
			Base:   &urlfetch.Transport{Context: gcs.c},
		},
	}
	ctx := cloud.NewContext(appengine.AppID(gcs.c), hc)
	wc := storage.NewWriter(ctx, gcs.Bucket, fileName)
	wc.ContentType = mime.TypeByExtension(filepath.Ext(fileName))

	return wc, nil
}

// Open returns io.ReadCloser of the file.
func (gcs *GCS) Open(fileName string) (io.ReadCloser, error) {
	hc := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(gcs.c, storage.ScopeReadWrite),
			Base:   &urlfetch.Transport{Context: gcs.c},
		},
	}
	ctx := cloud.NewContext(appengine.AppID(gcs.c), hc)
	rc, err := storage.NewReader(ctx, gcs.Bucket, fileName)
	if err != nil {
		log.Printf("Open: unable to open read for bucket %q, file %q: %v", gcs.Bucket, fileName, err)
		return nil, err
	}

	return rc, nil
}

// Delete deletes the file from the storage.
func (gcs *GCS) Delete(fileName string) error {
	hc := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(gcs.c, storage.ScopeReadWrite),
			Base:   &urlfetch.Transport{Context: gcs.c},
		},
	}
	ctx := cloud.NewContext(appengine.AppID(gcs.c), hc)
	return storage.DeleteObject(ctx, gcs.Bucket, fileName)
}
