#!/usr/bin/env bash

set -e

TAG=v2.0.83

curl --location --output internal/singlefile/vendor/single-file-bundle.js "https://raw.githubusercontent.com/gildas-lormeau/single-file-cli/refs/tags/$TAG/lib/single-file-bundle.js"
sed -i -e 's/export { script, zipScript, hookScript };//' internal/singlefile/vendor/single-file-bundle.js
sed -i -e "1i /* AGPL licensed. SEE: https://raw.githubusercontent.com/gildas-lormeau/single-file-cli/refs/tags/$TAG/LICENSE. */" internal/singlefile/vendor/single-file-bundle.js
