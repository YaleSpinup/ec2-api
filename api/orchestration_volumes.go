package api

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
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

func (o *ec2Orchestrator) modifyVolume(ctx context.Context, req *Ec2VolumeUpdateRequest, id string) (string, error) {
	if req == nil || id == "" {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to modify volume: %s", awsutil.Prettify(req))

	input := &ec2.ModifyVolumeInput{
		Iops:       req.Iops,
		VolumeType: req.Type,
		Size:       req.Size,
		VolumeId:   aws.String(id),
	}
	out, err := o.ec2Client.ModifyVolume(ctx, input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(out.VolumeId), nil
}

func (o *ec2Orchestrator) getVolumes(ctx context.Context, ids ...string) ([]*ec2.Volume, error) {
	f := &ec2.Filter{
		Name:   aws.String("volume-id"),
		Values: aws.StringSlice(ids),
	}
	input := ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{f},
	}

	out, err := o.ec2Client.Service.DescribeVolumesWithContext(ctx, &input)
	if err != nil {
		return nil, common.ErrCode("getting details for volumes", err)
	}
	return out.Volumes, nil

}
