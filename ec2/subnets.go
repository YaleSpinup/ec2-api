package ec2

import (
	"context"
	"strings"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListSubnets(ctx context.Context, vpc string) ([]map[string]string, error) {
	if vpc != "" && !strings.HasPrefix(vpc, "vpc-") {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	filters := []*ec2.Filter{
		isAvailable(),
	}

	if vpc != "" {
		log.Infof("listing subnets for vpc %s", vpc)
		filters = append(filters, inVpc(vpc))
	} else {
		log.Info("listing subnets")
	}

	out, err := e.Service.DescribeSubnetsWithContext(ctx, &ec2.DescribeSubnetsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, ErrCode("failed to list subnets", err)
	}

	log.Debugf("got output describing subnets: %+v", out)

	subnets := make([]map[string]string, len(out.Subnets))
	for i, subnet := range out.Subnets {
		var name string
		for _, t := range subnet.Tags {
			if aws.StringValue(t.Key) == "Name" {
				name = aws.StringValue(t.Value)
				break
			}
		}

		subnets[i] = map[string]string{
			aws.StringValue(subnet.SubnetId): name,
		}
	}

	return subnets, nil
}
