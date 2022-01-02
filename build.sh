#!/bin/bash
set -e

###################
# CUSTOM CODE
rm -rf filestash
git clone --depth 1 https://github.com/mickael-kerjean/filestash || true
cp -R plg_backend_s3sts filestash/server/plugin/

cat > filestash/server/plugin/index.go <<EOF
package plugin

import (
    . "github.com/mickael-kerjean/filestash/server/common"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_ftp"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_git"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_ldap"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_local"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_mysql"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_nop"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_samba"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_sftp"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_editor_onlyoffice"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_handler_console"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_handler_syncthing"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_image_light"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_security_scanner"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_security_svg"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_starter_http"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_starter_tor"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_video_transcoder"

    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3sts"
)

func init() {
    Log.Debug("Plugin loader")
}
EOF

cd filestash
###################
# Install dependencies
npm install # frontend dependencies
make build_init # install the required static libraries
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
docker build -t $USER/filestash .
docker push $USER/filestash
