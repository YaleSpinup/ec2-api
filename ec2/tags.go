package ec2

import (
	"context"
	"strings"

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

	instanceIDs := []*string{}
	for _, id := range ids {
		if strings.HasPrefix(id, "i-") {
			instanceIDs = append(instanceIDs, aws.String(id))
		}
	}

	describeVolumesInput := ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: instanceIDs,
			},
		},
		MaxResults: aws.Int64(1000),
	}

	for {
		out, err := e.Service.DescribeVolumesWithContext(ctx, &describeVolumesInput)
		if err != nil {
			return common.ErrCode("describing volumes for instance", err)
		}

		log.Debugf("got describe volumes output %+v", out)

		for _, v := range out.Volumes {
			ids = append(ids, aws.StringValue(v.VolumeId))
		}

		if out.NextToken != nil {
			describeVolumesInput.NextToken = out.NextToken
			continue
		}

		break
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
