package api

import (
	"context"
	"fmt"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (o *ec2Orchestrator) createImage(ctx context.Context, req *Ec2ImageCreateRequest) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Debugf("got request to create image: %s", awsutil.Prettify(req))

	instance, err := o.ec2Client.GetInstance(ctx, *req.InstanceId)
	if err != nil {
		return "", apierror.New(apierror.ErrBadRequest, err.Error(), nil)
	}

	input := &ec2.CreateImageInput{
		InstanceId:  req.InstanceId,
		Name:        req.Name,
		Description: req.Description,
		NoReboot:    aws.Bool(!aws.BoolValue(req.ForceReboot)),
	}
	if aws.BoolValue(req.CopyTags) {
		input.TagSpecifications = []*ec2.TagSpecification{{
			ResourceType: aws.String("image"),
			Tags:         instance.Tags,
		}}
	}

	imageId, err := o.ec2Client.CreateImage(ctx, input)
	if err != nil {
		return "", err
	}

	return imageId, nil
}

func (o *ec2Orchestrator) deleteImage(ctx context.Context, id string) error {

	log.Debugf("got request to delete image %s", id)

	input := &ec2.DeregisterImageInput{
		ImageId: aws.String(id),
	}
	if err := o.ec2Client.DeleteImage(ctx, input); err != nil {
		return err
	}

	Snapshotinput := &ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{{Name: aws.String("description"), Values: aws.StringSlice([]string{fmt.Sprintf("*for %s from vol*", id)})}},
	}

	out, err := o.ec2Client.DescribeSnapshots(ctx, Snapshotinput)
	if err != nil {
		return err
	}
	for _, s := range out {
		input := &ec2.DeleteSnapshotInput{
			SnapshotId: (s.SnapshotId),
		}
		if err := o.ec2Client.DeleteSnapshot(ctx, input); err != nil {
			return err
			//TODO: need to review the return error condition in the loop
		}

	}
	return nil
}
