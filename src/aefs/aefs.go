package aefs

import (
	"fmt"
	"io"
	"log"
	"strings"

	"appengine"
	"appengine/blobstore"
	"appengine/file"

	"golang.org/x/net/context"

	"github.com/tractrix/common-go/google"
)

// AppEngineFS provides virtual file system whose backend is Google Cloud Storage (GCS).
// Note that it uses Files API, which has been deprecated, for providing GCS emulation
// using the local disk because GCS client library for Go does not seem to support such
// emulation as of now.
//
// TODO: Remove this and use GCS (declared in storage_gcs.go) instead after GCS client
//       library supports working with development server.
type AppEngineFS struct {
	c appengine.Context

	Bucket string
}

// NewAppEngineFS returns a new AppEngineFS bound to the context and the bucket.
func NewAppEngineFS(context context.Context, bucketName string) *AppEngineFS {
	return &AppEngineFS{
		c:      google.ClassicContextFromContext(context),
		Bucket: bucketName,
	}
}

func (fs *AppEngineFS) bucketName() (string, error) {
	if fs == nil || fs.Bucket == "" {
		return file.DefaultBucketName(fs.c)
	}
	return fs.Bucket, nil
}

func (fs *AppEngineFS) makeFileNameAbsolute(fileName string) (string, error) {
	if strings.HasPrefix(fileName, "/gs/") {
		return fileName, nil
	}
	if strings.HasPrefix(fileName, "/") {
		return "", fmt.Errorf("makeFileNameAbsolute: unknown absolute filename pattern")
	}
	bucketName, err := fs.bucketName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/gs/%s/%s", bucketName, fileName), nil
}

// Create returns a io.WriteCloser of a file with the name.
func (fs *AppEngineFS) Create(fileName string) (io.WriteCloser, error) {
	wc, _, err := file.Create(fs.c, fileName, &file.CreateOptions{
		MIMEType:   "text/plain",
		BucketName: fs.Bucket,
	})
	if err != nil {
		log.Printf("Create: unable to create bucket %q, file %q: %v", fs.Bucket, fileName, err)
		return nil, err
	}

	return wc, nil
}

// Open returns io.ReadCloser of the file.
func (fs *AppEngineFS) Open(fileName string) (io.ReadCloser, error) {
	absFileName, err := fs.makeFileNameAbsolute(fileName)
	if err != nil {
		log.Printf("Open: unable to make absolute file name for bucket %q, file %q: %v", fs.Bucket, fileName, err)
		return nil, err
	}
	fr, err := file.Open(fs.c, absFileName)
	if err != nil {
		log.Printf("Open: unable to open read for bucket %q, file %q: %v", fs.Bucket, fileName, err)
		return nil, err
	}

	return fr, nil
}

// Delete deletes the file from the storage.
func (fs *AppEngineFS) Delete(fileName string) error {
	absFileName, err := fs.makeFileNameAbsolute(fileName)
	if err != nil {
		log.Printf("Delete: unable to make absolute file name for bucket %q, file %q: %v", fs.Bucket, fileName, err)
		return err
	}

	// NOTE: Appengine development server does not support file deletion through file service,
	//       thus blobstore service is used to delete file as a workaround.
	blobkey, err := blobstore.BlobKeyForFile(fs.c, absFileName)
	if err != nil {
		log.Printf("Delete: unable to generate blob key for bucket %q, file %q: %v", fs.Bucket, fileName, err)
		return err
	}

	return blobstore.Delete(fs.c, blobkey)
}
