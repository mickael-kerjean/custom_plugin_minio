package plg_backend_s3sts

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	. "github.com/mickael-kerjean/filestash/server/common"
	"io"
	"os"

	s3 "github.com/mickael-kerjean/filestash/server/plugin/plg_backend_s3"
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

		config := &aws.Config{
			Region:   aws.String("us-east-2"),
			Endpoint: aws.String(stsEndpoint()),
		}
		svc := sts.New(session.New(config))

		input := &sts.AssumeRoleWithWebIdentityInput{
			DurationSeconds:  aws.Int64(3600),
			RoleArn:          aws.String("arn:aws:iam::123456789012:role/FederatedWebIdentityRole"),
			RoleSessionName:  aws.String("filestash"),
			WebIdentityToken: aws.String(params["access_token"]),
		}

		result, err := svc.AssumeRoleWithWebIdentity(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case sts.ErrCodeMalformedPolicyDocumentException:
					Log.Error(sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
				case sts.ErrCodePackedPolicyTooLargeException:
					Log.Error(sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
				case sts.ErrCodeIDPRejectedClaimException:
					Log.Error(sts.ErrCodeIDPRejectedClaimException, aerr.Error())
				case sts.ErrCodeIDPCommunicationErrorException:
					Log.Error(sts.ErrCodeIDPCommunicationErrorException, aerr.Error())
				case sts.ErrCodeInvalidIdentityTokenException:
					Log.Error(sts.ErrCodeInvalidIdentityTokenException, aerr.Error())
				case sts.ErrCodeExpiredTokenException:
					Log.Error(sts.ErrCodeExpiredTokenException, aerr.Error())
				case sts.ErrCodeRegionDisabledException:
					Log.Error(sts.ErrCodeRegionDisabledException, aerr.Error())
				default:
					Log.Error(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				Log.Error(err.Error())
			}
			return nil, err
		}
		credentials := result.Credentials
		params["access_key_id"] = *credentials.AccessKeyId
		params["secret_access_key"] = *credentials.SecretAccessKey
		params["session_token"] = *credentials.SessionToken
		fmt.Println(*credentials.AccessKeyId)

		params["endpoint"] = stsEndpoint()
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
