package archivebox

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func sha256sum(root *os.Root, name string) (string, int64, error) {
	file, err := root.Open(name)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hash := sha256.New()
	size, err := io.Copy(io.Discard, io.TeeReader(file, hash))
	if err != nil {
		return "", 0, err
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), size, nil
}
