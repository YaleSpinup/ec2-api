package iam

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	log "github.com/sirupsen/logrus"
)

// Iam is a wrapper around the aws IAM service with some default config info
type Iam struct {
	session *session.Session
	Service iamiface.IAMAPI
}

type IAMOption func(*Iam)

// New creates a new Iam
func New(opts ...IAMOption) *Iam {
	i := Iam{}

	for _, opt := range opts {
		opt(&i)
	}

	if i.session != nil {
		i.Service = iam.New(i.session)
	}

	return &i
}

func WithSession(sess *session.Session) IAMOption {
	return func(e *Iam) {
		log.Debug("using aws session")
		e.session = sess
	}
}
