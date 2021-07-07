package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListSnapshots(ctx context.Context, org string, per int64, next *string) ([]map[string]*string, *string, error) {
	log.Infof("listing snapshots")

	var filters []*ec2.Filter
	if org != "" {
		filters = []*ec2.Filter{inOrg(org)}
	}

	input := ec2.DescribeSnapshotsInput{
		OwnerIds: aws.StringSlice([]string{"self"}),
		Filters:  filters,
	}

	if next != nil {
		input.NextToken = next
	}

	if per != 0 {
		input.MaxResults = aws.Int64(per)
	}

	out, err := e.Service.DescribeSnapshotsWithContext(ctx, &input)
	if err != nil {
		return nil, nil, ErrCode("listing snapshots", err)
	}

	log.Debugf("returning list of %d snapshots", len(out.Snapshots))

	list := make([]map[string]*string, len(out.Snapshots))
	for i, s := range out.Snapshots {
		list[i] = map[string]*string{
			"id": s.SnapshotId,
		}
	}

	return list, out.NextToken, nil
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
		return nil, ErrCode("getting details for snapshots", err)
	}

	log.Debugf("returning snapshots: %+v", out.Snapshots)

	return out.Snapshots, nil
}
