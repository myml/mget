// mget project mget.go
package mget

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const Thread = 5
const CacheSize = 1024 * 1024

func Get(url string) (io.ReadCloser, int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return resp.Body, resp.ContentLength, nil
	}
	resp.Body.Close()
	r, w := io.Pipe()
	go mget(url, w, resp.ContentLength)
	return r, resp.ContentLength, nil
}
func mget(url string, w *io.PipeWriter, length int64) {
	var start int64
	for {
		var readers []io.Reader
		for i := 0; i < Thread; i++ {
			if start+CacheSize >= length {
				readers = append(readers, rget(url, start, 0, i))
				start = length
				break
			} else {
				readers = append(readers, rget(url, start, start+CacheSize, i))
				start += CacheSize + 1
			}
		}
		r := io.MultiReader(readers...)
		_, err := io.Copy(w, r)
		if err != nil {
			w.CloseWithError(err)
			return
		}
		if start >= length {
			w.Close()
			return
		}
	}
}
func rget(url string, start, end int64, job int) io.Reader {
	r, w := io.Pipe()
	go func() {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			w.CloseWithError(errors.WithStack(err))
			return
		}
		if end == 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
		} else {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.CloseWithError(errors.WithStack(err))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 206 {
			w.CloseWithError(errors.New("not support http range"))
			return
		}
		if job == 0 {
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				w.CloseWithError(err)
			} else {
				w.Close()
			}
			return
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.CloseWithError(errors.WithStack(err))
			return
		}
		_, err = w.Write(b)
		if err != nil {
			w.CloseWithError(err)
			return
		}
		w.Close()
	}()
	return r
}
