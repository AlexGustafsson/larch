package api

import "time"

type Page[T any] struct {
	Page     int       `json:"page"`
	Size     int       `json:"size"`
	Count    int       `json:"count"`
	Total    int       `json:"total"`
	Embedded T         `json:"_embedded"`
	Links    PageLinks `json:"_links"`
}

type PageLinks struct {
	Curies   []Link `json:"curies"`
	Self     Link   `json:"self"`
	First    Link   `json:"first"`
	Previous Link   `json:"prev,omitzero"`
	Next     Link   `json:"next,omitzero"`
	Last     Link   `json:"last"`
	Page     Link   `json:"page"`
}

type Link struct {
	Href      string `json:"href"`
	Name      string `json:"name,omitempty"`
	Templated bool   `json:"templated,omitempty"`
}

type Snapshot struct {
	ID       string           `json:"id"`
	URL      string           `json:"url"`
	Origin   string           `json:"origin"`
	Date     time.Time        `json:"date"`
	Embedded SnapshotEmbedded `json:"_embedded"`
	Links    SnapshotLinks    `json:"_links"`
}

type SnapshotEmbedded struct {
	Artifacts []Artifact `json:"larch:artifact"`
}

type SnapshotLinks struct {
	Curies    []Link `json:"curies"`
	Self      Link   `json:"self"`
	Origin    Link   `json:"larch:origin"`
	Artifacts Link   `json:"larch:artifacts"`
}

type Artifact struct {
	ContentType     string            `json:"contentType"`
	ContentEncoding string            `json:"contentEncoding,omitempty"`
	Digest          string            `json:"digest"`
	Size            int64             `json:"size"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Links           ArtifactLinks     `json:"_links"`
}

type ArtifactLinks struct {
	Curies   []Link `json:"curies"`
	Self     Link   `json:"self"`
	Snapshot Link   `json:"larch:snapshot"`
	Origin   Link   `json:"larch:origin"`
	Blob     Link   `json:"larch:blob"`
}

type SnapshotPageEmbedded struct {
	Snapshots []Snapshot `json:"larch:snapshot"`
}

type ArtifactPageEmbedded struct {
	Artifacts []Artifact `json:"larch:artifact"`
}

type ArtifactEmbedded struct {
	Artifacts []Artifact `json:"larch:artifact"`
}
