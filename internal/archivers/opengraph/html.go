package opengraph

import "encoding/xml"

type Document struct {
	XMLName xml.Name `xml:"html"`
	Meta    []Meta   `xml:"head>meta"`
}

type Meta struct {
	Property string `xml:"property,attr"`
	Content  string `xml:"content,attr"`
}
