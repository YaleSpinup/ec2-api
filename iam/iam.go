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
	// DefaultKMSKeyId string
	// DefaultSgs      []string
	// DefaultSubnets  []string
	// org             string
}

type IAMOption func(*Iam)

// New creates a new Ec2
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

// func WithCredentials(key, secret, token, region string) IAMOption {
// 	return func(e *Iam) {
// 		log.Debugf("creating new session with key id %s in region %s", key, region)
// 		sess := session.Must(session.NewSession(&aws.Config{
// 			Credentials: credentials.NewStaticCredentials(key, secret, token),
// 			Region:      aws.String(region),
// 		}))
// 		e.session = sess
// 	}
// }

// func WithDefaultKMSKeyId(keyId string) IAMOption {
// 	return func(e *Iam) {
// 		log.Debugf("setting default kms keyid %s", keyId)
// 		e.DefaultKMSKeyId = keyId
// 	}
// }

// func WithDefaultSgs(sgs []string) IAMOption {
// 	return func(e *Iam) {
// 		log.Debugf("setting default security groups %+v", sgs)
// 		e.DefaultSgs = sgs
// 	}
// }

// func WithDefaultSubnets(subnets []string) IAMOption {
// 	return func(e *Iam) {
// 		log.Debugf("setting default subnets %+v", subnets)
// 		e.DefaultSubnets = subnets
// 	}
// }

// func WithOrg(org string) IAMOption {
// 	return func(e *Iam) {
// 		log.Debugf("setting org to %s", org)
// 		e.org = org
// 	}
// }
