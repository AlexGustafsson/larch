package chrome

import (
	"context"
	"fmt"

	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var _ chromedp.Action = (*printAction)(nil)

type printAction struct {
	SnapshotWriter libraries.SnapshotWriter
}

// Do implements chromedp.Action.
func (p printAction) Do(ctx context.Context) error {
	pdf, _, err := page.PrintToPDF().
		WithDisplayHeaderFooter(true).
		WithLandscape(false).
		Do(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("chrome/print.pdf")
	size, digest, err := p.SnapshotWriter.WriteArtifact(ctx, name, pdf)
	if err != nil {
		return err
	}

	err = p.SnapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
		Digest:      digest,
		ContentType: "application/pdf",
		Size:        size,
		Annotations: map[string]string{
			"larch.artifact.path": name,
			"larch.artifact.type": "vnd.larch.chrome.pdf.v1",
		},
	})
	if err != nil {
		return err
	}

	return nil
}
