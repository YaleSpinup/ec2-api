package api

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (o *ec2Orchestrator) createSnapshot(ctx context.Context, req *Ec2SnapshotCreateRequest) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to create snapshot: %s", awsutil.Prettify(req))

	instance, err := o.ec2Client.GetVolume(ctx, id)
	if err != nil {
		return "", apierror.New(apierror.ErrBadRequest, err.Error(), nil)
	}

	input := &ec2.CreateSnapshotInput{
		VolumeId:    req.VolumeId,
		Description: req.Description,
	}

	if aws.BoolValue(req.CopyTags) {
		input.TagSpecifications = []*ec2.TagSpecification{{
			ResourceType: aws.String("snapshot"),
			Tags:         instance.Tags,
		}}
	}

	snapshotId, err := o.ec2Client.CreateSnapshot(ctx, input)
	if err != nil {
		return "", err
	}

	return snapshotId, nil
}
