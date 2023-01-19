package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListSnapshots(ctx context.Context, input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("list snapshots: %+v", input)
	out, err := e.Service.DescribeSnapshotsWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to list snapshot", err)
	}

	if out == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "unexpected list snapshot response", nil)
	}

	return out, nil
}

func (e *Ec2) GetSnapshot(ctx context.Context, ids ...string) ([]*ec2.Snapshot, error) {
	if len(ids) == 0 {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting details about snapshot ids %+v", ids)

	input := ec2.DescribeSnapshotsInput{
		OwnerIds:    aws.StringSlice([]string{"self"}),
		SnapshotIds: aws.StringSlice(ids),
	}

	out, err := e.Service.DescribeSnapshotsWithContext(ctx, &input)
	if err != nil {
		return nil, common.ErrCode("getting details for snapshots", err)
	}

	log.Debugf("returning snapshots: %+v", out.Snapshots)

	return out.Snapshots, nil
}

func (e *Ec2) CreateSnapshot(ctx context.Context, input *ec2.CreateSnapshotInput) (string, error) {
	if input == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("creating snapshot: %s", aws.StringValue(input.VolumeId))
	out, err := e.Service.CreateSnapshotWithContext(ctx, input)
	if err != nil {
		return "", common.ErrCode("failed to create snapshot", err)
	}

	log.Debugf("got output creating snapshot: %+v", out)

	if out == nil || len(aws.StringValue(out.SnapshotId)) == 0 {
		return "", apierror.New(apierror.ErrBadRequest, "unexpected create snapshot response", nil)
	}

	return aws.StringValue(out.SnapshotId), nil
}

func (e *Ec2) DeleteSnapshot(ctx context.Context, input *ec2.DeleteSnapshotInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("deleting snapshot %s", *input.SnapshotId)

	_, err := e.Service.DeleteSnapshotWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to delete snapshot", err)
	}
	return nil
}
