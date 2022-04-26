package api

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (o *ec2Orchestrator) createVolume(ctx context.Context, req *Ec2VolumeCreateRequest) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to create volume: %s", awsutil.Prettify(req))

	input := &ec2.CreateVolumeInput{
		AvailabilityZone: req.AZ,
		Encrypted:        req.Encrypted,
		Iops:             req.Iops,
		KmsKeyId:         req.KmsKeyId,
		Size:             req.Size,
		SnapshotId:       req.SnapshotId,
		VolumeType:       req.Type,
	}

	out, err := o.ec2Client.CreateVolume(ctx, input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(out.VolumeId), nil
}

func (o *ec2Orchestrator) deleteVolume(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to delete volume %s", id)

	if err := o.ec2Client.DeleteVolume(ctx, id); err != nil {
		return err
	}

	return nil
}
