package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
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
		return nil, nil, ErrCode("listing volumes", err)
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
		return nil, ErrCode("getting details for volumes", err)
	}

	log.Debugf("returning volumes: %+v", out.Volumes)

	return out.Volumes, nil
}