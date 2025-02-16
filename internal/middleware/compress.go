package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var CompresTypes = []string{
	"application/json",
	"text/html",
}

type CompressWrite struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func (c *CompressWrite) Write(b []byte) (int, error) {
	res, err := c.zw.Write(b)
	return res, err
}

type CompressRead struct {
	io.ReadCloser
	zr *gzip.Reader
}

func (c *CompressRead) Read(p []byte) (int, error) {
	res, err := c.zr.Read(p)
	return res, err
}

func Compress(h http.HandlerFunc) http.HandlerFunc {
	foo := func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), `gzip`) {
			h.ServeHTTP(w, r)
			return
		}
		gz := gzip.NewWriter(w)
		defer gz.Close()
		cw := &CompressWrite{w, gz}
		cw.Header().Set("Content-Encoding", "gzip")

		if !strings.Contains(r.Header.Get("Content-Encoding"), `gzip`) {
			h.ServeHTTP(cw, r)
			return
		}

		gzR, err := gzip.NewReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cr := &CompressRead{r.Body, gzR}

		r.Body = cr

		h.ServeHTTP(cw, r)
	}
	return foo
}
