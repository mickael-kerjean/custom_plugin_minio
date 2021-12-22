package plg_backend_s3sts

import (
	. "github.com/mickael-kerjean/filestash/server/common"
	"io"
	"os"

	s3 "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3"
	credentials "github.com/minio/minio-go/v7/pkg/credentials"
)

func init() {
	Backend.Register("s3sts", S3STSBackend{})
	stsEndpoint()
}

var stsEndpoint = func() string {
	return Config.Get("s3sts.sts.endpoint").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Default = "https://localhost:9000"
		f.Name = "endpoint"
		f.Type = "text"
		f.Placeholder = "URL of STS endpoint"
		return f
	}).String()
}

type S3STSBackend struct {
	Backend IBackend
}

func (this S3STSBackend) Init(params map[string]string, app *App) (IBackend, error) {
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
		// STS - exchange token for temp credentials
		var getWebTokenExpiry func() (*credentials.WebIdentityToken, error)
		getWebTokenExpiry = func() (*credentials.WebIdentityToken, error) {
			return &credentials.WebIdentityToken{
				Token:  params["access_token"],
				Expiry: 3600, // TODO:
			}, nil
		}

		params["endpoint"] = stsEndpoint()
		sts, err := credentials.NewSTSWebIdentity(params["endpoint"], getWebTokenExpiry)
		if err != nil {
			Log.Error("Could not get STS credentials: %s", err)
			return nil, err
		}
		credentials, _ := sts.Get()
		params["access_key_id"] = credentials.AccessKeyID
		params["secret_access_key"] = credentials.SecretAccessKey
		params["session_token"] = credentials.SessionToken

		return s3.S3Backend{}.Init(params, app)
	}
	return this, nil
}

func (this S3STSBackend) LoginForm() Form {
	return Form{
		Elmnts: []FormElement{
			{
				Name:  "type",
				Type:  "hidden",
				Value: "s3sts",
			},
			{
				ReadOnly: true,
				Name:     "oauth2",
				Type:     "text",
				Value:    "/api/session/auth/s3sts",
			},
		},
	}
}

func (this S3STSBackend) OAuthURL() string {
	return OpenIDGetURL()
}

func (this S3STSBackend) Ls(path string) ([]os.FileInfo, error) {
	return []os.FileInfo{}, ErrNotImplemented
}

func (this S3STSBackend) Cat(path string) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}

func (this S3STSBackend) Mkdir(path string) error {
	return ErrNotImplemented
}

func (this S3STSBackend) Rm(path string) error {
	return ErrNotImplemented
}

func (this S3STSBackend) Mv(from, to string) error {
	return ErrNotImplemented
}

func (this S3STSBackend) Save(path string, content io.Reader) error {
	return ErrNotImplemented
}

func (this S3STSBackend) Touch(path string) error {
	return ErrNotImplemented
}
