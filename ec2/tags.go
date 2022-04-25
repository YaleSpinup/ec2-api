package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) UpdateTags(ctx context.Context, rawTags map[string]string, ids ...string) error {
	if len(ids) == 0 || len(rawTags) == 0 {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	var tags []*ec2.Tag
	for key, val := range rawTags {
		tags = append(tags, &ec2.Tag{Key: aws.String(key), Value: aws.String(val)})
	}

	log.Infof("updating resources: %v with tags %+v", ids, tags)

	input := ec2.CreateTagsInput{
		Resources: aws.StringSlice(ids),
		Tags:      tags,
	}

	if _, err := e.Service.CreateTagsWithContext(ctx, &input); err != nil {
		return common.ErrCode("creating tags", err)
	}

	return nil
}
