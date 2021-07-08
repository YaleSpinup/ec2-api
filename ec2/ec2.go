package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	log "github.com/sirupsen/logrus"
)

// Ec2 is a wrapper around the aws EC2 service with some default config info
type Ec2 struct {
	session         *session.Session
	Service         ec2iface.EC2API
	DefaultKMSKeyId string
	DefaultSgs      []string
	DefaultSubnets  []string
	org             string
}

type EC2Option func(*Ec2)

// New creates a new Ec2
func New(opts ...EC2Option) Ec2 {
	e := Ec2{}

	for _, opt := range opts {
		opt(&e)
	}

	if e.session != nil {
		e.Service = ec2.New(e.session)
	}

	return e
}

func WithSession(sess *session.Session) EC2Option {
	return func(e *Ec2) {
		log.Debug("using aws session")
		e.session = sess
	}
}

func WithCredentials(key, secret, token, region string) EC2Option {
	return func(e *Ec2) {
		log.Debugf("creating new session with key id %s in region %s", key, region)
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, token),
			Region:      aws.String(region),
		}))
		e.session = sess
	}
}

func WithDefaultKMSKeyId(keyId string) EC2Option {
	return func(e *Ec2) {
		log.Debugf("setting default kms keyid %s", keyId)
		e.DefaultKMSKeyId = keyId
	}
}

func WithDefaultSgs(sgs []string) EC2Option {
	return func(e *Ec2) {
		log.Debugf("setting default security groups %+v", sgs)
		e.DefaultSgs = sgs
	}
}

func WithDefaultSubnets(subnets []string) EC2Option {
	return func(e *Ec2) {
		log.Debugf("setting default subnets %+v", subnets)
		e.DefaultSubnets = subnets
	}
}

func WithOrg(org string) EC2Option {
	return func(e *Ec2) {
		log.Debugf("setting org to %s", org)
		e.org = org
	}
}
