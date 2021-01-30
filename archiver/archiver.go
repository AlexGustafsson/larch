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
	fileMutex       sync.Mutex
	pool            *pipeline.Pool
	ResolverAddress net.IP
	ResolverPort    uint16
	UserAgent       string
}

// NewArchiver creates a new archiver following best practices.
func NewArchiver(parallelism uint) *Archiver {
	archiver := &Archiver{
		MaxDepth:        1,
		Render:          false,
		RenderQuality:   100,
		file:            &warc.File{},
		pool:            pipeline.NewPool(parallelism),
		ResolverAddress: net.ParseIP("192.168.1.1"),
		ResolverPort:    uint16(53),
		UserAgent:       "Larch (github.com/AlexGustafsson/larc)",
	}

	return archiver
}

// Schedule schedules a job for processing.
func (archiver *Archiver) Schedule(jobs ...*pipeline.Job) {
	for _, job := range jobs {
		log.Debugf("Scheduling job '%s' (%s)", job.Name, job.Description)
		job.JobCompletedCallback = archiver.OnJobCompleted
		job.JobFailedCallback = archiver.OnJobFailed
		job.PerformJobCallback = archiver.OnPerformJob

		archiver.pool.Submit(job)
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
		archiver.fileMutex.Lock()
		archiver.file.Records = append(archiver.file.Records, record)
		archiver.fileMutex.Unlock()

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
}

// OnJobFailed handles a failed job, potentially rescheduling it.
func (archiver *Archiver) OnJobFailed(job *pipeline.Job, err error) {
	log.Errorf("Job '%s' failed: %v", job.Name, err)
	if job.Tries < job.MaximumTries {
		archiver.Schedule(job)
	}
}

// Archive archives a URL as a WARC archive.
func (archiver *Archiver) Archive(urls ...*url.URL) (*warc.File, error) {
	archiver.pool.Start()

	for _, url := range urls {
		httpJob := jobs.CreateHTTPJob(url, archiver.UserAgent)
		archiver.Schedule(httpJob)

		robotsJob, err := jobs.CreateRobotsJob(url, archiver.UserAgent)
		if err != nil {
			return nil, err
		}
		archiver.Schedule(robotsJob)

		dnsJob, err := jobs.CreateDNSJob(url, archiver.ResolverAddress, archiver.ResolverPort)
		if err != nil {
			return nil, err
		}
		archiver.Schedule(dnsJob)

		if archiver.Render {
			// Render the initial page
			renderJob := jobs.CreateRenderJob(url, archiver.RenderQuality)
			archiver.Schedule(renderJob)
		}
	}

	archiver.pool.Wait()
	return archiver.file, nil
}
