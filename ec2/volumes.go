package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// CreateVolume creates a new volume and returns the volume details
func (e *Ec2) CreateVolume(ctx context.Context, input *ec2.CreateVolumeInput) (*ec2.Volume, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("creating volume of type %s, size %d", aws.StringValue(input.VolumeType), aws.Int64Value(input.Size))

	out, err := e.Service.CreateVolumeWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to create volume", err)
	}

	log.Debugf("got output creating volume: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected volume output", nil)
	}

	return out, nil
}

func (e *Ec2) DeleteVolume(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("deleting volume %s", id)

	out, err := e.Service.DeleteVolumeWithContext(ctx, &ec2.DeleteVolumeInput{
		VolumeId: aws.String(id),
	})
	if err != nil {
		return common.ErrCode("failed to delete volume", err)
	}

	log.Debugf("got output deleting volume: %+v", out)

	return nil
}

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

func (e *Ec2) ModifyVolume(ctx context.Context, input *ec2.ModifyVolumeInput) (*ec2.VolumeModification, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("Modifying volume of type %s, size %d, iop %d", aws.StringValue(input.VolumeType), aws.Int64Value(input.Size), aws.Int64Value(input.Iops))

	out, err := e.Service.ModifyVolumeWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to modify volume", err)
	}

	log.Debugf("got output modify volume: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected volume output", nil)
	}

	return out.VolumeModification, nil
}

func (e *Ec2) DetachVolume(ctx context.Context, input *ec2.DetachVolumeInput) (string, error) {
	if input == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("detaching volumes %v, force = %t", input.VolumeId, aws.BoolValue(input.Force))

	out, err := e.Service.DetachVolumeWithContext(ctx, input)
	if err != nil {
		return "", apierror.New(apierror.ErrInternalError, "failed to detach volume", err)
	}

	log.Debugf("got output to detach volume: %+v", out)

	if out == nil {
		return "", apierror.New(apierror.ErrInternalError, "Unexpected detach volume output", nil)
	}

	return aws.StringValue(out.VolumeId), nil
}

func (e *Ec2) AttachVolume(ctx context.Context, input *ec2.AttachVolumeInput) (string, error) {
	if input == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("Attaching volume of device %s", aws.StringValue(input.Device))

	out, err := e.Service.AttachVolumeWithContext(ctx, input)
	if err != nil {
		return "", common.ErrCode("failed to attach volume", err)
	}

	log.Debugf("got output attach volume: %+v", out)

	if out == nil {
		return "", apierror.New(apierror.ErrInternalError, "Unexpected volume output", nil)
	}

	return aws.StringValue(out.VolumeId), nil
}
