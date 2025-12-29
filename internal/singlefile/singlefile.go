package singlefile

import (
	_ "embed" // Embed js
	"regexp"
)

// SEE: https://github.com/gildas-lormeau/single-file-core/blob/main/single-file-bootstrap.js.
//
// SEE: ../../tools/scripts/fetch-single-file.sh
//
//go:embed vendor/single-file-bundle.js
var singlefilebundlejs []byte

var Script string
var HookScript string

func init() {
	match := regexp.
		MustCompile(`const script = "(.*)";const hookScript = "(.*)";const zipScript = "(.*)";`).
		FindSubmatch(singlefilebundlejs)
	if match == nil {
		panic("invalid single-file-bundle.js")
	}

	Script = "eval(\"" + string(match[1]) + "\"); window.singlefile = singlefile"
	HookScript = "eval(\"" + string(match[2]) + "\")"
}
