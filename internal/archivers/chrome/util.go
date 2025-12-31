package chrome

import "github.com/chromedp/cdproto/runtime"

func awaitPromise(p *runtime.EvaluateParams) *runtime.EvaluateParams {
	return p.WithAwaitPromise(true)
}
