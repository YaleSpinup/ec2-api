package ssm

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

func (s *SSM) GetCommandInvocation(ctx context.Context, instanceId, commandId string) (*ssm.GetCommandInvocationOutput, error) {
	if instanceId == "" || commandId == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "both instanceId and commandId should be present", nil)
	}
	out, err := s.Service.GetCommandInvocationWithContext(ctx, &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandId),
		InstanceId: aws.String(instanceId),
	})
	if err != nil {
		return nil, common.ErrCode("failed to get command invocation", err)
	}
	log.Debugf("got output describing SSM Command: %+v", out)
	return out, nil
}

func (s *SSM) SendCommand(ctx context.Context, input *ssm.SendCommandInput) (*ssm.Command, error) {
	if input == nil || aws.StringValue(input.DocumentName) == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("sending command with doc name: %s, params: %+v", aws.StringValue(input.DocumentName), input.Parameters)

	out, err := s.Service.SendCommandWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to send command", err)
	}
	log.Debugf("got output sending command: %+v", out)
	return out.Command, nil
}
