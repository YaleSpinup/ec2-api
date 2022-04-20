package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) UpdateTags(ctx context.Context, id string, rawTags map[string]string) error {
	if id == "" || len(rawTags) == 0 {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	resources := e.getResourcesById(id)
	tags := toEc2Tags(rawTags)

	log.Infof("updating resources: %v with tags %+v", resources, tags)

	input := ec2.CreateTagsInput{
		Resources: resources,
		Tags:      tags,
	}

	_, err := e.Service.CreateTagsWithContext(ctx, &input)
	if err != nil {
		return common.ErrCode("creating tags", err)
	}

	return nil
}

func (e *Ec2) getResourcesById(id string) []*string {
	return []*string{aws.String(id)}
}

func toEc2Tags(rawTags map[string]string) (tags []*ec2.Tag) {
	for key, val := range rawTags {
		tags = append(tags, &ec2.Tag{Key: aws.String(key), Value: aws.String(val)})
	}
	return tags
}
