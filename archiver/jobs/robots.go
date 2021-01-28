package jobs

import (
	"net/url"

	"github.com/AlexGustafsson/larch/archiver/pipeline"
)

// CreateRobotsJob fetches the Robots TXT file.
func CreateRobotsJob(target *url.URL, userAgent string) (*pipeline.Job, error) {
	robotsTarget, err := url.Parse(target.Scheme + "://" + target.Host + "/robots.txt")
	if err != nil {
		return nil, err
	}

	return CreateHTTPJob(robotsTarget, userAgent), nil
}
