package plg_backend_s3sts

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	. "github.com/mickael-kerjean/filestash/server/common"
	s3 "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3"
	"io"
	"os"
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
			Log.Error("s3sts::init 'OAuth2Authenticate' %+v", err)
			return nil, ErrAuthenticationFailed
		}
		params["code"] = ""
		params["id_token"] = token
	}

	if params["id_token"] != "" {
		if err := OpenIDVerifyToken(params["id_token"]); err != nil {
			Log.Error("s3sts::init 'OpenIDVerifyToken'", err.Error())
			return nil, ErrAuthenticationFailed
		}

		params["endpoint"] = stsEndpoint()
		config := &aws.Config{
			Region:   aws.String("us-east-2"),
			Endpoint: aws.String(params["endpoint"]),
		}
		svc := sts.New(session.New(config))
		Log.Debug("s3sts - sts session ok")

		input := &sts.AssumeRoleWithWebIdentityInput{
			DurationSeconds:  aws.Int64(3600),
			RoleArn:          aws.String("arn:aws:iam::123456789012:role/FederatedWebIdentityRole"),
			RoleSessionName:  aws.String("filestash"),
			WebIdentityToken: aws.String(params["id_token"]),
		}

		result, err := svc.AssumeRoleWithWebIdentity(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case sts.ErrCodeMalformedPolicyDocumentException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
				case sts.ErrCodePackedPolicyTooLargeException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
				case sts.ErrCodeIDPRejectedClaimException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodeIDPRejectedClaimException, aerr.Error())
				case sts.ErrCodeIDPCommunicationErrorException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodeIDPCommunicationErrorException, aerr.Error())
				case sts.ErrCodeInvalidIdentityTokenException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodeInvalidIdentityTokenException, aerr.Error())
				case sts.ErrCodeExpiredTokenException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodeExpiredTokenException, aerr.Error())
				case sts.ErrCodeRegionDisabledException:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", sts.ErrCodeRegionDisabledException, aerr.Error())
				default:
					Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				Log.Error("s3sts::init 'AssumeRoleWithWebIdentity'", err.Error())
			}
			return nil, ErrAuthenticationFailed
		}

		credentials := result.Credentials
		params["id_token"] = ""
		params["access_key_id"] = *credentials.AccessKeyId
		params["secret_access_key"] = *credentials.SecretAccessKey
		params["session_token"] = *credentials.SessionToken
		Log.Debug("s3sts - success param [%v]", params)
	}

	if params["access_key_id"] != "" && params["secret_access_key"] != "" && params["session_token"] != "" {
		Log.Debug("s3sts - sts access_key_id [%s]", params["access_key_id"])
		backend, err := s3.S3Backend{}.Init(params, app)
		if err != nil {
			Log.Error("s3sts::init 's3 init'", err.Error())
			return nil, ErrAuthenticationFailed
		}
		this.Backend = backend
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

func checkS3Error(ops string, err error) error {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "InvalidAccessKeyId" {
			Log.Error("s3sts::%s %s", ops, aerr.Error())
			return ErrNotAuthorized
		}
	}
	return err
}

func (this S3STSBackend) Ls(path string) ([]os.FileInfo, error) {
	fileInfo, err := this.Backend.Ls(path)
	if err != nil {
		return nil, checkS3Error("ls", err)
	}
	return fileInfo, nil
}

func (this S3STSBackend) Cat(path string) (io.ReadCloser, error) {
	body, err := this.Backend.Cat(path)
	if err != nil {
		return nil, checkS3Error("cat", err)
	}
	return body, nil
}

func (this S3STSBackend) Mkdir(path string) error {
	return checkS3Error("mkdir", this.Backend.Mkdir(path))
}

func (this S3STSBackend) Rm(path string) error {
	return checkS3Error("rm", this.Backend.Rm(path))
}

func (this S3STSBackend) Mv(from, to string) error {
	return checkS3Error("mv", this.Backend.Mv(from, to))
}

func (this S3STSBackend) Save(path string, content io.Reader) error {
	return checkS3Error("save", this.Backend.Save(path, content))
}

func (this S3STSBackend) Touch(path string) error {
	return checkS3Error("touch", this.Backend.Touch(path))
}
