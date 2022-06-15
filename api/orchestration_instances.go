package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

func (o *ec2Orchestrator) createInstance(ctx context.Context, req *Ec2InstanceCreateRequest) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to create instance: %s", awsutil.Prettify(req))

	input := &ec2.RunInstancesInput{
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		InstanceType:     req.Type,
		ImageId:          req.Image,
		SubnetId:         req.Subnet,
		SecurityGroupIds: req.Sgs,
		KeyName:          req.Key,
		UserData:         req.Userdata64,
	}

	if req.BlockDevices != nil {
		input.BlockDeviceMappings = blockDeviceMappingsFromRequest(req.BlockDevices)
	}

	if req.InstanceProfile != nil {
		input.IamInstanceProfile = &ec2.IamInstanceProfileSpecification{
			Name: req.InstanceProfile,
		}
	}

	// set CpuCredits parameter for burstable instances
	// default to standard, unless specified
	if strings.HasPrefix(aws.StringValue(req.Type), "t") {
		cpucredits := aws.String("standard")
		if req.CpuCredits != nil {
			cpucredits = req.CpuCredits
		}
		input.CreditSpecification = &ec2.CreditSpecificationRequest{
			CpuCredits: cpucredits,
		}
	}

	out, err := o.ec2Client.CreateInstance(ctx, input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(out.InstanceId), nil
}

func (o *ec2Orchestrator) deleteInstance(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to delete instance %s", id)

	if err := o.ec2Client.DeleteInstance(ctx, id); err != nil {
		return err
	}

	return nil
}

func blockDeviceMappingsFromRequest(r []Ec2BlockDevice) []*ec2.BlockDeviceMapping {
	blockDeviceMappings := []*ec2.BlockDeviceMapping{}

	for _, bd := range r {
		blockDeviceMappings = append(blockDeviceMappings, &ec2.BlockDeviceMapping{
			DeviceName: bd.DeviceName,
			Ebs: &ec2.EbsBlockDevice{
				Encrypted:  bd.Ebs.Encrypted,
				VolumeSize: bd.Ebs.VolumeSize,
				VolumeType: bd.Ebs.VolumeType,
			},
		})
	}

	return blockDeviceMappings
}

// instancesState is used to start, stop and reboot a given instance
func (o *ec2Orchestrator) instancesState(ctx context.Context, state string, ids ...string) error {
	if len(ids) == 0 || state == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	state = strings.ToLower(state)
	switch state {
	case "start":
		return o.ec2Client.StartInstance(ctx, ids...)
	case "stop", "poweroff":
		isForce := state == "poweroff"
		return o.ec2Client.StopInstance(ctx, isForce, ids...)
	case "reboot":
		return o.ec2Client.RebootInstance(ctx, ids...)
	default:
		msg := fmt.Sprintf("unknown power state %q", state)
		return apierror.New(apierror.ErrBadRequest, msg, nil)
	}
}

func (o *ssmOrchestrator) sendInstancesCommand(ctx context.Context, req *SsmCommandRequest, id ...string) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to send command: %s", awsutil.Prettify(req))
	input := &ssm.SendCommandInput{
		DocumentName:   aws.String(req.DocumentName),
		Parameters:     req.Parameters,
		TimeoutSeconds: req.TimeoutSeconds,
		InstanceIds:    aws.StringSlice(id),
	}
	cmd, err := o.ssmClient.SendCommand(ctx, input)
	if err != nil {
		return "", err
	}
	return aws.StringValue(cmd.CommandId), nil
}

func (o *ec2Orchestrator) attachVolume(ctx context.Context, req *Ec2VolumeAttachmentRequest, id string) (string, error) {
	if req == nil || id == "" {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Debugf("got request to attach volume: %s", awsutil.Prettify(req))

	input := &ec2.AttachVolumeInput{
		Device:     req.Device,
		InstanceId: aws.String(id),
		VolumeId:   req.VolumeID,
	}
	attributeInput := &ec2.ModifyInstanceAttributeInput{
		Attribute: aws.String("blockDeviceMapping"),
		BlockDeviceMappings: []*ec2.InstanceBlockDeviceMappingSpecification{{
			DeviceName: req.Device,
			Ebs: &ec2.EbsInstanceBlockDeviceSpecification{
				DeleteOnTermination: req.DeleteOnTermination,
				VolumeId:            req.VolumeID,
			},
		},
		},
	}

	out, err := o.ec2Client.AttachVolume(ctx, input, attributeInput)
	if err != nil {
		return "", err
	}
	_, err = o.ec2Client.Service.ModifyInstanceAttributeWithContext(ctx, attributeInput)
	if err != nil {
		return "", common.ErrCode("failed to add instance attributes", err)
	}

	return out, nil
}
