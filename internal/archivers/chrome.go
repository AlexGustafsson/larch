package archivers

import (
	"context"
	"fmt"
	"os"

	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/singlefile"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type ChromeArchiver struct {
}

func (a *ChromeArchiver) Archive(ctx context.Context, snapshotWriter libraries.SnapshotWriter, url string) error {
	// TODO: Might need to have --disable-web-security for singlepage
	ctx, cancel := chromedp.NewContext(ctx, chromedp.WithErrorf(func(s string, a ...any) { fmt.Fprintf(os.Stderr, s, a...) }), chromedp.WithBrowserOption(chromedp.WithBrowserLogf(func(s string, a ...any) { fmt.Fprintf(os.Stderr, s, a...) }), chromedp.WithBrowserErrorf(func(s string, a ...any) { fmt.Fprintf(os.Stderr, s, a...) })))
	defer cancel()

	var screenshot []byte
	var pdf []byte
	var html string
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1280, 720),
		chromedp.Navigate(url),
		chromedp.Evaluate(singlefile.HookScript, nil),
		chromedp.Evaluate(singlefile.Script, nil),
		chromedp.Evaluate(`singlefile.getPageData({removeHiddenElements: true,
  removeUnusedStyles: true,
  removeUnusedFonts: true,
  removeImports: true,
  blockScripts: true,
  blockAudios: true,
  blockVideos: true,
  compressHTML: false,
  removeAlternativeFonts: true,
  removeAlternativeMedias: true,
  removeAlternativeImages: true,
  groupDuplicateImages: true}).then(x => x.content)`, &html, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
		chromedp.CaptureScreenshot(&screenshot),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdf, _, err = page.PrintToPDF().
				WithDisplayHeaderFooter(true).
				WithLandscape(false).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return err
	}

	configSize, configDigest, err := snapshotWriter.WriteFile(ctx, "chrome/config.json", []byte(`{}`))
	if err != nil {
		return err
	}

	screenshotSize, screenshotDigest, err := snapshotWriter.WriteFile(ctx, "chrome/screenshot.png", screenshot)
	if err != nil {
		return err
	}

	err = snapshotWriter.Index(ctx, libraries.Manifest{
		MediaType: "application/vnd.larch.artifact.manifest.v1+json",
		Config: libraries.Layer{
			Digest:    configDigest,
			MediaType: "vnd.larch.disk.config.v1+json",
			Size:      configSize,
			Annotations: map[string]string{
				"larch.artifact.path": "chrome/config.json",
			},
		},
		Layers: []libraries.Layer{
			{
				Digest:    screenshotDigest,
				MediaType: "image/png",
				Size:      screenshotSize,
				Annotations: map[string]string{
					"larch.artifact.path": "chrome/screenshot.png",
				},
			},
		},
	})
	if err != nil {
		return err
	}

	pdfSize, pdfDigest, err := snapshotWriter.WriteFile(ctx, "chrome/print.pdf", pdf)
	if err != nil {
		return err
	}

	err = snapshotWriter.Index(ctx, libraries.Manifest{
		MediaType: "application/vnd.larch.artifact.manifest.v1+json",
		Config: libraries.Layer{
			Digest:    configDigest,
			MediaType: "vnd.larch.disk.config.v1+json",
			Size:      configSize,
			Annotations: map[string]string{
				"larch.artifact.path": "chrome/config.json",
			},
		},
		Layers: []libraries.Layer{
			{
				Digest:    pdfDigest,
				MediaType: "application/pdf",
				Size:      pdfSize,
				Annotations: map[string]string{
					"larch.artifact.path": "chrome/page.pdf",
				},
			},
		},
	})
	if err != nil {
		return err
	}

	singlepageSize, singlepageDigest, err := snapshotWriter.WriteFile(ctx, "chrome/singlepage.html", []byte(html))
	if err != nil {
		return err
	}

	err = snapshotWriter.Index(ctx, libraries.Manifest{
		MediaType: "application/vnd.larch.artifact.manifest.v1+json",
		Config: libraries.Layer{
			Digest:    configDigest,
			MediaType: "vnd.larch.disk.config.v1+json",
			Size:      configSize,
			Annotations: map[string]string{
				"larch.artifact.path": "chrome/config.json",
			},
		},
		Layers: []libraries.Layer{
			{
				Digest:    singlepageDigest,
				MediaType: "application/html",
				Size:      singlepageSize,
				Annotations: map[string]string{
					"larch.artifact.path": "chrome/singlepage.html",
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
