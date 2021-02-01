package netscape

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Bookmark is a bookmark entry.
type Bookmark struct {
	URL          string
	AddedDate    string
	LastModified string
	Description  string
	Title        string
	Tags         []string
	// Categories is a flat tree of the categories the bookmark is part of.
	// The first category is the top-most, the second a subcategory and so on.
	Categories []*Category
}

// Category is a category of a bookmark.
type Category struct {
	Title        string
	AddedDate    string
	LastModified string
}

// Unmarshal unmarshals the NETSCAPE-Bookmark-file-1 format
func Unmarshal(reader io.Reader) ([]*Bookmark, error) {
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	bookmarks := make([]*Bookmark, 0)

	document.Find("a").Each(func(i int, a *goquery.Selection) {
		bookmark := &Bookmark{
			URL:          a.AttrOr("href", ""),
			AddedDate:    a.AttrOr("add_date", ""),
			LastModified: a.AttrOr("last_modified", ""),
			Description:  a.AttrOr("description", ""),
			Title:        a.Text(),
			Tags:         strings.Split(a.AttrOr("tags", ""), ","),
			Categories:   findCategories(a),
		}
		bookmarks = append(bookmarks, bookmark)
	})

	return bookmarks, nil
}

func findCategories(a *goquery.Selection) []*Category {
	list := a.Closest("DL").Prev()
	categories := make([]*Category, 1)
	categories[0] = &Category{
		Title:        list.Text(),
		AddedDate:    list.AttrOr("add_date", ""),
		LastModified: list.AttrOr("last_modified", ""),
	}

	if list.Length() > 0 && len(list.Text()) > 0 {
		categories = append(findCategories(list), categories...)
	}

	return categories
}
