package archiver

import (
	"net/url"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// FetchRobotsTXT fetches the Robots TXT file.
func (archiver *Archiver) FetchRobotsTXT(target *url.URL) (*warc.Record, *warc.Record, error) {
	robotsTarget, err := url.Parse(target.Scheme + "://" + target.Host + "/robots.txt")
	if err != nil {
		return nil, nil, err
	}

	return archiver.Fetch(robotsTarget)
}
