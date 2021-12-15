#!/bin/bash
set -e

###################
# CUSTOM CODE
git clone --depth 1 https://github.com/mickael-kerjean/filestash || true
cp -R plg_backend_csit filestash/server/plugin/

cat > filestash/server/plugin/index.go <<EOF
package plugin

import (
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_image_bimg"
	_ "github.com/mickael-kerjean/filestash/server/plugin/plg_starter_http"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_security_scanner"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_security_svg"
    . "github.com/mickael-kerjean/filestash/server/common"

    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_csit"
)

func init() {
    Log.Debug("Plugin loader")
}
EOF

cd filestash
###################
# Prepare Build
mkdir -p ./dist/data/state/config
cp ../config.json ./dist/data/state/config/

###################
# Frontend build
npm install --silent --legacy-peer-deps
make build_frontend

###################
# Backend build
cd filestash && go generate -x ./server/... && cd ../
make build_backend
timeout 1 ./dist/filestash || true

cd ..
docker pull machines/filestash
docker build -t mickaelkerjean/sg .
docker push mickaelkerjean/sg
