package archiver

import (
	"net"
	"net/url"
	"sync"

	"github.com/AlexGustafsson/larch/archiver/jobs"
	"github.com/AlexGustafsson/larch/archiver/pipeline"
	"github.com/AlexGustafsson/larch/formats/warc"
	log "github.com/sirupsen/logrus"
)

// Archiver contains options for an archiver.
type Archiver struct {
	MaxDepth        uint32
	Render          bool
	RenderQuality   int
	file            *warc.File
	syncGroup       sync.WaitGroup
	ResolverAddress net.IP
	ResolverPort    uint16
	UserAgent       string
}

// NewArchiver creates a new archiver following best practices.
func NewArchiver() *Archiver {
	archiver := &Archiver{
		MaxDepth:        1,
		Render:          false,
		RenderQuality:   100,
		file:            &warc.File{},
		ResolverAddress: net.ParseIP("192.168.1.1"),
		ResolverPort:    uint16(53),
		UserAgent:       "Larch (github.com/AlexGustafsson/larc)",
	}

	return archiver
}

// file, err := archiver.Archive("https://google.se", "https://google.com")
// archiver.schedule(1, 2)
// for workers: go func() {
// 	job := archiver.consume()
// 	record, err := job.perform()
// 	if err
// 		if job.retry
// 			job.retries++
// 		archiver.schedule(job)
// 	job.complete()
// }

// Schedule schedules a job for processing.
func (archiver *Archiver) Schedule(jobs ...*pipeline.Job) {
	// TODO: lock scheduling array? Do it outside the loop for some optimization?
	for _, job := range jobs {
		log.Debugf("Scheduling job '%s' (%s)", job.Name, job.Description)
		job.JobCompletedCallback = archiver.OnJobCompleted
		job.JobFailedCallback = archiver.OnJobFailed
		job.PerformJobCallback = archiver.OnPerformJob

		// TODO: Don't schedule them all - create a worker pool to use instead
		archiver.syncGroup.Add(1)
		go func(job *pipeline.Job) {
			job.Perform()
		}(job)
	}
}

// OnPerformJob handles jobs before being performed.
func (archiver *Archiver) OnPerformJob(job *pipeline.Job) {
	log.Debugf("Performing job '%s'", job.Name)
}

// OnJobCompleted handles a completed job.
func (archiver *Archiver) OnJobCompleted(job *pipeline.Job, records ...*warc.Record) {
	log.Infof("Completed job '%s'", job.Name)

	for _, record := range records {
		// TODO: lock?
		archiver.file.Records = append(archiver.file.Records, record)

		switch record.Header.ContentType {
		case "application/http;msgtype=response":
			if record.Payload != nil {
				scrapeJob := jobs.CreateScrapeJob(record.Payload.(*jobs.HTTPResponsePayload))
				archiver.Schedule(scrapeJob)
			}
		}

		switch record.Header.TargetURI {
		case "<metadata://github.com/AlexGustafsson/larch/scrape.txt>":
			// httpJobs := jobs.CreateHTTPJobs(record.Payload)
			// archiver.Schedule(httpJobs...)
		}
	}

	archiver.syncGroup.Done()
}

// OnJobFailed handles a failed job, potentially rescheduling it.
func (archiver *Archiver) OnJobFailed(job *pipeline.Job, err error) {
	log.Errorf("Job '%s' failed: %v", job.Name, err)
	if job.Tries < job.MaximumTries {
		archiver.Schedule(job)
	}

	archiver.syncGroup.Done()
}

// Archive archives a URL as a WARC archive.
func (archiver *Archiver) Archive(urls ...*url.URL) (*warc.File, error) {
	for _, url := range urls {
		httpJob := jobs.CreateHTTPJob(url, archiver.UserAgent)
		archiver.Schedule(httpJob)

		// TODO: Schedule before each new domain
		// to make it actually usable
		robotsJob, err := jobs.CreateRobotsJob(url, archiver.UserAgent)
		if err != nil {
			return nil, err
		}
		archiver.Schedule(robotsJob)

		// TODO: Schedule before each new domain
		// TODO: Handle concurrent to
		dnsJob, err := jobs.CreateDNSJob(url, archiver.ResolverAddress, archiver.ResolverPort)
		if err != nil {
			return nil, err
		}
		archiver.Schedule(dnsJob)

		if archiver.Render {
			// TODO: Handle concurrent to...
			// Render the initial page
			renderJob := jobs.CreateRenderJob(url, archiver.RenderQuality)
			archiver.Schedule(renderJob)
		}
	}

	archiver.syncGroup.Wait()
	return archiver.file, nil
}
