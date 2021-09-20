package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListVPCs(ctx context.Context) ([]map[string]string, error) {

	out, err := e.Service.DescribeVpcsWithContext(ctx, &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			isAvailable(),
		},
	})
	if err != nil {
		return nil, ErrCode("describing vpcs", err)
	}

	log.Debugf("got output describing VPCs: %+v", out)

	vpcs := make([]map[string]string, len(out.Vpcs))
	for i, v := range out.Vpcs {
		vpcs[i] = map[string]string{"id": aws.StringValue(v.VpcId)}
	}

	return vpcs, nil
}
