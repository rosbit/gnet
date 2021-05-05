// remove go1.16 dependency build go1.16

package gnet

import (
	// "io/fs"
	"os"
	"time"
	"path"
	"net/url"
	"net/http"
)

// result of HTTP response, returned by FileInfo.Sys()
type Result struct {
	Status int
	Resp *http.Response
	Err error
}

func HttpRequest(url string, options ...Option) *File /*fs.File*/ {
	return gnet_fs(url, "", options...)
}

func Get(url string, options ...Option) *File /*fs.File*/ {
	return gnet_fs(url, http.MethodGet, options...)
}

func Post(url string, options ...Option) *File /*fs.File*/ {
	return gnet_fs(url, http.MethodPost, options...)
}

func Put(url string, options ...Option) *File /*fs.File*/ {
	return gnet_fs(url, http.MethodPut, options...)
}

func Delete(url string, options ...Option) *File /*fs.File*/ {
	return gnet_fs(url, http.MethodPut, options...)
}

func Head(url string, options ...Option) *File /*fs.File*/ {
	return gnet_fs(url, http.MethodHead, options...)
}

func gnet_fs(url string, method string, options ...Option) *File /*fs.File*/ {
	option := getOptions(options...)
	option.dontReadRespBody = true
	return gnet_fs_i(url, method, option)
}

func gnet_fs_i(url, method string, option *Options) *File /*fs.File*/ {
	option.method = method
	return &File{
		url: url,
		option: option,
	}
}

/*
// ---- implementation of fs.FS ----
type gufs_t struct {
}

var (
	gufs = &gufs_t{}
)

func (gufs *gufs_t) Open(name string) (fs.File, error) {
	return Get(name)
}*/

// ---- implementation of fs.File ----
type File struct {
	url string
	option *Options
	Result
}

func (f *File) Stat() (*FileInfo /*fs.FileInfo*/, error) {
	f.run()
	if f.Err != nil {
		return nil, f.Err
	}
	return &FileInfo{f: f}, nil
}

func (f *File) Read(p []byte) (int, error) {
	f.run()
	if f.Err != nil {
		return 0, f.Err
	}
	if f.Resp.Body == nil {
		return 0, os.ErrNotExist /*fs.ErrNotExist*/
	}
	return f.Resp.Body.Read(p)
}

func (f *File) Close() error {
	f.run()
	if f.Err != nil {
		return f.Err
	}
	if f.Resp.Body == nil {
		return os.ErrNotExist /*fs.ErrNotExist*/
	}
	return f.Resp.Body.Close()
}

func (f *File) run() {
	if f.Status > 0 {
		return
	}

	var call httpFunc_i
	if f.option.jsonCall {
		call = json_i
	} else {
		call = http_i
	}
	f.Status, _, f.Resp, f.Err = call(f.url, f.option)
}

// ---- implementation of fs.FileInfo ----
type FileInfo struct {
	f *File
	u *url.URL
	e error
}

// base name of the file
func (fi *FileInfo) Name() string {
	fi.parse()
	if fi.e != nil {
		return ""
	}
	if len(fi.u.Path) == 0 {
		return ""
	}
	return path.Base(fi.u.Path)
}

// length in bytes for regular files; system-dependent for others
func (fi *FileInfo) Size() int64 {
	return fi.f.Resp.ContentLength
}

// file mode bits
/*
func (fi *FileInfo) Mode() fs.FileMode {
	return fs.ModeSocket
}*/

// modification time
func (fi *FileInfo) ModTime() time.Time {
	t, _ := GetLastModified(fi.f.Resp)
	return t
}

// abbreviation for Mode().IsDir()
func (fi *FileInfo) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (fi *FileInfo) Sys() interface{} {
	return &fi.f.Result
}

func (fi *FileInfo) parse() {
	if fi.e != nil || fi.u != nil {
		return
	}
	fi.u, fi.e = url.Parse(fi.f.url)
}
