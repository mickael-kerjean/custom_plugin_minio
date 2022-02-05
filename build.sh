#!/bin/bash
set -e

###################
# CUSTOM CODE
rm -rf filestash
git clone --depth 1 https://github.com/mickael-kerjean/filestash || true
rm -rf filestash/server/plugin/plg_backend_s3sts || true
rm -rf filestash/server/plugin/plg_authorisation_readonly || true
rm -rf filestash/server/plugin/plg_search_elasticsearch || true
cp -R server filestash/
cp -R client filestash/

cat > filestash/server/plugin/index.go <<EOF
package plugin

import (
    . "github.com/mickael-kerjean/filestash/server/common"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_authenticate_admin"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_authenticate_ldap"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_authenticate_openid"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_authenticate_passthrough"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_authenticate_saml"
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
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_search_stateless"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_security_scanner"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_security_svg"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_starter_http"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_starter_tor"
    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_video_transcoder"

    _ "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3sts"
	  _ "github.com/mickael-kerjean/filestash/server/plugin/plg_authorisation_readonly"
	  _ "github.com/mickael-kerjean/filestash/server/plugin/plg_search_elasticsearch"
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
find server/plugin/plg_* -type f -name "install.sh" -exec {} \;
go mod vendor
go generate -x ./server/...
make build_backend
timeout 1 ./dist/filestash || true
cd ..

#cd ..
BUILD_DATE=$(date '+%Y%m%d-%H%M')
docker pull machines/filestash
docker pull $USER/filestash
docker build -t $USER/filestash .
docker tag $USER/filestash $USER/filestash:$BUILD_DATE
docker push $USER/filestash
docker push $USER/filestash:$BUILD_DATE
