package api

import (
	"context"

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
