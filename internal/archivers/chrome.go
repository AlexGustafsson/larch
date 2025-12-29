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

	// import { writeFile } from "fs/promises"

	// import { chromium } from "playwright"
	// // @ts-expect-error
	// import { getHookScriptSource, getScriptSource, getZipScriptSource } from "single-file-cli/lib/single-file-script.js"

	// // Fix Access-Control errors when saving resources
	// const browser = await chromium.launch({ args: ["--disable-web-security"] })
	// const page = await browser.newPage()

	// // https://github.com/gildas-lormeau/single-file-cli/blob/v2.0.75/lib/cdp-client.js#L235-L243
	// await page.addInitScript({ content: getHookScriptSource() })
	// await page.addInitScript({ content: (await getScriptSource({})) + "; window.singlefile = singlefile" })

	// await page.goto("https://developer.mozilla.org/en-US/")
	// await page.waitForTimeout(3000)

	// // https://github.com/gildas-lormeau/single-file-cli/blob/v2.0.75/single-file-cli-api.js#L258
	// // https://github.com/gildas-lormeau/single-file-cli/blob/v2.0.75/lib/cdp-client.js#L332
	// // https://github.com/gildas-lormeau/single-file-core/blob/212a657/single-file.js#L125
	// // @ts-expect-error
	// const pageData = await page.evaluate(async options => await singlefile.getPageData(options), {
	//   zipScript: getZipScriptSource(),
	//   // Some flags in the link below should work; just convert them to CamelCase
	//   // https://github.com/gildas-lormeau/single-file-cli/blob/v2.0.75/options.js#L33-L159
	//   // For example,
	//   blockImages: true,
	//   compressHTML: false
	// })

	// await page.close()
	// await browser.close()
	// await writeFile("./mdn.html", pageData.content)

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

	w, err := snapshotWriter.NextWriter(ctx, "test.png")
	if err != nil {
		return err
	}

	w.Write(screenshot)
	w.Close()

	w, err = snapshotWriter.NextWriter(ctx, "test.pdf")
	if err != nil {
		return err
	}

	w.Write(pdf)
	w.Close()

	w, err = snapshotWriter.NextWriter(ctx, "test.html")
	if err != nil {
		return err
	}

	w.Write([]byte(html))
	w.Close()

	return nil
}
