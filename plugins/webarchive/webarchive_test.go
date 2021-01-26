package webarchive

import (
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestRead(t *testing.T) {
	file, err := os.Open("./resources/test.webarchive")
	if err != nil {
		t.Error(err)
		return
	}

	archive, err := Read(file)
	if err != nil {
		t.Error(err)
		return
	}

	if diff := deep.Equal(archive.MainResource.MIMEType, "text/html"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(archive.MainResource.URL, "https://archive.org/web/researcher/ArcFileFormat.php"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(archive.MainResource.TextEncodingName, "UTF-8"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(string(archive.MainResource.Data), "Foo, Bar"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(len(archive.SubResources), 1); diff != nil {
		t.Error(diff)
		return
	}

	if diff := deep.Equal(archive.SubResources[0].URL, "https://archive.org/includes/build/npm/jquery-ui.min.js?v1.12.1"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(string(archive.SubResources[0].Data), "Bar, Foo"); diff != nil {
		t.Error(diff)
	}
}
