package middlas

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"time"
)

// Pack2gzip - он и есть pack to gzip
func Pack2gzip(data2pack []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.ModTime = time.Now()
	_, err := zw.Write(data2pack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Write %w ", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Close %w ", err)
	}
	return buf.Bytes(), nil
}

// UnpackFromGzip - распаковщик
func UnpackFromGzip(data2unpack io.Reader) (io.Reader, error) {
	gzipReader, err := gzip.NewReader(data2unpack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader %w ", err)
	}
	if err := gzipReader.Close(); err != nil {
		return nil, fmt.Errorf("zr.Close %w ", err)
	}
	return gzipReader, nil
}

// Ptr возвращает ссылку на int64 или float64 в зависимости от агрумента или параметра типа в случае int64 или float64 без дробной части
func Ptr[PP int64 | float64](w PP) *PP {
	i := w
	return &i
}
