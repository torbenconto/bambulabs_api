package ftp

import (
	"os"
	"time"
)

type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (f FileInfo) Name() string { return f.name }
func (f FileInfo) Size() int64  { return f.size }

// NOT IMPLEMENTED, DO NOT EXPOSE
func (f FileInfo) Mode() os.FileMode  { return os.FileMode(0) }
func (f FileInfo) ModTime() time.Time { return f.modTime }
func (f FileInfo) IsDir() bool        { return f.isDir }
func (f FileInfo) Sys() any           { return nil }
