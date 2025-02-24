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

func (o *ec2Orchestrator) createSnapshot(ctx context.Context, req *Ec2SnapshotCreateRequest) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to create snapshot: %s", awsutil.Prettify(req))

	volumes, err := o.ec2Client.GetVolume(ctx, *req.VolumeId)
	if err != nil {
		return "", apierror.New(apierror.ErrBadRequest, err.Error(), nil)
	}
	if len(volumes) == 0 {
		return "", apierror.New(apierror.ErrBadRequest, "volume information not found", nil)
	}

	input := &ec2.CreateSnapshotInput{
		VolumeId:    req.VolumeId,
		Description: req.Description,
	}

	if aws.BoolValue(req.CopyTags) {
		input.TagSpecifications = []*ec2.TagSpecification{{
			ResourceType: aws.String("snapshot"),
			Tags:         volumes[0].Tags,
		}}
	}

	snapshotId, err := o.ec2Client.CreateSnapshot(ctx, input)
	if err != nil {
		return "", err
	}

	return snapshotId, nil
}

func (o *ec2Orchestrator) deleteSnapshot(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Debugf("got request to delete snapshot %s", id)

	input := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(id),
	}

	if err := o.ec2Client.DeleteSnapshot(ctx, input); err != nil {
		return err
	}

	return nil
}

func (o *ec2Orchestrator) listSnapshots(ctx context.Context, perPage int64, pageToken *string, filters ...*ec2.Filter) ([]*ec2.Snapshot, *string, error) {

	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: aws.StringSlice([]string{"self"}),
		Filters:  filters,
	}

	if pageToken != nil {
		input.NextToken = pageToken
	}

	if perPage != 0 {
		input.MaxResults = aws.Int64(perPage)
	}

	out, err := o.ec2Client.ListSnapshots(ctx, input)
	if err != nil {
		return nil, nil, common.ErrCode("listing snapshots", err)
	}

	log.Debugf("returning list of %d snapshots", len(out.Snapshots))

	return out.Snapshots, out.NextToken, nil
}
