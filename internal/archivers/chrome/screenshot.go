package chrome

import (
	"context"
	"fmt"

	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/chromedp/chromedp"
)

var _ chromedp.Action = (*screenshotAction)(nil)

type screenshotAction struct {
	SnapshotWriter libraries.SnapshotWriter
	Width          int64
	Height         int64
}

// Do implements chromedp.Action.
func (s screenshotAction) Do(ctx context.Context) error {
	err := chromedp.EmulateViewport(s.Width, s.Height).Do(ctx)
	if err != nil {
		return err
	}

	var screenshot []byte
	err = chromedp.CaptureScreenshot(&screenshot).Do(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("chrome/screenshots/%dx%d.png", s.Width, s.Height)
	size, digest, err := s.SnapshotWriter.WriteArtifact(ctx, name, screenshot)
	if err != nil {
		return err
	}

	err = s.SnapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
		Digest:      digest,
		ContentType: "image/png",
		Size:        size,
		Annotations: map[string]string{
			"larch.artifact.path": name,
			"larch.artifact.type": "vnd.larch.chrome.screenshot.v1",
		},
	})
	if err != nil {
		return err
	}

	return nil
}
