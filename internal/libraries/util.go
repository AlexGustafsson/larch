package libraries

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

var _ DigestWriteCloser = (*digestWriter)(nil)

type digestWriter struct {
	w      io.WriteCloser
	hash   hash.Hash
	writer io.Writer
}

func newDigestWriter(w io.WriteCloser) *digestWriter {
	hash := sha256.New()
	return &digestWriter{
		w:      w,
		hash:   hash,
		writer: io.MultiWriter(w, hash),
	}
}

func (w *digestWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *digestWriter) Close() error {
	return w.w.Close()
}

func (w *digestWriter) Digest() string {
	return "sha256:" + hex.EncodeToString(w.hash.Sum(nil))
}

var _ DigestReadCloser = (*digestReader)(nil)

type digestReader struct {
	r      io.ReadCloser
	hash   hash.Hash
	reader io.Reader
}

func newDigestReader(r io.ReadCloser) *digestReader {
	hash := sha256.New()
	return &digestReader{
		r:      r,
		hash:   hash,
		reader: io.TeeReader(r, hash),
	}
}

func (r *digestReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *digestReader) Close() error {
	return r.r.Close()
}

func (r *digestReader) Digest() string {
	return "sha256:" + hex.EncodeToString(r.hash.Sum(nil))
}
