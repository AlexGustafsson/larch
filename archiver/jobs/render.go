package jobs

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/archiver/pipeline"
	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// RenderSite creates a render of the site.
func CreateRenderJob(url *url.URL, quality int) *pipeline.Job {
	perform := func(job *pipeline.Job) ([]*warc.Record, error) {
		// create context
		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		// capture entire browser viewport, returning png with quality=90

		var buffer []byte
		err := chromedp.Run(ctx, createScreenshotTasks(url, 90, &buffer))
		if err != nil {
			return nil, fmt.Errorf("Unable to render site: %v", err)
		}

		record, err := newRenderRecord(url, buffer)
		return []*warc.Record{record}, err
	}

	return pipeline.NewJob("Render", "Renders the site", perform)
}

func createScreenshotAction(quality int64, buffer *[]byte) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
		if err != nil {
			return err
		}

		width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

		// Force viewport emulation
		orientation := &emulation.ScreenOrientation{
			Type:  emulation.OrientationTypePortraitPrimary,
			Angle: 0,
		}
		err = emulation.SetDeviceMetricsOverride(width, height, 1, false).WithScreenOrientation(orientation).Do(ctx)
		if err != nil {
			return err
		}

		clip := &page.Viewport{
			X:      contentSize.X,
			Y:      contentSize.Y,
			Width:  contentSize.Width,
			Height: contentSize.Height,
			Scale:  1,
		}
		*buffer, err = page.CaptureScreenshot().WithQuality(quality).WithClip(clip).Do(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

func createScreenshotTasks(url *url.URL, quality int64, buffer *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url.String()),
		chromedp.ActionFunc(createScreenshotAction(quality, buffer)),
	}
}

// newRenderRecord creates a record for a rendering of a site.
func newRenderRecord(url *url.URL, buffer []byte) (*warc.Record, error) {
	id, err := warc.CreateID()
	if err != nil {
		return nil, err
	}

	record := &warc.Record{
		Header: &warc.Header{
			Type:          warc.TypeConversion,
			TargetURI:     url.String(),
			RecordID:      id,
			Date:          time.Now(),
			ContentType:   "image/png",
			ContentLength: uint64(len(buffer)),
		},
		Payload: &warc.RawPayload{
			Data:   buffer,
			Length: uint64(len(buffer)),
		},
	}

	return record, nil
}
