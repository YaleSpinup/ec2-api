package api

import (
	"context"

	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/YaleSpinup/ec2-api/iam"
	"github.com/YaleSpinup/ec2-api/ssm"
	log "github.com/sirupsen/logrus"
)

type sessionParams struct {
	role         string
	inlinePolicy string
	policyArns   []string
}

type ec2Orchestrator struct {
	ec2Client *ec2.Ec2
	server    *server
}

func (s *server) newEc2Orchestrator(ctx context.Context, sp *sessionParams) (*ec2Orchestrator, error) {
	log.Debugf("initializing ec2Orchestrator")

	session, err := s.assumeRole(
		ctx,
		s.session.ExternalID,
		sp.role,
		sp.inlinePolicy,
		sp.policyArns...,
	)
	if err != nil {
		return nil, err
	}

	return &ec2Orchestrator{
		ec2Client: ec2.New(ec2.WithSession(session.Session)),
		server:    s,
	}, nil
}

type ssmOrchestrator struct {
	ssmClient *ssm.SSM
	server    *server
}

func (s *server) newSSMOrchestrator(ctx context.Context, sp *sessionParams) (*ssmOrchestrator, error) {
	log.Debugf("initializing ssmOrchestrator")

	session, err := s.assumeRole(
		ctx,
		s.session.ExternalID,
		sp.role,
		sp.inlinePolicy,
		sp.policyArns...,
	)
	if err != nil {
		return nil, err
	}

	return &ssmOrchestrator{
		ssmClient: ssm.New(ssm.WithSession(session.Session)),
		server:    s,
	}, nil
}

type iamOrchestrator struct {
	iamClient *iam.Iam
	server    *server
}

func (s *server) newIAMOrchestrator(ctx context.Context, sp *sessionParams) (*iamOrchestrator, error) {
	log.Debugf("initializing iamOrchestrator")

	session, err := s.assumeRole(
		ctx,
		s.session.ExternalID,
		sp.role,
		sp.inlinePolicy,
		sp.policyArns...,
	)
	if err != nil {
		return nil, err
	}

	return &iamOrchestrator{
		iamClient: iam.New(iam.WithSession(session.Session)),
		server:    s,
	}, nil
}
