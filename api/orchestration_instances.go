package api

import (
	"context"
	"strings"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
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
