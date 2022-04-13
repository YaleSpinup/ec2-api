package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListVolumes(ctx context.Context, org string, per int64, next *string) ([]map[string]*string, *string, error) {
	log.Infof("listing volumes")

	var filters []*ec2.Filter
	if org != "" {
		filters = []*ec2.Filter{inOrg(org)}
	}

	input := ec2.DescribeVolumesInput{
		Filters: filters,
	}

	if next != nil {
		input.NextToken = next
	}

	if per != 0 {
		input.MaxResults = aws.Int64(per)
	}

	out, err := e.Service.DescribeVolumesWithContext(ctx, &input)
	if err != nil {
		return nil, nil, common.ErrCode("listing volumes", err)
	}

	log.Debugf("returning list of %d volumes", len(out.Volumes))

	list := make([]map[string]*string, len(out.Volumes))
	for i, v := range out.Volumes {
		list[i] = map[string]*string{
			"id": v.VolumeId,
		}
	}

	return list, out.NextToken, nil
}

func (e *Ec2) GetVolume(ctx context.Context, ids ...string) ([]*ec2.Volume, error) {
	if len(ids) == 0 {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting details about volume ids %+v", ids)

	input := ec2.DescribeVolumesInput{
		VolumeIds: aws.StringSlice(ids),
	}

	out, err := e.Service.DescribeVolumesWithContext(ctx, &input)
	if err != nil {
		return nil, common.ErrCode("getting details for volumes", err)
	}

	log.Debugf("returning volumes: %+v", out.Volumes)

	return out.Volumes, nil
}

// ListVolumeModifications returns the modifications events for a volume
func (e *Ec2) ListVolumeModifications(ctx context.Context, id string) ([]*ec2.VolumeModification, error) {
	if id == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting modifications for volume %s", id)

	modifications := []*ec2.VolumeModification{}

	input := ec2.DescribeVolumesModificationsInput{
		Filters: []*ec2.Filter{
			withVolumeId(id),
		},
		MaxResults: aws.Int64(500),
	}

	for {
		out, err := e.Service.DescribeVolumesModificationsWithContext(ctx, &input)
		if err != nil {
			return nil, common.ErrCode("describing modifications for volume", err)
		}

		log.Debugf("got describe volume modifications output %+v", out)

		modifications = append(modifications, out.VolumesModifications...)

		if out.NextToken != nil {
			input.NextToken = out.NextToken
			continue
		}

		break
	}

	return modifications, nil
}

// ListVolumeSnapshots returns the snapshots for a volume
func (e *Ec2) ListVolumeSnapshots(ctx context.Context, id string) ([]string, error) {
	if id == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting snapshots for volume %s", id)

	snapshots := []string{}

	input := ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			withVolumeId(id),
		},
		MaxResults: aws.Int64(1000),
	}

	for {
		out, err := e.Service.DescribeSnapshotsWithContext(ctx, &input)
		if err != nil {
			return nil, common.ErrCode("describing snapshots for volume", err)
		}

		log.Debugf("got describe volume snapshots output %+v", out)

		for _, s := range out.Snapshots {
			snapshots = append(snapshots, aws.StringValue(s.SnapshotId))
		}

		if out.NextToken != nil {
			input.NextToken = out.NextToken
			continue
		}

		break
	}

	return snapshots, nil
}
