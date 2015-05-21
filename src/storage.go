package storage

import ( (
	"errors"
	"io"
	"io/ioutil"
)

// TestFunc2 is a test function
func TestFunc2() error {
	return errors.New("this is a test")
}

// Accessor is an interface for access to storage.
type Accessor interface {
	Create(fileName string) (io.WriteCloser, error)
	Open(fileName string) (io.ReadCloser, error)
	Delete(fileName string) error
}

// CreateFile retrieves io.WriteCloser for given Accessor and writes fileBody.
// io.WriteCloser is closed inside function.
func CreateFile(ac Accessor, fileName string, fileBody []byte) error {
	if ac == nil {
		return errors.New("invalid accessor")
	}

	wc, err := ac.Create(fileName)
	if err != nil {
		return err
	}

	_, err = wc.Write(fileBody)
	if closeErr := wc.Close(); closeErr != nil {
		return closeErr
	}

	return err
}

// ReadFile retrieves io.ReadCloser for given Accessor and read contents from it.
// io.ReadCloser is closed inside function
func ReadFile(ac Accessor, fileName string) ([]byte, error) {
	if ac == nil {
		return nil, errors.New("invalid accessor")
	}

	rc, err := ac.Open(fileName)
	if err != nil {
		return nil, err
	}

	fileBody, err := ioutil.ReadAll(rc)
	if err != nil {
		fileBody = nil
	}
	if closeErr := rc.Close(); closeErr != nil {
		return nil, closeErr
	}

	return fileBody, err
}
