package ssm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	log "github.com/sirupsen/logrus"
)

type SSM struct {
	session *session.Session
	Service ssmiface.SSMAPI
}

type SSMOption func(*SSM)

// New creates a new SSM
func New(opts ...SSMOption) *SSM {
	e := SSM{}

	for _, opt := range opts {
		opt(&e)
	}

	if e.session != nil {
		e.Service = ssm.New(e.session)
	}

	return &e
}

func WithSession(sess *session.Session) SSMOption {
	return func(e *SSM) {
		log.Debug("using aws session")
		e.session = sess
	}
}

func WithCredentials(key, secret, token, region string) SSMOption {
	return func(e *SSM) {
		log.Debugf("creating new session with key id %s in region %s", key, region)
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, token),
			Region:      aws.String(region),
		}))
		e.session = sess
	}
}
