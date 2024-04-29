package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
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
		return nil, common.ErrCode("failed to create instance", err)
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
		return common.ErrCode("failed to delete instance", err)
	}

	log.Debugf("got output deleting instance: %+v", out)

	return nil
}

//$ec2 = AWS::createClient('ec2');
//
//try {
//// Describe the available instance types
//$result = $ec2->describeInstanceTypeOfferings([
//'LocationType' => 'availability-zone',
//'Filters' => [
//[
//'Name' => 'location',
//'Values' => [$availabilityZone],
//],
//],
//]);
//
//$instanceTypes = [];
//
//// Extract the instance types from the response
//foreach ($result->get('InstanceTypeOfferings') as $offering) {
//$instanceTypes[] = $offering['InstanceType'];
//}
//
//return response()->json([
//'instanceTypes' => $instanceTypes,
//]);
//} catch (Aws\Exception\AwsException $e) {
//// Handle AWS errors
//return response()->json([
//'error' => $e->getMessage(),
//], 500);
//} catch (Exception $e) {
//// Handle other exceptions
//return response()->json([
//'error' => $e->getMessage(),
//], 500);
//}

func (e *Ec2) ListInstanceTypeOfferings(ctx context.Context, azs []string, per int64, next *string) ([]*ec2.InstanceTypeOffering, *string, error) {
	log.Infof("listing ec2 instance type offerings")

	var nextToken *string
	if next != nil {
		nextToken = next
	}

	filters := []*ec2.Filter{
		{
			Name:   aws.String("location"),
			Values: aws.StringSlice(azs),
		},
	}

	out, err := e.Service.DescribeInstanceTypeOfferingsWithContext(ctx, &ec2.DescribeInstanceTypeOfferingsInput{
		Filters:      filters,
		LocationType: aws.String("availability-zone"),
		MaxResults:   aws.Int64(per),
		NextToken:    nextToken,
	})

	if err != nil {
		return nil, nil, common.ErrCode("listing instance type offerings", err)
	}

	log.Debugf("gout output from instance type offerings list %+v", out)

	return out.InstanceTypeOfferings, out.NextToken, nil
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
		return nil, nil, common.ErrCode("listing instances", err)
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

			var instanceName *string

			for _, tag := range i.Tags {
				if *tag.Key == "Name" {
					instanceName = tag.Value
					break
				}
			}

			list = append(list, map[string]*string{
				"id":    i.InstanceId,
				"name":  instanceName,
				"state": i.State.Name,
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
		return nil, common.ErrCode("getting instance", err)
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

	log.Infof("getting volumes for instance %s/%s", e.org, id)

	volumes := []string{}

	input := ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.instance-id"),
				Values: aws.StringSlice(
					[]string{id},
				),
			},
		},
		MaxResults: aws.Int64(1000),
	}

	for {
		out, err := e.Service.DescribeVolumesWithContext(ctx, &input)
		if err != nil {
			return nil, common.ErrCode("describing volumes for instance", err)
		}

		log.Debugf("got describe volumes output %+v", out)

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
			return nil, common.ErrCode("describing snapshots for volumes of an instance", err)
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
		return nil, common.ErrCode("describing volume", err)
	}

	log.Debugf("got output describing volumes %+v", out)

	if len(out.Volumes) != 1 {
		return nil, apierror.New(apierror.ErrBadRequest, "unexpected count", nil)
	}

	return out.Volumes[0], nil
}

func (e *Ec2) StartInstance(ctx context.Context, ids ...string) error {
	if len(ids) == 0 {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("starting instance %s/%v", e.org, ids)
	inp := &ec2.StartInstancesInput{
		InstanceIds: aws.StringSlice(ids),
	}
	if _, err := e.Service.StartInstancesWithContext(ctx, inp); err != nil {
		return common.ErrCode("starting instance", err)
	}
	return nil
}

func (e *Ec2) StopInstance(ctx context.Context, force bool, ids ...string) error {
	if len(ids) == 0 {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("stopping instance %s/%v", e.org, ids)
	inp := &ec2.StopInstancesInput{
		Force:       aws.Bool(force),
		InstanceIds: aws.StringSlice(ids),
	}
	if _, err := e.Service.StopInstancesWithContext(ctx, inp); err != nil {
		return common.ErrCode("stopping instance", err)
	}
	return nil
}

func (e *Ec2) RebootInstance(ctx context.Context, ids ...string) error {
	if len(ids) == 0 {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("rebooting instance %s/%v", e.org, ids)
	inp := &ec2.StartInstancesInput{
		InstanceIds: aws.StringSlice(ids),
	}
	if _, err := e.Service.StartInstancesWithContext(ctx, inp); err != nil {
		return common.ErrCode("rebooting instance", err)
	}
	return nil
}
