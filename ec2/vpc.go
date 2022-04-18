package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
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
		return nil, common.ErrCode("describing vpcs", err)
	}

	log.Debugf("got output describing VPCs: %+v", out)

	vpcs := make([]map[string]string, len(out.Vpcs))
	for i, v := range out.Vpcs {
		vpcs[i] = map[string]string{"id": aws.StringValue(v.VpcId)}
	}

	return vpcs, nil
}

func (e *Ec2) GetVPCByID(ctx context.Context, id string) (*ec2.Vpc, error) {
	out, err := e.Service.DescribeVpcsWithContext(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, common.ErrCode("describing vpc", err)
	}
	log.Debugf("got output describing VPC : %+v", out)

	if len(out.Vpcs) == 0 {
		return nil, apierror.New(apierror.ErrNotFound, "vpc not found", nil)
	}
	return out.Vpcs[0], nil
}
