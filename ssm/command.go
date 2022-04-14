package ssm

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

func (s *SSM) GetCommandInvocation(ctx context.Context, instance_id, command_id string) (*ssm.GetCommandInvocationOutput, error) {
	if instance_id == "" || command_id == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "both instance_id and command_id should be present", nil)
	}
	out, err := s.Service.GetCommandInvocationWithContext(ctx, &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(command_id),
		InstanceId: aws.String(instance_id),
	})
	if err != nil {
		return nil, common.ErrCode("failed to get command invocation", err)
	}
	log.Debugf("got output describing SSM Command: %+v", out)
	return out, nil
}
