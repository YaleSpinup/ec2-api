package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListImages(ctx context.Context, org, name string) ([]map[string]*string, error) {
	log.Infof("listing ec2 images (name: '%s', org: '%s')", name, org)

	filters := []*ec2.Filter{
		{
			Name:   aws.String("is-public"),
			Values: aws.StringSlice([]string{"false"}),
		},
		{
			Name:   aws.String("state"),
			Values: aws.StringSlice([]string{"available"}),
		},
	}

	if org != "" {
		filters = append(filters, inOrg(org))
	}

	if name != "" {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("name"),
			Values: aws.StringSlice([]string{name}),
		})
	}

	out, err := e.Service.DescribeImagesWithContext(ctx, &ec2.DescribeImagesInput{
		Owners:  aws.StringSlice([]string{"self"}),
		Filters: filters,
	})
	if err != nil {
		return nil, common.ErrCode("listing images", err)
	}

	log.Debugf("returning list of %d images", len(out.Images))

	list := make([]map[string]*string, len(out.Images))
	for j, i := range out.Images {
		list[j] = map[string]*string{
			"id":   i.ImageId,
			"name": i.Name,
		}
	}

	return list, nil
}

func (e *Ec2) GetImage(ctx context.Context, ids ...string) ([]*ec2.Image, error) {
	if len(ids) == 0 {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting details about image ids %+v", ids)

	input := ec2.DescribeImagesInput{
		ImageIds: aws.StringSlice(ids),
	}

	out, err := e.Service.DescribeImagesWithContext(ctx, &input)
	if err != nil {
		return nil, common.ErrCode("getting details for snapshots", err)
	}

	log.Debugf("returning images: %+v", out.Images)

	return out.Images, nil
}

// CreateImage creates a new image and returns the image details
func (e *Ec2) CreateImage(ctx context.Context, input *ec2.RunInstancesInput) (*ec2.Instance, error) {
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
