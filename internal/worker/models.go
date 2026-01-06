package worker

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type JobRequest struct {
	Token    string
	Archiver Archiver
	Job      Job
}

type Job struct {
	ID         string
	Library    string
	Deadline   time.Time
	URL        string
	Origin     string
	SnapshotID string
	Status     string
	Requested  time.Time
	Accepted   time.Time
	Started    time.Time
	Ended      time.Time
	Error      string
}

type Archiver struct {
	ChromeArchiver     *ChromeArchiver
	ArchiveOrgArchiver *ArchiveOrgArchiver
	OpenGraphArchiver  *OpenGraphArchiver
}

type ChromeArchiver struct {
	SavePDF               bool
	SaveSinglefile        bool
	ScreenshotResolutions []Resolution
}

type Resolution string

func (r Resolution) Rect() (int64, int64, error) {
	width, height, ok := strings.Cut(string(r), "x")
	if !ok {
		return 0, 0, fmt.Errorf("invalid resolution")
	}

	w, werr := strconv.ParseInt(width, 10, 64)
	h, herr := strconv.ParseInt(height, 10, 64)
	return w, h, errors.Join(werr, herr)
}

type ArchiveOrgArchiver struct{}

type OpenGraphArchiver struct{}

type Strategy struct {
	Library   string
	Archivers []Archiver
}
