package ec2

import (
	"context"

	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// DescribeAllInstanceTypes paginates through DescribeInstanceTypes and returns all instance type metadata.
func (e *Ec2) DescribeAllInstanceTypes(ctx context.Context) ([]*ec2.InstanceTypeInfo, error) {
	log.Infof("describing all ec2 instance types")

	var result []*ec2.InstanceTypeInfo
	input := &ec2.DescribeInstanceTypesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		out, err := e.Service.DescribeInstanceTypesWithContext(ctx, input)
		if err != nil {
			return nil, common.ErrCode("describing instance types", err)
		}

		result = append(result, out.InstanceTypes...)

		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	log.Infof("described %d instance types", len(result))
	return result, nil
}

// ListAllInstanceTypeOfferings paginates through DescribeInstanceTypeOfferings
// and returns all offerings across all availability zones in the region.
func (e *Ec2) ListAllInstanceTypeOfferings(ctx context.Context) ([]*ec2.InstanceTypeOffering, error) {
	log.Infof("listing all instance type offerings by availability zone")

	var result []*ec2.InstanceTypeOffering
	input := &ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: aws.String("availability-zone"),
		MaxResults:   aws.Int64(1000),
	}

	for {
		out, err := e.Service.DescribeInstanceTypeOfferingsWithContext(ctx, input)
		if err != nil {
			return nil, common.ErrCode("listing instance type offerings", err)
		}

		result = append(result, out.InstanceTypeOfferings...)

		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	log.Infof("listed %d instance type offerings", len(result))
	return result, nil
}
