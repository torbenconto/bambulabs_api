package commands

import (
	"os"

	"github.com/torbenconto/bambulabs_api/internal/ftp"
)

type FTP struct {
	client *ftp.Client
}

func CreateFTPInstance(ftpClient *ftp.Client) *FTP {
	return &FTP{client: ftpClient}
}

func (f *FTP) StoreFile(path string, file os.File) error {
	return f.client.StoreFile(path, file)
}

// ListDirectory calls the underlying ftp client to list the contents of a directory on the printer.
func (f *FTP) ListDirectory(path string) ([]os.FileInfo, error) {
	return f.client.ListDir(path)
}

// RetrieveFile calls the underlying ftp client to retrieve a file from the printer.
func (f *FTP) RetrieveFile(path string) (os.File, error) {
	return f.client.RetrieveFile(path)
}

// DeleteFile calls the underlying ftp client to delete a file from the printer.
func (f *FTP) DeleteFile(path string) error {
	return f.client.DeleteFile(path)
}
