package resourcegroups

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroups/resourcegroupsiface"
	log "github.com/sirupsen/logrus"
)

// ResourceGroups is a wrapper around the aws Resource Groups service
type ResourceGroups struct {
	session *session.Session
	Service resourcegroupsiface.ResourceGroupsAPI
}

type Option func(*ResourceGroups)

// New creates a new ResourceGroups
func New(opts ...Option) *ResourceGroups {
	rg := ResourceGroups{}

	for _, opt := range opts {
		opt(&rg)
	}

	if rg.session != nil {
		rg.Service = resourcegroups.New(rg.session)
	}

	return &rg
}

func WithSession(sess *session.Session) Option {
	return func(rg *ResourceGroups) {
		log.Debug("using aws session")
		rg.session = sess
	}
}

func WithCredentials(key, secret, token, region string) Option {
	return func(rg *ResourceGroups) {
		log.Debugf("creating new session with key id %s in region %s", key, region)
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, token),
			Region:      aws.String(region),
		}))
		rg.session = sess
	}
}
