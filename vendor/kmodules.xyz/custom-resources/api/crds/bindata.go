// Package crds Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// appcatalog.appscode.com_appbindings.v1.yaml
// appcatalog.appscode.com_appbindings.yaml
package crds

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _appcatalogAppscodeCom_appbindingsV1Yaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x5a\x5f\x6f\xdb\x38\xb6\x7f\xcf\xa7\x38\xf0\x3c\xa4\x05\x6c\x19\x73\xe7\xe5\xc2\xf7\x61\x6e\x9a\x66\x80\xee\xa4\x69\x11\x67\xba\x18\x6c\x17\x1b\x4a\x3a\xb6\x39\xa1\x48\x0d\x49\xd9\xf1\x16\xfd\xee\x8b\x73\x48\x59\xb2\x2d\xc9\xce\x74\xab\x87\x36\x16\xa9\xc3\xf3\xf7\x77\xfe\x48\xa2\x94\x9f\xd0\x3a\x69\xf4\x0c\x44\x29\xf1\xd9\xa3\xa6\x5f\x2e\x79\xfa\x5f\x97\x48\x33\x5d\xff\x78\xf1\x24\x75\x3e\x83\xeb\xca\x79\x53\xdc\xa3\x33\x95\xcd\xf0\x2d\x2e\xa4\x96\x5e\x1a\x7d\x51\xa0\x17\xb9\xf0\x62\x76\x01\x90\x59\x14\x74\xf3\x41\x16\xe8\xbc\x28\xca\x19\xe8\x4a\xa9\x0b\x00\x25\x52\x54\x8e\xf6\x00\x88\xb2\x4c\x9e\xaa\x14\xad\x46\x8f\x7c\x8a\x16\x05\xce\x20\x13\x5e\x28\xb3\xbc\x00\x08\xbf\x45\x59\xa6\x52\xe7\x52\x2f\x5d\x22\xca\x32\x2e\xd3\x9f\x2e\x33\x39\x26\x99\x29\x2e\x5c\x89\x19\x51\x5d\x5a\x53\x95\xfc\x48\xe7\xb6\x40\x32\x9e\x9f\x09\x8f\x4b\x63\x65\xfd\x7b\xd2\x3a\x99\x7e\xd5\x4f\xd6\x3f\x59\x00\x80\xa0\x87\xab\xb2\x7c\x13\x98\xe2\x9b\x4a\x3a\xff\xeb\xc1\xc2\xad\x74\x9e\x17\x4b\x55\x59\xa1\xf6\x04\xe1\xfb\x4e\xea\x65\xa5\x84\x6d\xaf\x5c\x00\xb8\xcc\x94\x38\x83\x3b\xe2\xb4\x14\x19\xe6\x17\x00\xeb\x60\x1d\xe6\x74\x02\x22\xcf\x59\xe9\x42\x7d\xb4\x52\x7b\xb4\xd7\x46\x55\x85\xde\xc9\xf1\x87\x33\xfa\xa3\xf0\xab\x19\x24\xa4\x98\xc4\x6f\xcb\x20\x45\xad\xd2\x87\xe6\x06\xad\xcd\xc0\x79\x5b\x8b\x72\xfc\x78\x3c\x7c\x8f\xc2\xa7\xbd\x7b\xc3\x44\x6a\xd7\x48\x8e\xfc\x62\x8f\xe4\xd5\x72\x9f\xa7\x5c\xf8\x70\x23\x2c\xaf\x7f\x14\xaa\x5c\x89\x1f\x83\xea\xb2\x15\x16\x62\x16\xf7\x9b\x12\xf5\xd5\xc7\x77\x9f\x7e\x9a\xef\xdd\x06\x28\xad\x29\xd1\xfa\x9d\x89\xc3\xd5\xf2\xf6\xd6\x5d\x80\x1c\x5d\x66\x65\xe9\x39\x0c\x2e\x89\x60\xd8\x05\x39\xb9\x39\x3a\xf0\x2b\xac\x2d\x81\x79\xe4\x01\xcc\x02\xfc\x4a\x3a\xb0\x58\x5a\x74\xa8\x3d\x8b\xb8\x47\x18\x68\x93\xd0\x60\xd2\x3f\x30\xf3\x09\xcc\xd1\x12\x19\x70\x2b\x53\xa9\x1c\x32\xa3\xd7\x68\x3d\x58\xcc\xcc\x52\xcb\x7f\xef\x68\x3b\xf0\x86\x0f\x55\xc2\x63\x74\xa6\xe6\x62\xcb\x6b\xa1\x60\x2d\x54\x85\x63\x10\x3a\x87\x42\x6c\xc1\x22\x9d\x02\x95\x6e\xd1\xe3\x2d\x2e\x81\xf7\xc6\x22\x48\xbd\x30\x33\x58\x79\x5f\xba\xd9\x74\xba\x94\xbe\x8e\xf2\xcc\x14\x45\xa5\xa5\xdf\x4e\x33\xa3\xbd\x95\x69\xe5\x8d\x75\xd3\x1c\xd7\xa8\xa6\x4e\x2e\x27\xc2\x66\x2b\xe9\x31\xf3\x95\xc5\xa9\x28\xe5\x84\x59\xd7\x9e\xa1\xa2\xc8\x7f\xb0\x11\x17\xdc\xe5\x1e\xaf\x47\xee\x11\x2e\x8e\xa4\x01\x0b\x50\x40\x81\x74\x20\xe2\xa3\x41\x8a\x46\xd1\x74\x8b\xb4\x73\x7f\x33\x7f\x80\xfa\x68\x36\xc6\xa1\xf6\x59\xef\xcd\x83\xae\x31\x01\x29\x4c\xea\x05\xda\x60\xc4\x85\x35\x05\xd3\x44\x9d\x97\x46\x6a\xcf\x3f\x32\x25\x51\x1f\xaa\xdf\x55\x69\x21\x3d\xd9\xfd\xcf\x0a\x9d\x27\x5b\x25\x70\x2d\xb4\x36\x1e\x52\x84\xaa\x24\xff\xcd\x13\x78\xa7\xe1\x5a\x14\xa8\xae\x85\xc3\xef\x6e\x00\xd2\xb4\x9b\x90\x62\xcf\x33\x41\x1b\xb5\x0f\x37\x07\xad\xb5\x16\x6a\x90\xed\xb1\x57\x83\x7c\xf3\x12\x33\x32\x1c\xe9\x8e\x1e\x82\x85\xb1\x84\x71\x7b\xcf\x76\xc7\x26\x5d\x41\xdd\xd7\x46\x2f\xe4\xf2\x70\xed\xe0\xcc\xeb\xd6\xd6\x5d\x98\xae\xcc\x86\x02\x27\x2a\x93\x60\x1e\x36\xd2\xaf\x98\x1d\x4a\x3a\x47\x24\x01\xee\xf1\xcf\x4a\x5a\x86\xda\xfd\xab\x9f\x4b\xe6\x54\xbc\xa9\x74\xae\xb0\x6b\xed\x90\xd3\xab\xb0\x35\x38\xf4\xc7\x9b\xf7\x80\x9a\xb2\x4b\x0e\xd7\x57\x90\x86\xa5\xcd\x4a\x66\x2b\xd8\x48\xa5\x20\xc5\x4e\x92\x00\x95\xc3\x9c\xa4\x5b\x0b\x25\xc9\xc3\x82\x92\xd1\xae\x29\x1a\x32\x62\x75\x11\x44\xae\x71\xa9\x47\x62\x20\xa3\x14\xc2\xcf\x20\xdd\xfa\xee\xc3\x7a\x7c\xa6\xbe\xa4\x76\x98\x55\x16\xe7\x4f\xb2\x7c\xb8\x9d\x7f\x42\x2b\x17\xdb\x33\x34\xf1\xae\xeb\x39\xc8\xa5\x13\xa9\x42\x07\x0f\xb7\xf3\x3d\x39\xd6\xb4\x4e\x7f\x1e\xa3\x6a\x7d\x6d\x56\xa8\x5b\xe6\x26\x4d\x44\x83\x47\xf9\xe1\x81\xfe\x92\x8e\x84\x31\x7a\xa9\xf8\xb8\xcc\x54\x56\x2c\x29\x44\xe1\x77\x53\xf5\x90\x8e\x10\x5d\xb9\xa0\xe8\xc6\x8a\xda\x79\x14\x79\xb7\x66\x83\xe2\x52\x63\x14\x8a\x2e\x9e\xd9\x5c\xd9\x39\x5e\x33\x7a\x8c\x7b\x1f\x83\xdf\x58\x5c\xa0\x45\x4d\x30\x67\x1a\xcb\x67\xc8\x11\xd6\x81\x7c\xf5\xc5\x4a\xb8\x91\x7e\x85\x16\x1a\x92\xc6\xc2\x63\x65\xd5\x23\x14\x95\x63\xd0\xa2\x60\x95\x0b\x49\x3a\xf9\xac\xe1\x1d\x79\x50\x9f\x1f\x6e\x30\x5d\x19\xf3\x44\x6c\xd9\x4a\xeb\x5a\xe7\x52\x47\xc4\xac\x9c\x47\x3b\xa6\x1f\x1a\xb6\xa6\x6a\x2b\x72\xc7\x40\x32\xea\x24\x3e\x1c\x73\x50\x57\x04\x3d\x6b\x87\x59\xe4\x91\x36\x3f\xd6\x70\x44\x3f\x42\x68\xec\x74\x97\xec\xa2\xff\xb2\x97\xe4\x89\x50\x60\xae\xa9\xd8\x39\x97\x27\xda\x1c\x4c\xaa\xc1\x94\xa1\x96\x83\xdf\xee\x6f\x99\xca\x59\x38\x00\xec\x47\xda\x83\xd4\x20\xf4\xb6\x4e\x43\xc1\x2f\xc8\xd3\xa3\x70\xdf\x26\x93\xb1\xfe\x4c\x99\x1e\x56\xc8\xdb\xc1\xaf\x84\xaf\x79\x07\x7c\x2e\x0d\x01\x56\xba\x3d\x01\x46\xd0\x02\x24\xa9\xfd\x4f\xff\x73\x82\x6d\x2a\x7e\x96\x68\x7b\x76\xfd\x59\xa1\xed\x81\xa2\x23\xc6\x2f\x1f\x79\x37\x5b\x63\x67\x8a\x1a\x9b\x79\x29\xea\x68\xcc\x0e\x6e\xaa\xc3\x42\xa0\x7d\x5d\x5e\xfe\x7c\x79\xb9\x6f\xbf\xef\x6f\x25\x2e\x16\xcf\x8e\x87\x79\x8c\x71\x17\xd9\x0c\x4f\x13\x47\x95\xc3\x31\x03\x09\x3e\x8b\xa2\xec\xcb\x6a\xe1\xa2\xe2\x65\x1c\x4a\x18\xc2\x89\x1d\x70\xc4\x88\x97\xd1\x05\x44\x59\x2a\x89\x39\x08\x07\xa5\xc5\x85\x7c\x1e\x20\xc9\xd0\x41\x35\x58\x74\x83\x28\xd6\x74\x4a\x07\x50\x55\x75\x78\x88\x36\x84\x37\x7d\x5a\xa1\xab\x36\x41\x38\xfb\x9b\x02\xdc\x46\x8c\xe8\x56\xca\x84\x81\xa5\x67\x89\xc2\xa2\x67\x29\xc8\x38\x90\x44\x8e\x8a\xb0\xfa\xaa\xac\x3a\x2b\x7f\x30\xbe\x2f\xe5\x3a\xb6\x2f\xca\x84\x4c\x5a\x63\xa0\x28\xcb\x31\x69\xde\x79\xa1\x73\x61\x8f\x0b\xa0\x70\x11\x34\x91\x5d\xe0\xd5\xe3\x3f\x76\x76\xf9\xe7\xca\x38\x3f\x23\xe9\xa6\x8c\x67\xaf\x13\xb8\x79\x16\x99\x57\x5b\x30\x9a\x51\x96\x4f\xef\x21\x69\xda\x99\xa8\x3b\x01\x11\xa6\x3c\xd2\x21\x8f\x75\xfa\x20\x37\xe0\x1c\xd8\x43\xd4\x1b\xea\x16\x62\x4e\xac\xf3\xd2\x7e\x4e\xfa\xbf\x5d\x32\x6f\x8e\x5f\x48\x54\x7d\xa2\xd7\x99\x9e\xb9\x21\x66\xa0\x90\xcb\x15\x73\x4b\x3d\x87\x5a\x53\x7b\x25\x05\xe0\x73\x6c\xc7\xde\xde\xcd\x59\xa3\xa6\xc7\xb0\xdc\x80\xba\xd8\x7f\xbc\xc2\x64\x99\x8c\xe1\xf1\xa9\x4a\x71\xb2\xbb\xff\x08\x59\x68\x24\xe2\x09\x20\xf5\x24\xb2\xdf\x43\x92\x0e\xa5\x7e\x91\xc1\x97\x55\x95\x22\x08\x50\x62\x8b\xa1\x75\x92\x46\xb1\xe1\x5f\x27\xb5\x4a\xa9\xf5\x11\xca\x99\x1e\x8a\xf4\xbc\x86\x77\x1f\x41\xe4\xb9\x45\xe7\xd8\x22\x57\x21\x41\xb5\xa0\x32\xf4\x9d\x72\x01\xb1\xb7\x22\xb2\x43\x14\x6b\x34\x85\x12\x6d\x21\x9d\x93\x29\x57\x53\x20\xc8\xc7\x12\xaa\xc4\x98\xb1\xda\x46\x7c\x9c\xef\xe3\xb1\x14\x8e\x53\xa8\xb0\xa9\xf4\x56\xec\xa0\xba\xae\x8e\xd8\xbb\x5b\x88\x36\x06\x01\xc3\x7a\x94\x39\x75\x53\x0b\x89\x36\xc8\xeb\x3d\x16\xa5\x8f\x24\x89\x29\x41\xff\x5a\xf2\xde\x54\x38\x99\x81\xa8\xfc\x0a\xc8\x88\xf0\x79\x44\x2b\x33\xe2\x69\x63\x6c\xfe\xff\x9f\xbb\xab\x1b\x20\xed\x91\x6d\x85\x52\x66\x43\x9e\xfe\x8b\x15\xcb\x82\xda\x52\x78\xf5\x79\xf4\x43\x92\x24\x9f\x47\xaf\x59\xab\x21\xfb\x94\xc2\x8a\x02\x3d\x7b\xcb\xe7\xd1\xcf\x61\xbd\xcf\xb3\x2c\xb6\x69\x8f\x01\xb9\xe6\xeb\x29\xb4\x06\x41\x6f\x00\x7f\x1a\x8e\x4e\xb4\x67\xa3\x8f\x0d\xef\xa1\x91\x47\x5f\x23\x4f\x4b\x2c\x6f\xb8\x63\x0e\x9d\x4d\x57\x9b\x65\xb4\xa6\x06\xbe\xb1\x6a\x88\x46\xa9\x95\xd4\x08\xbf\x5f\xbd\xbf\x9d\xfe\x6d\xfe\xe1\x0e\x4a\xb1\x55\x46\xe4\x91\xa0\xb7\x42\x3b\x45\x5d\x78\x67\xf7\xe2\x0d\x10\xa6\xaf\x85\x22\xb7\xe5\xe7\xeb\x01\x4d\xc4\x9e\x16\xf7\x8c\x10\x24\xc3\xdd\x87\x07\x70\x98\x59\xec\x02\x65\x63\x21\xf4\x36\x79\x9d\xf0\x37\x14\x64\x3a\xaf\xf1\xeb\xee\xe6\xd3\xcd\x7d\x4b\x58\x58\x19\x95\x53\x85\xe0\xa4\x97\xeb\x2e\xbc\x90\x3a\xe4\x43\x69\x74\x02\x0f\x86\x35\xd8\x56\x1d\x05\x7c\x66\xb4\x17\x04\x39\xcc\x57\xfb\x91\x71\x07\xc5\x56\x35\x7e\x75\xfb\xf7\xab\xdf\xe7\xe0\xbc\xb1\x18\x48\xb5\x9e\x0d\x51\x39\x67\x9a\x1d\x0e\x34\x98\x9f\x9e\x27\xcd\x64\x77\x82\x45\x8a\x79\x8e\xf9\xa4\x9e\xd1\xcc\xc0\xdb\xea\x58\xd8\xbd\x87\x18\x4e\xec\x1a\x27\x95\x7e\xd2\x66\xa3\x27\x6c\x01\xd7\xf9\x68\x90\xfb\x84\x2f\xce\xa3\x72\xba\xfa\x00\x5e\xf1\x26\x0c\xae\xb1\x4e\x18\xcd\x40\xe3\xb2\xab\xaf\xd2\xf5\x80\xb6\x55\xf2\xb2\x39\x39\xd9\x58\x64\x24\x11\xca\x81\x70\xce\x64\x92\xfc\xb0\x99\x43\x34\xb4\x8f\xeb\xe1\xe1\xfe\xa7\xbf\xf7\xd9\xaf\xf3\xee\x5a\x12\xc6\xb6\xd1\x77\xce\x9f\xf6\x67\xf0\xb9\xc9\xdc\x34\x33\x3a\xc3\xd2\xbb\xa9\x59\x53\x8a\xc4\xcd\x74\x63\xec\x93\xd4\xcb\x09\x09\x30\x09\x46\x77\x3c\xaf\x77\xd3\x1f\xf8\xbf\x1e\x40\x7a\xf8\xf0\xf6\xc3\x0c\xae\xf2\x1c\x0c\x37\x9f\x95\xc3\x45\xa5\x42\x38\xb9\xa4\x35\x8a\x1d\xf3\x38\x70\x0c\x95\xcc\x7f\xee\x2e\xd3\xfe\x2a\x5a\x05\xf3\x3e\x10\x18\x90\x6f\x9f\xc2\xac\x5b\xe9\x02\x46\xd5\x0f\x70\x30\xc4\x48\x8b\x71\x93\xe2\xae\xb2\x0d\x98\xd4\x05\x5a\x27\x3c\x60\x1e\x8a\x8f\xe8\x05\x90\xe2\x22\x04\x21\x6e\x19\xc5\xa5\x76\x68\x07\xa0\x2b\x90\xe0\xd8\x3c\xda\x22\x3d\x76\x89\x79\xd4\x09\xec\x2b\x26\x22\xb4\xd4\x4b\x85\x07\xd2\x47\x6c\xe8\x36\xf2\xbe\x26\xf6\xe4\xb6\xe8\x2b\xab\x31\x6f\xe6\xaa\xa9\x35\x4f\x68\xdb\xd2\x76\xd3\x6c\x69\xe0\x50\xde\x33\xb4\xd9\xdd\x63\xbe\xc1\x4c\x50\x0a\xcf\xe5\x22\x84\x43\xe4\x86\x7a\x13\xb3\x96\x79\x3d\x4f\x76\x14\x39\xe4\x50\xe4\x06\x75\x31\xd9\x57\xd6\xa0\xc8\x56\x51\x4e\x10\x2d\xd2\x6d\x35\x38\x6f\x2b\x1e\xd9\x8e\xb9\x78\x70\x54\xdd\xc5\x52\xb7\x9b\x28\x71\xf1\x22\xff\xab\x55\x43\xf5\x6f\x2e\xca\xee\x76\x43\x7a\x07\xa8\xbd\xa5\xde\xcf\x1b\xd8\xac\x84\xc7\x35\x4f\xbe\x9b\x39\x52\x66\xb4\xab\x0a\xa4\x8a\xa9\xa4\x18\x4f\xe0\x97\x56\xf9\xd4\xcb\x6c\xa7\xd1\xb9\xe9\xdf\x99\x3c\x4c\xda\x33\x55\xe5\xa1\xb2\x7b\xc2\x2d\x8c\x7e\x9b\xdf\xdc\xdf\x5d\xbd\xbf\x19\x75\x93\x4e\xab\x38\x80\xaf\xb9\x8a\x5d\x58\xc0\x70\xd2\x25\xe3\x78\x48\xf7\xf5\xac\xa1\xd2\x79\x10\xaa\x93\x24\x1f\xfb\xf6\xcd\xbf\xe8\xe4\x51\xab\xb8\x37\xb0\x12\x6b\x6c\xfb\x12\x5c\x87\xd7\x81\x8d\x25\x7a\x89\x06\xed\x73\x5b\x0a\x0b\x43\xb5\x17\xf9\xd2\x61\x7c\x1d\x35\x39\x94\x68\x0e\x1c\x97\x5f\xb8\x1d\x20\x56\x5f\xcb\xf9\x65\x64\x91\xe4\xff\x15\xb7\xa3\x19\x7c\x19\x51\x90\x8d\x66\x6d\xa5\xc2\xc8\x1b\xba\x53\xcb\xfb\xf5\x2b\x7c\xd0\xa1\x3d\xeb\xa4\x19\xd3\xc5\x01\xe3\x97\x97\x0e\x0a\x4a\xe2\xf1\x7d\xc9\x5e\x9f\xd6\x85\xd5\xa7\x06\x78\x22\xcf\x7f\xc5\xde\xf9\xcc\xfe\x4b\x05\xde\xda\x7a\x75\x03\x62\xcf\x1e\xc2\x13\xb5\xd0\x04\xec\xde\x8a\xf6\x76\xf9\x64\xfc\x0e\x98\x9a\xf7\xd5\x73\xe7\x08\x43\x57\xfd\xb2\xf3\xe6\x99\xd8\x3c\x7e\xb3\x38\x20\xe0\x25\x15\x9a\x54\x7c\xd2\xf3\xe4\xc3\x91\xc0\x38\x26\x6e\x57\x29\x4e\x47\x3c\xac\x19\x20\xda\x0c\x3b\x04\xd5\x5a\x07\xb8\xb0\x8b\x89\x96\xf3\x3d\xe1\x96\xa3\x7b\x90\xe8\xae\x71\x5a\xca\x35\xea\x03\x07\x6f\xe9\x70\x06\x5f\x60\xb4\x30\xe4\x6d\x5f\x60\x94\x0a\x3b\x1a\xd2\x00\xf0\x5e\xda\x05\x5f\xe1\x2b\x17\xc8\x44\xf9\x58\x8d\x30\xfa\x92\x2c\x8c\x49\x52\x61\xbf\xf6\x60\x44\x7d\xf1\xcb\x59\x7e\x41\xb8\xa3\xbd\x9b\xc0\x51\x7d\x9b\xef\xe6\x00\xe7\x5a\x3e\x5c\x7d\xaa\xeb\x9f\x25\x9d\x35\xb4\x63\x67\x3c\xdb\x4b\x1e\x0e\x4a\xd6\xe8\xc8\x22\xef\x9b\x5a\x9c\xcd\x45\x58\xfe\x44\x8a\x7b\x11\x37\xb1\xd1\x7e\xa5\x8d\x9e\xa4\x52\x0b\xbb\x7d\x1d\xd5\x1f\xf8\xea\x2f\x8c\x9a\xeb\x84\x73\x7e\xab\x68\xeb\x17\x0b\x15\x04\x89\x72\xbc\x2a\x0d\x8f\x27\xb6\x40\x32\x86\xb3\x5e\x9f\xd6\x3a\x9c\x1b\x7a\xef\x16\x90\x1a\xbf\x8a\xa7\x09\x3d\x4c\xb4\x65\x27\xae\x8e\x0e\x87\xa1\x81\x8a\x74\x20\x97\x9a\xbd\x9d\xbb\xce\xe6\xa1\x41\xe2\xfc\x66\x8c\x9e\x1a\xd2\xf9\xc9\xf7\x85\x51\xfa\xd3\xa6\x19\x9e\xa5\xb6\xbf\x1e\x69\x90\xa0\x77\xeb\x53\x4f\x09\xc9\xe3\xd5\x93\xf2\x4f\x82\xe2\xfa\xc6\x89\xc3\x33\xd8\x3a\x93\xb9\x5f\xac\xe9\xc9\xd2\x9d\xe9\x8c\xf7\x0f\xe6\xb4\x02\xed\xb2\xb7\xa7\x02\x10\x4a\xc5\xaf\x14\x42\x35\x17\x3e\x2f\xc1\x67\xe9\x7c\x53\x78\x34\x75\x73\x0b\xf1\x7a\x49\x7e\x73\x0e\x0c\x45\xcb\x3d\x2e\x5e\x14\x71\x47\x2f\x34\xeb\x42\x76\xaf\xda\x1d\x74\x5e\xd6\x55\xde\x29\x6d\x6f\x87\xf4\x32\xd1\xe0\xe4\x2b\xc7\x0e\xe9\x3a\xfb\xef\x13\x04\xce\x82\x35\x68\xcf\x1e\x5e\xcc\x52\x98\x58\x7c\x1f\xbe\x4e\x86\xcb\x19\x5b\x2c\x16\x66\x8d\xe7\x96\x87\xf7\xf5\xee\xc1\x68\x0a\x34\x1d\x88\x5e\xde\x8f\x7d\x86\x63\xab\x0f\x59\xce\x71\x9a\x97\xa6\xf7\x98\xd2\x03\xaf\x4d\xc3\x7c\xb2\x42\xf9\xaf\xe0\x6d\x3f\x88\x9e\x61\xb0\xd8\x83\x9c\x69\xb0\xb8\xfb\x84\xc1\xd8\xc1\xff\x82\xc1\x2e\xdd\x80\x2c\xe7\x98\x6d\x31\x00\xe5\x47\xd2\xf4\x94\x65\x81\xfd\x6f\x2d\x5f\xbc\x79\x19\x1f\xb8\x09\xbc\x84\x4f\x41\x70\x40\x0f\x67\xb2\x70\xda\x6d\x48\x59\xbd\x8b\xbd\x6f\x6d\x4e\xb8\xd4\xe0\x72\x58\x14\xd6\x1e\x8d\x63\x78\x65\x78\xb4\xf7\xb0\x2d\x9b\xc9\xfb\x42\x64\x52\x49\x2f\x3c\x92\x5f\x2c\xad\x28\x0a\xe1\x65\x06\x2b\xa1\x73\x45\x59\x94\x92\x6a\x59\xaa\xf8\xe9\xd1\x31\x44\x0e\x68\x70\xdd\xf5\x79\xe9\x11\x3b\xf5\xe7\xa5\xdf\x9f\xa3\x6e\x4b\x4e\xf6\xbe\xb4\xbb\x38\x69\x83\xa3\x9b\x3c\xa0\xcf\x5b\x23\x79\x2a\x22\xc5\xb2\x3d\xdf\x77\x55\xba\xfb\x28\x74\x06\x5f\xbe\x5e\xfc\x27\x00\x00\xff\xff\xba\x4e\xa7\x08\x6b\x2e\x00\x00")

