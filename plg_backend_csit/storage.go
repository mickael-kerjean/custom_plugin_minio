package plg_backend_sg

import (
	. "github.com/mickael-kerjean/filestash/server/common"
	s3 "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3"
)

func init() {
	s3Config()
}

var s3Config = func() map[string]string {
	return map[string]string{
		"endpoint": Config.Get("sg.minio.endpoint").Schema(func(f *FormElement) *FormElement {
			if f == nil {
				f = &FormElement{}
			}
			f.Default = "https://play.minio.io:9000/"
			f.Name = "s3_endpoint"
			f.Type = "text"
			f.Placeholder = "Endpoint for S3"
			return f
		}).String(),
		"access_key_id": Config.Get("sg.minio.access_key_id").Schema(func(f *FormElement) *FormElement {
			if f == nil {
				f = &FormElement{}
			}
			f.Default = "Q3AM3UQ867SPQQA43P2F"
			f.Name = "access_key_id"
			f.Type = "text"
			f.Placeholder = "access_key_id"
			return f
		}).String(),
		"secret_access_key": Config.Get("sg.minio.secret_access_key").Schema(func(f *FormElement) *FormElement {
			if f == nil {
				f = &FormElement{}
			}
			f.Default = "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
			f.Name = "secret_access_key"
			f.Type = "text"
			f.Placeholder = "secret_access_key"
			return f
		}).String(),
	}
}

func StorageBackend(app *App) (IBackend, error) {
	return s3.S3Backend{}.Init(s3Config(), app)
}
