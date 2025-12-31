package chrome

import (
	"context"

	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/singlefile"
	"github.com/chromedp/chromedp"
)

const singlefilescript = `singlefile.getPageData({removeHiddenElements: true,
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
  groupDuplicateImages: true}).then(x => x.content)`

var _ chromedp.Action = (*saveToSinglefileAction)(nil)

type saveToSinglefileAction struct {
	SnapshotWriter libraries.SnapshotWriter
}

// Do implements chromedp.Action.
func (s saveToSinglefileAction) Do(ctx context.Context) error {
	err := chromedp.Evaluate(singlefile.HookScript, nil).Do(ctx)
	if err != nil {
		return err
	}

	err = chromedp.Evaluate(singlefile.Script, nil).Do(ctx)
	if err != nil {
		return err
	}

	var html string
	err = chromedp.Evaluate(singlefilescript, &html, awaitPromise).Do(ctx)
	if err != nil {
		return err
	}

	name := "chrome/singlepage.html"
	size, digest, err := s.SnapshotWriter.WriteArtifact(ctx, name, []byte(html))
	if err != nil {
		return err
	}

	err = s.SnapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
		Digest:      digest,
		ContentType: "application/html",
		Size:        size,
		Annotations: map[string]string{
			"larch.artifact.path": name,
			"larch.artifact.type": "vnd.larch.chrome.singlepage.v1",
		},
	})
	if err != nil {
		return err
	}

	return nil
}
