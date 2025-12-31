package chrome

import (
	"context"
	"fmt"
	"os"

	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/chromedp/chromedp"
)

var _ (archivers.Archiver) = (*Archiver)(nil)

type Archiver struct {
	ScreenshotResolutions []Resolution
	SavePDF               bool
	SaveSinglefile        bool
}

type Resolution struct {
	Width  int64
	Height int64
}

func (a *Archiver) Archive(ctx context.Context, snapshotWriter libraries.SnapshotWriter, url string) error {
	// TODO: Might need to have --disable-web-security for singlepage
	ctx, cancel := chromedp.NewContext(ctx, chromedp.WithErrorf(func(s string, a ...any) { fmt.Fprintf(os.Stderr, s, a...) }), chromedp.WithBrowserOption(chromedp.WithBrowserLogf(func(s string, a ...any) { fmt.Fprintf(os.Stderr, s, a...) }), chromedp.WithBrowserErrorf(func(s string, a ...any) { fmt.Fprintf(os.Stderr, s, a...) })))
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1280, 720),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if a.SaveSinglefile {
				err := saveToSinglefileAction{
					SnapshotWriter: snapshotWriter,
				}.Do(ctx)
				if err != nil {
					return err
				}
			}

			if a.SavePDF {
				err := printAction{
					SnapshotWriter: snapshotWriter,
				}.Do(ctx)
				if err != nil {
					return err
				}
			}

			for _, resolution := range a.ScreenshotResolutions {
				err := screenshotAction{
					SnapshotWriter: snapshotWriter,
					Width:          resolution.Width,
					Height:         resolution.Height,
				}.Do(ctx)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}