func appcatalogAppscodeCom_appbindingsV1YamlBytes() ([]byte, error) {
	return bindataRead(
		_appcatalogAppscodeCom_appbindingsV1Yaml,
		"appcatalog.appscode.com_appbindings.v1.yaml",
	)
}

func appcatalogAppscodeCom_appbindingsV1Yaml() (*asset, error) {
	bytes, err := appcatalogAppscodeCom_appbindingsV1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "appcatalog.appscode.com_appbindings.v1.yaml", size: 11883, mode: os.FileMode(420), modTime: time.Unix(1573722179, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _appcatalogAppscodeCom_appbindingsYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x5a\xdd\x6f\xdb\xb8\xb2\x7f\xcf\x5f\x31\xf0\x3e\xa4\x05\xfc\x81\xde\x7d\xb9\xf0\x7d\xd8\x9b\xa6\x29\xd0\xdd\x36\x2d\xe2\x6c\x0f\x16\xa7\x07\x27\x94\x38\xb2\xb8\x91\x48\x2d\x49\xd9\xf1\x29\xfa\xbf\x1f\x0c\x49\x7d\xd8\xa6\x64\xa7\xbb\xeb\x87\x36\x16\xc9\xe1\xcc\x6f\xbe\x47\x66\x95\xf8\x8c\xda\x08\x25\x97\xc0\x2a\x81\x4f\x16\x25\x7d\x33\xf3\xc7\xff\x35\x73\xa1\x16\x9b\x57\x09\x5a\xf6\xea\xe2\x51\x48\xbe\x84\xeb\xda\x58\x55\xde\xa1\x51\xb5\x4e\xf1\x0d\x66\x42\x0a\x2b\x94\xbc\x28\xd1\x32\xce\x2c\x5b\x5e\x00\xa4\x1a\x19\x3d\xbc\x17\x25\x1a\xcb\xca\x6a\x09\xb2\x2e\x8a\x0b\x80\x82\x25\x58\x18\xda\x03\xc0\xaa\x6a\xfe\x58\x27\xa8\x25\x5a\x74\x57\x49\x56\xe2\x12\x52\x66\x59\xa1\xd6\x17\x00\xfe\x3b\xab\xaa\x44\x48\x2e\xe4\xda\xcc\x59\x55\x85\x65\xfa\xd3\xa4\x8a\xe3\x3c\x55\xe5\x85\xa9\x30\x25\xaa\x8c\x73\xc7\x0e\x2b\x3e\x69\x21\x2d\xea\x6b\x55\xd4\xa5\x74\x37\xce\xe0\xe7\xd5\xc7\xdb\x4f\xcc\xe6\x4b\x98\xd3\x81\xb9\xdd\x55\xe8\x58\xf1\x17\xdd\x37\x5f\xe9\xf9\x12\x8c\xd5\x42\xae\xa3\x07\x37\x1e\xb1\xde\xd9\xcf\xbd\x27\x63\xc7\x1b\x98\xe6\x47\x18\xf5\x88\x5d\xad\xfb\x7c\x70\x66\xe9\xeb\x5a\xab\xba\x72\x68\x44\x11\xf0\x67\x03\xb4\x29\xb3\xb8\x56\x5a\x34\xdf\x67\x3d\x50\xe9\x5b\x73\xb2\xf9\xea\x74\x03\xe0\x55\x7c\x55\x55\xaf\x3d\xde\xee\x61\x21\x8c\xfd\xe5\x60\xe1\xbd\x30\xd6\x2d\x56\x45\xad\x59\xb1\xa7\x23\xf7\xdc\x08\xb9\xae\x0b\xa6\xfb\x2b\x17\x00\x95\x46\x83\x7a\x83\xbf\xca\x47\xa9\xb6\xf2\xad\xc0\x82\x9b\x25\x64\xac\x30\xc4\x8b\x49\x15\x09\x7c\x4b\x82\x54\x2c\x45\x4e\xcf\xea\x44\x07\x6b\x33\x4b\xf8\xfa\xed\x02\x60\xc3\x0a\xc1\x1d\x78\x5e\x3a\x55\xa1\xbc\xfa\xf4\xee\xf3\x8f\xab\x34\xc7\x92\xf9\x87\x74\x99\xaa\x50\xdb\x16\x04\x6f\x73\xad\xb5\xb7\xcf\x00\x38\x9a\x54\x8b\xca\x51\x84\x4b\x22\xe5\xf7\x00\x27\xfb\x46\x03\x36\x47\x08\x3a\x47\x0e\xc6\x5d\x03\x2a\x03\x9b\x0b\x03\x1a\x9d\x58\xd2\x3a\x96\x7a\x64\x81\xb6\x30\x09\x2a\xf9\x1d\x53\x3b\x87\x15\x89\xae\x0d\x98\x5c\xd5\x05\x87\x54\xc9\x0d\x6a\x0b\x1a\x53\xb5\x96\xe2\x3f\x2d\x65\x03\x56\xb9\x2b\x0b\x66\x31\x00\xdd\x7c\x9c\x51\x4b\x56\x10\x08\x35\x4e\x81\x49\x0e\x25\xdb\x81\x46\xba\x03\x6a\xd9\xa3\xe6\xb6\x98\x39\x7c\x50\x1a\x41\xc8\x4c\x2d\x21\xb7\xb6\x32\xcb\xc5\x62\x2d\x6c\xe3\xdf\xa9\x2a\xcb\x5a\x0a\xbb\x5b\xa4\x4a\x5a\x2d\x92\xda\x2a\x6d\x16\x1c\x37\x58\x2c\x8c\x58\xcf\x98\x4e\x73\x61\x31\xb5\xb5\xc6\x05\xab\xc4\xcc\x31\x2e\xad\x0b\x12\x25\xff\xa1\x55\xcf\x65\x8f\xd3\x03\x1f\xf0\x1f\x67\x5f\x83\xb8\x93\x91\x81\x30\xc0\xc2\x31\xcf\x7f\x07\x2f\x3d\x22\x54\xee\x6e\x56\xf7\xd0\x5c\xea\x54\xb0\x8f\xb9\x43\xbb\x3b\x66\x3a\xe0\x09\x28\x21\x33\xd4\x5e\x71\x99\x56\xa5\xa3\x88\x92\x57\x4a\x48\xeb\xbe\xa4\x85\x40\xb9\x0f\xba\xa9\x93\x52\x58\xd2\xf4\x1f\x35\x1a\x4b\xfa\x99\xc3\x35\x93\x52\x59\x48\x10\xea\x8a\x5c\x94\xcf\xe1\x9d\x84\x6b\x56\x62\x71\xcd\x0c\xfe\xed\xb0\x13\xc2\x66\x46\x90\x9e\x06\xbe\x1f\x9c\xf7\x37\x7a\xb4\xda\xc7\x4d\x1c\x8d\x6a\xa8\xf3\xff\x55\x85\x29\xa9\x8a\xf0\xa2\x23\x90\x29\x4d\x9e\xde\x3b\x19\xf3\x3e\x17\x9a\x1c\xbc\xd7\x4a\x66\x62\xbd\xbf\x72\x70\xdb\x75\x6f\x63\xeb\x88\xb9\xda\x92\x73\x04\xf0\x28\xcc\xc1\x56\xd8\xdc\x31\x42\xf9\x04\xee\xf0\x8f\x5a\x68\x17\x39\xfa\x9f\x21\x6e\x1c\x47\xec\x75\x2d\x79\x81\xc7\x2b\x87\x1c\x5d\xf9\x8d\xde\x48\x3f\xdd\x7c\x00\x94\x14\x45\x39\x5c\x5f\x41\xe2\x97\xb6\xb9\x48\x73\xd8\x8a\xa2\x70\x96\x61\x8e\x38\x09\xe0\xab\x26\x8a\xa1\x07\x11\xf5\x86\xec\x3b\x25\x26\x33\x2f\x58\x13\x5f\x48\xae\x08\x91\x4c\xe9\x92\xd9\x25\x24\x3b\x8b\x91\xe5\xa8\x1d\x34\x1f\x21\x0d\xa6\xb5\xc6\xd5\xa3\xa8\xee\xdf\xaf\x3e\xa3\x16\xd9\xee\xa4\xfc\xef\x62\xa7\x80\x0b\xc3\x92\x02\x0d\xdc\xbf\x5f\xed\xf1\xbf\xa1\x75\xfa\xf3\x30\x2a\x36\x9f\x6d\x8e\xb2\xa7\x4a\x92\x3f\x28\x33\x48\x0d\xf7\xf4\x97\x30\x24\x86\x92\xeb\xc2\x5d\x96\xaa\x5a\xb3\x35\xb9\x1b\xfc\xa6\xea\x28\xe1\x10\x60\x6b\xe3\xc1\xed\xf4\x26\x8d\x45\xc6\x63\x68\x7a\xb8\x12\xa5\x0a\x64\xc7\xdc\x3a\xf5\xa4\xa7\x2d\x64\xf2\x10\x76\x3e\x78\x1b\xd1\x98\xa1\x46\x49\x61\x4a\x75\x7a\x4e\xd1\xf9\xcb\x98\x72\x01\x6e\x84\xcd\x51\x43\x47\x50\x69\x78\xa8\x75\xf1\x00\x65\x6d\x5c\xd8\x21\xc7\x13\x99\x20\x24\xbe\x48\x78\x97\xb9\x0b\xb6\x98\xe4\x4a\x3d\x46\x49\x52\xae\xaa\xa5\x6c\x70\x16\x32\xc4\xbb\xda\x58\xd4\x53\xfa\x22\x61\xa7\xea\x3e\x7c\xed\xf5\xf3\x49\x84\xe4\x98\x57\x41\x53\xcd\x44\x57\x0e\x63\xff\x03\x6d\x7d\x68\x42\x0a\x7d\xf1\xe6\xdf\x22\xd6\x79\xf6\xe5\x00\xc1\x51\x83\x77\xdc\x52\x05\x76\x1e\x37\xb4\xd5\xab\x50\x82\xaa\x7c\x41\x09\xbf\xde\xbd\x77\x34\x0e\x7c\xdc\x1c\x66\x8b\x3d\xc8\x25\x30\xb9\x6b\x12\x87\xb7\x02\xb2\xe7\x20\xd4\xf7\xcb\xa2\xb4\x3d\x4b\x96\xfb\x1c\xdd\x66\xb0\x39\xb3\x2d\xcf\xf8\x54\x29\x83\x1c\x92\xdd\x09\x2b\xec\xc2\x8c\x90\xf6\xc7\xff\x19\x65\x97\x4a\x93\x35\xea\xe8\x9e\x3f\x6a\xd4\xd1\x00\x73\xc4\xf0\xe5\x83\xdb\xeb\xd0\x6f\xa1\x6f\xe2\xac\x5b\x0a\xb8\x4c\x9d\x11\xab\x7a\x18\xfc\xcb\xcb\x9f\x2e\x2f\x23\xda\xfa\xdb\xb4\xe2\xca\xb7\x33\x2d\x7e\x15\xbc\xd7\x04\x06\xfd\x59\xe2\xa5\x36\x38\x75\x01\x02\x9f\x58\x59\x15\xe8\xcb\x87\xe9\xa0\x98\xae\xb8\x20\xff\x6f\x03\x42\xf0\x65\x11\x14\xce\xaa\xaa\x10\xc8\x81\x19\x2a\xc0\x33\xf1\x04\xce\xf5\x0f\xea\xa6\xfe\xa7\x51\x7a\x10\x68\xb1\x20\xf2\x54\xed\x1c\x5e\x21\x15\xc5\x91\x75\x0b\xaf\xa7\xff\xdd\x4e\xaa\x83\x8f\xc7\x20\x9c\xb9\xb0\x10\x5d\x20\x03\x8f\x2e\x78\xfe\x07\xc3\xfd\x41\xf1\xd3\x7c\x6a\x5d\x9c\x11\xe9\x5d\x2c\x5e\x8b\x4d\x68\x0f\x0a\xe5\x33\x5d\x13\xb7\x58\x55\x4d\x09\x67\x63\x99\xe4\x4c\x73\x0a\x1f\x51\x50\x08\x6b\x78\xf1\xf0\xcf\x16\xeb\x7f\xe5\xca\xd8\x25\xc9\xb4\x70\x71\xe8\xe5\x1c\x6e\x9e\x58\x6a\x8b\x1d\x28\xe9\xe2\xa2\xbf\x5b\xf5\xb2\x43\x94\x72\x3c\x51\x50\x44\x78\xa0\x2b\x1e\x9a\x40\x4f\x8a\x75\x99\x8a\xac\x8f\x35\x6e\x10\x25\xd9\xe4\x8f\xfd\xdc\xf1\x7f\x6d\xaa\xed\xd2\x55\x46\xbd\x5d\x9b\x71\xdd\xad\x74\x69\x9c\x51\xb1\xce\x1d\xa7\x54\xd5\x17\x1b\x6a\x5d\x04\x03\x7c\x0a\xad\xce\x9b\xdb\x95\x43\x52\x95\x04\xab\x30\xa1\x9a\x7f\x81\xf3\xf5\x7c\x0a\x0f\x8f\x75\x82\xb3\xf6\x79\x1c\x8a\xd4\x17\xeb\x81\x3e\x08\x39\x0b\xac\x3b\xe2\xd4\x71\xb9\xf0\xe8\xe0\x48\x10\x18\x14\x6c\x87\xbe\x09\x11\xaa\x70\x8a\x7d\x19\x8f\x90\x01\x4a\x6a\x2d\x58\x61\x94\x3b\x2d\xe1\xdd\x27\x60\x9c\x6b\x34\xc6\x61\x7e\xe5\x13\x47\x2f\xa4\xf9\xce\x4d\x64\xf1\xe8\xee\x3b\x17\x47\xd4\xd1\x6b\x62\x1e\x54\xa8\x4b\x61\x8c\x48\x5c\x35\x03\x8c\xac\x6a\x4e\x75\x90\xdb\x1b\xb4\x30\x98\xfd\x48\xbf\x15\x33\x2e\xad\x31\x9d\x08\xab\x59\x1b\x4e\x9b\x0a\xc5\xd9\x6d\x2f\xfa\x4c\x81\x35\x6a\x8e\x17\x15\x9c\x7a\x92\x4c\xa0\xf6\x92\x5a\x8b\x65\x65\x03\x41\x62\x88\xd1\xbf\x9a\xac\x35\x61\x46\xa4\xc0\x6a\x9b\x03\xa9\x0e\xbe\x4c\x68\x65\x49\x1c\x6d\x95\xe6\xff\xff\x25\x56\x63\xb8\xb2\x85\x74\xc7\x8a\x42\x6d\xc9\x86\xdf\x6a\xb6\x2e\xa9\xb1\x83\x17\x5f\x26\x3f\xcc\xe7\xf3\x2f\x93\x97\x0e\x4d\x9f\x1d\x2a\xa6\x59\x89\xd6\x59\xc8\x97\xc9\x4f\x7e\x3d\x4a\x98\x69\xec\x53\x9e\x02\xba\x9a\x2b\x5a\xea\x8c\x04\xae\xc1\x58\xd2\x71\x32\xda\xe8\x4c\x3e\x75\x1c\xfb\xf6\x17\x6d\x13\x45\x7a\xc2\x58\xd5\x74\x14\xbe\x03\x92\x32\x16\xbb\x3a\x2d\x7a\x9f\x13\xb2\x10\x12\xe1\xb7\xab\x0f\xef\x17\x3f\xaf\x3e\xde\x42\xc5\x76\x85\x62\x3c\x90\xb3\x9a\x49\x53\x50\xf7\x4a\xe9\x5b\x01\xc5\xdf\x0d\x2b\x62\x25\x8d\x3b\xdd\x8c\x32\x42\x1c\xe9\x71\x1e\xfc\xdd\xc0\xed\xc7\x7b\x30\x98\x6a\x12\x42\x83\xef\x18\x78\x48\xb9\x47\x44\xb7\xe4\x36\x92\x37\x91\xe8\xf6\xe6\xf3\xcd\x5d\x5f\xcc\x5c\x15\x9c\x72\xb6\x11\x56\x6c\x7c\x37\x4d\x99\x49\x28\x39\x87\x7b\x45\x48\x1d\x91\xec\x43\x46\x4e\x4d\xed\x35\xa3\xf0\xe1\x79\xea\x91\x98\xf6\xab\xdd\xab\xf7\xff\xb8\xfa\x6d\x05\xc6\x2a\x7d\xec\x40\x8e\x50\xef\xa4\xf7\xbd\x95\xa3\x78\x64\x2e\x23\xb9\xe5\x69\xd6\x0d\x3c\x67\x58\x26\xc8\x39\xf2\x59\x33\xcb\x58\x82\xd5\xf5\xe1\xe5\x7b\x47\x9a\xf9\xd9\xac\xf6\x03\xb4\x59\x16\x26\x68\x47\x07\xbd\xb4\xa3\x76\xb7\x0a\x80\xc4\x6a\x6e\xb7\x42\x66\xa6\x91\x5a\xb9\x10\xee\xbb\x01\xc0\xe5\x71\xed\x20\x9b\xa9\x5d\xaf\xd4\x74\xea\x73\x89\x42\xa3\x8b\x13\xac\x30\xc0\x8c\x51\xa9\x70\x36\xd7\xf6\xee\x1d\xe5\xc3\x28\x3b\xd6\x63\x0c\xf5\x17\xfb\x95\xd6\x6d\x4f\xb2\xd0\x90\xd9\xe8\x74\x66\x7f\x18\xcd\x55\x6a\x16\xa9\x92\x29\x56\xd6\x2c\xd4\x86\x12\x1b\x6e\x17\x5b\xa5\x1f\x85\x5c\xcf\x88\xf5\x99\x57\xb2\x71\x83\x6b\xb3\xf8\xc1\xfd\x17\x0d\x35\xf7\x1f\xdf\x7c\x5c\xc2\x15\xe7\xa0\x5c\x5b\x57\x1b\xcc\xea\xc2\x3b\x8d\x99\xf7\xc6\x92\x53\x37\x24\x9b\x42\x2d\xf8\x4f\xb1\x22\xea\x7b\xe2\x90\x57\xe7\x3d\xb9\x3a\x59\xf0\x78\x34\x7a\x2f\x8c\x8f\x3e\xcd\x76\x67\xf0\xc1\x97\x82\xaf\x24\xd8\xd6\x94\x21\xde\xf4\xf4\x7b\xc4\x74\x4c\xdf\x2b\x5f\x26\x04\x9d\x43\x82\x19\xa9\xc3\xe6\xb8\x73\x51\x59\x48\x83\xba\x0d\x4a\xb1\x94\x16\x7c\xef\xe0\xb9\xb0\x78\x2c\xde\x51\xe5\xbd\x0f\x47\x88\xb9\x42\xae\x0b\x3c\x90\x3a\xf8\xbd\x69\xa4\x8d\xe9\xe3\x48\x7e\xd0\x68\x6b\x2d\x91\x77\xf3\xc5\x44\xab\x47\xd4\x83\x52\x46\xc8\x36\x72\x37\x4e\x7a\x1a\xc3\x39\xbc\xc6\x94\x51\xc2\xe5\x22\xf3\x46\x1e\xa1\xeb\x39\xa1\x3e\x40\x6d\x04\x6f\x26\xaa\x86\x3c\x84\xcc\x87\x14\xdf\x8c\x28\xa8\xa0\x40\x96\xe6\x41\x1e\x60\xa3\x84\xfb\x00\x18\xab\x6b\x37\xb6\x9c\xba\xd4\x6f\xa8\xfa\x0a\x45\xe8\xce\xdd\x17\xb3\xad\x08\xcd\x41\x6b\x5b\xb5\xf1\x89\x71\x56\x59\x10\xd6\x00\x4a\xab\xa9\x9b\xb2\x0a\xb6\x39\xb3\xb8\x89\xd6\x2b\xfd\x19\x4c\xaa\xa4\xa9\x4b\xa4\x4a\xa7\x22\x2f\x9e\xc3\xdb\x7e\xd9\x33\xa4\xd6\x18\xaa\xbb\xbe\x9a\xfd\x94\x39\x2d\x6a\xee\x6b\xe2\x47\xdc\xc1\xe4\xd7\xd5\xcd\xdd\xed\xd5\x87\x9b\xc9\x14\x92\x3a\x0c\x9a\x9b\xfb\x43\xd7\x13\x8b\x1c\xb4\x8f\x30\x74\xd1\xd9\xa7\xec\xa6\x77\xaf\x25\x77\x83\xec\x70\xc1\x9b\xd7\xff\xa6\x3b\x26\xbd\x92\x5b\x41\xce\x36\xd1\xee\xa7\xb3\x1e\xb8\xf6\x2f\x86\x3a\x9d\xf4\x10\xf6\x20\x64\x8a\xea\x23\xb2\x95\x03\xcf\x89\x50\x3e\x6a\x39\x28\x75\x1c\x18\xaa\x7b\x83\x76\x10\x93\x96\x30\x83\xaf\x13\x8d\x24\xe7\x2f\xb8\x9b\xc4\xa2\xfa\xd7\x09\x39\xd4\x64\xb9\x07\xe6\xc4\x2a\x7a\xd2\x48\xff\xed\x1b\x7c\x94\x5d\xa3\xd4\x89\xd2\xde\x74\x19\x49\x5d\x00\x25\x25\xe3\xf0\x86\x60\xaf\x63\x3a\x8e\xc1\xe3\x43\x2f\xc6\xf9\x2f\x38\x30\xe9\xd8\x1f\xa6\xbb\x8d\xbd\xd7\x14\xc0\xf6\x74\xc0\x2c\xd1\xf2\xa5\x7a\xfb\x52\x73\xa0\xab\x26\x03\x88\x04\x22\x2f\xf9\x40\x87\x31\x3e\xb9\x03\xf8\xdd\x28\xf9\x89\xd9\xfc\xe6\x89\x18\x3c\x7c\x63\x36\x22\xd8\x25\x15\x86\xcd\x1b\x4f\xb2\xd6\x70\x7c\x1a\x12\xb0\xa9\x0b\x97\x60\xdc\xd8\x63\x90\x24\xb4\x23\x05\x46\x35\xd2\x81\xe7\x77\xb6\xdf\x19\xda\x23\xee\x9c\x07\x8f\x90\x6c\x7d\x9b\xda\x76\x79\x60\xdc\x3d\xe4\x96\xf0\x15\x26\x99\x22\xcb\xfa\x0a\x93\x84\xe9\xa8\x3d\x36\x1f\xda\x49\x7b\xe0\x1b\x7c\x73\xc5\x2c\xd1\x3d\x86\x0f\x26\x5f\xe7\x99\x52\xf3\x84\xe9\x6f\x93\xe9\x60\x9f\xe6\x3f\xfe\xd5\x57\x4b\xb9\x9d\x5e\x51\x75\xca\xdb\xee\xfb\x3c\x5d\xfb\xcf\x10\x64\x43\x93\x9a\x33\xc6\x5d\xce\xf0\xce\xb4\x8a\xfb\x83\x12\x33\x98\x2c\xe3\xf1\xf7\x23\x67\xde\xef\x17\x3f\x13\x58\xcf\xe0\x23\x34\xbc\x2f\xa4\x92\xb3\x44\x48\xa6\x77\x2f\x03\xe0\x9e\xa3\x7d\x63\xfb\x0e\x4c\xff\x8c\x48\x9b\x67\x0a\xe3\x05\x08\xfc\xbf\xa8\x94\x1b\x0e\xec\x80\x64\xf3\xf7\xbc\x3c\x85\x33\x9c\xeb\x5c\xef\x32\x48\x94\xcd\xc3\x5d\x4c\x8e\x91\xec\x69\xc6\xd5\x3a\x87\x83\x43\x4f\x43\x18\x10\x6b\xe9\x6c\xda\x75\x81\xdd\xa1\x11\xd2\xee\x7d\x10\x9d\x19\xc6\xf9\xc4\xbb\xb1\x20\xf5\x29\x65\x8c\x4d\x23\x01\x66\x11\x1f\x1f\xd8\xf8\x88\xc7\x9d\xaf\x5f\x39\x25\xf1\xcc\x03\x15\x7f\x87\x38\x36\xc3\x6c\x32\x91\x79\xab\x55\x79\x76\x3a\x72\xbb\x47\x73\x52\x89\x7a\x8d\xa6\xfd\xc5\x48\x84\x2b\xf7\x3e\xdd\xd7\x62\xfe\xe7\x0f\xf8\x24\x8c\xed\xca\x87\xae\xb6\xfd\xcb\x72\x95\x2f\x26\xee\x30\x7b\x86\xdf\x1c\xbd\x9a\x6b\x8a\xca\xfd\x3e\xc7\xc9\x3b\x66\xe8\x23\xd2\x0c\xdb\xe7\x69\x91\xe0\xc4\xeb\xb3\x88\x54\xd1\x3e\x77\xf4\xf8\x19\xe1\x08\xfa\x9d\xfd\x33\x99\xf1\xd3\x80\xbf\x9e\xa3\x13\x86\x7f\x72\x83\xc6\x52\x6d\xf0\xbc\x32\xed\xae\xd9\x3b\xea\x15\x9e\x22\x2d\x8c\x35\xc2\xfe\x13\xec\x8c\x7c\x24\x1e\x15\x4e\x1b\xc7\xf3\xd2\x6e\x48\xb5\x9e\xc7\xae\x29\x3d\x91\xdd\xfe\x74\x74\x1c\x0a\x7a\x27\x95\x13\x7a\x80\xb3\x94\x13\xf6\x9e\x50\x8e\x33\xe0\x67\x2b\xe7\xd2\x0c\xca\x70\x5a\x45\xd9\x60\xd8\x3d\x92\x62\xa0\x34\xf2\x6c\xff\x99\x52\xc2\xaa\xe7\x70\x80\x5b\xcf\x85\xff\x69\x02\x0e\xca\x7e\xd6\xe5\xa7\x8c\x83\xe0\x19\x58\xb2\xea\xf9\x66\x33\xb2\xe8\x97\x98\xd6\x6c\x5f\x1c\xf7\x7c\x6c\x10\x76\xbf\xab\xba\x39\x74\xc6\x52\x51\x08\xcb\x2c\x92\xee\xd7\x9a\x95\x25\xb3\x22\x85\x9c\x49\x5e\x50\x6e\xa3\x54\x57\x55\x45\xf8\xb9\xcb\x61\x90\x1b\xc4\x6b\x73\xfc\x63\xc4\x23\x46\x9a\x1f\x23\xfe\xbd\xbc\xc4\x34\x36\xdb\xfb\xb5\xd6\xc5\x28\xde\x07\x8f\x1a\xc1\x60\xf3\x8a\x15\x55\xce\x5e\x75\xcf\xc2\x8f\x71\xfd\x4f\x5d\x7b\xcb\xfe\x67\x36\xc8\x7b\xd3\x6c\xaa\xf3\xd8\xba\x19\x8c\xff\x37\x00\x00\xff\xff\x59\x5d\xdf\xe1\xab\x2c\x00\x00")

func appcatalogAppscodeCom_appbindingsYamlBytes() ([]byte, error) {
	return bindataRead(
		_appcatalogAppscodeCom_appbindingsYaml,
		"appcatalog.appscode.com_appbindings.yaml",
	)
}

func appcatalogAppscodeCom_appbindingsYaml() (*asset, error) {
	bytes, err := appcatalogAppscodeCom_appbindingsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "appcatalog.appscode.com_appbindings.yaml", size: 11435, mode: os.FileMode(420), modTime: time.Unix(1573722179, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"appcatalog.appscode.com_appbindings.v1.yaml": appcatalogAppscodeCom_appbindingsV1Yaml,
	"appcatalog.appscode.com_appbindings.yaml":    appcatalogAppscodeCom_appbindingsYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"appcatalog.appscode.com_appbindings.v1.yaml": {appcatalogAppscodeCom_appbindingsV1Yaml, map[string]*bintree{}},
	"appcatalog.appscode.com_appbindings.yaml":    {appcatalogAppscodeCom_appbindingsYaml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
