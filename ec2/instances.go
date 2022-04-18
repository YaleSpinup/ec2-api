package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// CreateInstance creates a new instance and returns the instance details
func (e *Ec2) CreateInstance(ctx context.Context, input *ec2.RunInstancesInput) (*ec2.Instance, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("creating instance of type %s", aws.StringValue(input.InstanceType))

	out, err := e.Service.RunInstancesWithContext(ctx, input)
	if err != nil {
		return nil, ErrCode("failed to create instance", err)
	}

	log.Debugf("got output creating instance: %+v", out)

	if out == nil || len(out.Instances) != 1 {
		return nil, apierror.New(apierror.ErrBadRequest, "Unexpected instance count", nil)
	}

	return out.Instances[0], nil
}

// DeleteInstance terminates an instance
func (e *Ec2) DeleteInstance(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("deleting instance %s", id)

	out, err := e.Service.TerminateInstancesWithContext(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(id),
		},
	})
	if err != nil {
		return ErrCode("failed to delete instance", err)
	}

	log.Debugf("got output deleting instance: %+v", out)

	return nil
}

// ListInstances lists the instances that are not terminated and not spot
func (e *Ec2) ListInstances(ctx context.Context, org string, per int64, next *string) ([]map[string]*string, *string, error) {
	log.Infof("listing ec2 instances")

	var nextToken *string
	if next != nil {
		nextToken = next
	}

	filters := []*ec2.Filter{
		notTerminated(),
	}

	if org != "" {
		filters = append(filters, inOrg(org))
	}

	out, err := e.Service.DescribeInstancesWithContext(ctx, &ec2.DescribeInstancesInput{
		Filters:    filters,
		MaxResults: aws.Int64(per),
		NextToken:  nextToken,
	})

	if err != nil {
		return nil, nil, ErrCode("listing insances", err)
	}

	log.Debugf("got output from instance list %+v", out)

	list := []map[string]*string{}
	for _, r := range out.Reservations {
		log.Debugf("reserveration: %s", aws.StringValue(r.ReservationId))
		for _, i := range r.Instances {
			log.Debugf("instance: %s", aws.StringValue(i.InstanceId))

			if aws.StringValue(i.InstanceLifecycle) == "spot" {
				continue
			}

			list = append(list, map[string]*string{
				"id": i.InstanceId,
			})
		}
	}

	return list, out.NextToken, nil
}

// GetInstance gets details about an instance by ID
func (e *Ec2) GetInstance(ctx context.Context, id string) (*ec2.Instance, error) {
	if id == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting details about ec2 instance %s/%s", e.org, id)

	out, err := e.Service.DescribeInstancesWithContext(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
		Filters: []*ec2.Filter{
			notTerminated(),
		},
	})

	if err != nil {
		return nil, ErrCode("getting instance", err)
	}

	log.Debugf("got output for instance %s: %+v", id, out)

	if len(out.Reservations) == 0 || len(out.Reservations[0].Instances) == 0 {
		return nil, apierror.New(apierror.ErrNotFound, "Resource not found", nil)
	}

	if len(out.Reservations) != 1 || len(out.Reservations[0].Instances) != 1 {
		return nil, apierror.New(apierror.ErrBadRequest, "Unexpected resource count", nil)
	}

	return out.Reservations[0].Instances[0], nil
}

// ListInstanceVolumes returns the volumes for an instance
func (e *Ec2) ListInstanceVolumes(ctx context.Context, id string) ([]string, error) {
	if id == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting snapshots for instance %s/%s", e.org, id)

	volumes := []string{}

	input := ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			withInstanceId(id),
		},
		MaxResults: aws.Int64(1000),
	}

	for {
		out, err := e.Service.DescribeVolumesWithContext(ctx, &input)
		if err != nil {
			return nil, ErrCode("describing snapshots for volumes", err)
		}

		log.Debugf("got describe snapshots output %+v", out)

		for _, v := range out.Volumes {
			volumes = append(volumes, aws.StringValue(v.VolumeId))
		}

		if out.NextToken != nil {
			input.NextToken = out.NextToken
			continue
		}

		break
	}

	return volumes, nil
}

// ListInstanceSnapshots returns the snapshots for all volumes for an instance
func (e *Ec2) ListInstanceSnapshots(ctx context.Context, id string) ([]string, error) {
	if id == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting snapshots for instance %s/%s", e.org, id)

	snapshots := []string{}

	input := ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			withInstanceId(id),
		},
		MaxResults: aws.Int64(1000),
	}

	for {
		out, err := e.Service.DescribeSnapshotsWithContext(ctx, &input)
		if err != nil {
			return nil, ErrCode("describing snapshots for volumes of an instance", err)
		}

		log.Debugf("got describe snapshots output %+v", out)

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

func (e *Ec2) GetInstanceVolume(ctx context.Context, id, volid string) (*ec2.Volume, error) {
	if id == "" || volid == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting volume %s for instance %s/%s", volid, e.org, id)

	out, err := e.Service.DescribeVolumesWithContext(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: aws.StringSlice([]string{volid}),
		Filters: []*ec2.Filter{
			withInstanceId(id),
		},
	})

	if err != nil {
		return nil, ErrCode("describing volume", err)
	}

	log.Debugf("got output describing volumes %+v", out)

	if len(out.Volumes) != 1 {
		return nil, apierror.New(apierror.ErrBadRequest, "unexpected count", nil)
	}

	return out.Volumes[0], nil
}
