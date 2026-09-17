package bambulabs_api

import (
	"io"
	"os"

	"github.com/torbenconto/bambulabs_api/internal/ftp"
)

type FileClient interface {
	List(path string) ([]os.FileInfo, error)
	Download(path string, w io.Writer) error
	Upload(path string, r io.Reader) error
	Delete(path string) error
}

type ftpFileClient struct {
	ftp *ftp.FtpClient
}

func newFTPFileClient(c *ftp.FtpClient) FileClient {
	return &ftpFileClient{ftp: c}
}

func (c *ftpFileClient) List(path string) ([]os.FileInfo, error) {
	if c.ftp == nil {
		return nil, ErrFTPUnavailable
	}
	return c.ftp.List(path)
}

func (c *ftpFileClient) Download(path string, w io.Writer) error {
	if c.ftp == nil {
		return ErrFTPUnavailable
	}
	return c.ftp.Retrieve(path, w)
}

func (c *ftpFileClient) Upload(path string, r io.Reader) error {
	if c.ftp == nil {
		return ErrFTPUnavailable
	}
	return c.ftp.Store(path, r)
}

func (c *ftpFileClient) Delete(path string) error {
	if c.ftp == nil {
		return ErrFTPUnavailable
	}
	return c.ftp.Delete(path)
}
