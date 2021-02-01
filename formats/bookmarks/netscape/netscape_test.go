package netscape

import (
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestUnmarshal(t *testing.T) {
	file, err := os.Open("./resources/bookmarks.html")
	if err != nil {
		t.Error(err)
		return
	}

	bookmarks, err := Unmarshal(file)
	if err != nil {
		t.Error(err)
		return
	}

	if diff := deep.Equal(len(bookmarks), 8); diff != nil {
		t.Error(diff)
		return
	}
}
