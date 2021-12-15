package plg_backend_sg

import (
	. "github.com/mickael-kerjean/filestash/server/common"
	"io"
	"os"
)

func init() {
	Backend.Register("sg", MinioSG{})
}

type MinioSG struct {
	Backend IBackend
}

func (this MinioSG) Init(params map[string]string, app *App) (IBackend, error) {
	if params["code"] != "" {
		token, err := OAuth2Authenticate(params["code"])
		if err != nil {
			return nil, err
		}
		params["code"] = ""
		params["access_token"] = token
	}
	if params["access_token"] != "" {
		if err := OpenIDVerifyToken(params["access_token"]); err != nil {
			return nil, err
		}
		if err := OAuth2IsAuthenticated(params["access_token"]); err != nil {
			return nil, err
		}
		return StorageBackend(app)
	}
	return this, nil
}

func (this MinioSG) LoginForm() Form {
	return Form{
		Elmnts: []FormElement{
			{
				Name:  "type",
				Type:  "hidden",
				Value: "sg",
			},
			{
				ReadOnly: true,
				Name:     "oauth2",
				Type:     "text",
				Value:    "/api/session/auth/sg",
			},
		},
	}
}

func (this MinioSG) OAuthURL() string {
	return OpenIDGetURL()
}

func (this MinioSG) Ls(path string) ([]os.FileInfo, error) {
	return []os.FileInfo{}, ErrNotImplemented
}

func (this MinioSG) Cat(path string) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}

func (this MinioSG) Mkdir(path string) error {
	return ErrNotImplemented
}

func (this MinioSG) Rm(path string) error {
	return ErrNotImplemented
}

func (this MinioSG) Mv(from, to string) error {
	return ErrNotImplemented
}

func (this MinioSG) Save(path string, content io.Reader) error {
	return ErrNotImplemented
}

func (this MinioSG) Touch(path string) error {
	return ErrNotImplemented
}
