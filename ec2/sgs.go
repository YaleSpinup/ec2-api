package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) CreateSecurityGroup(ctx context.Context, input *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("creating security group %s in vpc %s", aws.StringValue(input.GroupName), aws.StringValue(input.VpcId))

	out, err := e.Service.CreateSecurityGroupWithContext(ctx, input)
	if err != nil {
		return nil, ErrCode("failed to create security group", err)
	}

	log.Debugf("got output creating security group %+v", out)

	return out, nil
}

// ListSecurityGroups List all security groups in an aws account
func (e *Ec2) ListSecurityGroups(ctx context.Context, org string) ([]map[string]*string, error) {
	log.Infof("listing ec2 security groups (org: '%s')", org)

	var filters []*ec2.Filter
	if org != "" {
		filters = []*ec2.Filter{inOrg(org)}
	}

	out, err := e.Service.DescribeSecurityGroupsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, ErrCode("listing security groups", err)
	}

	log.Debugf("returning list of %d security groups", len(out.SecurityGroups))

	list := make([]map[string]*string, len(out.SecurityGroups))
	for i, s := range out.SecurityGroups {
		tags := s.Tags
		sgName := aws.StringValue(s.GroupName)

		// Loop through the tags and if Name exist, set the sgName value to it
		for _, t := range tags {
			if aws.StringValue(t.Key) == "Name" {
				sgName = aws.StringValue(t.Value)
				break
			}
		}

		list[i] = map[string]*string{
			aws.StringValue(s.GroupId): &sgName,
		}
	}

	return list, err
}

// GetSecurityGroup Get the given security groups by a list of ids
func (e *Ec2) GetSecurityGroup(ctx context.Context, ids ...string) ([]*ec2.SecurityGroup, error) {
	if len(ids) == 0 {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting details about security group ids %+v", ids)

	out, err := e.Service.DescribeSecurityGroupsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice(ids),
	})
	if err != nil {
		return nil, ErrCode("getting details for security groups", err)
	}

	log.Debugf("returning security groups: %+v", out.SecurityGroups)

	return out.SecurityGroups, err
}

// DeleteSecurityGroup deletes the given security group
func (e *Ec2) DeleteSecurityGroup(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("deleting security group %s", id)

	if _, err := e.Service.DeleteSecurityGroupWithContext(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(id),
	}); err != nil {
		return ErrCode("deleting security group", err)
	}

	return nil
}

func (e *Ec2) WaitUntilSecurityGroupExists(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("waiting for %s to exist", id)

	if err := e.Service.WaitUntilSecurityGroupExistsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}); err != nil {
		return ErrCode("waiting for security group to exist", err)
	}

	return nil
}

func (e *Ec2) AuthorizeSecurityGroup(ctx context.Context, direction, sg string, permissions []*ec2.IpPermission) error {
	if direction == "" || sg == "" || permissions == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("Authorizing security group %s for %s", direction, sg)

	switch direction {
	case "outbound":
		out, err := e.Service.AuthorizeSecurityGroupEgressWithContext(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(sg),
			IpPermissions: permissions,
		})
		if err != nil {
			return ErrCode("failed authorizing egress", err)
		}

		log.Debugf("got output authorizing security group egress: %+v", out)

		if !aws.BoolValue(out.Return) {
			return apierror.New(apierror.ErrBadRequest, "security group authorization rule failed", nil)
		}

	case "inbound":
		out, err := e.Service.AuthorizeSecurityGroupIngressWithContext(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(sg),
			IpPermissions: permissions,
		})
		if err != nil {
			return ErrCode("failed authorizing ingress", err)
		}

		log.Debugf("got output authorizing security group ingress: %+v", out)

		if !aws.BoolValue(out.Return) {
			return apierror.New(apierror.ErrBadRequest, "security group authorization rule failed", nil)
		}
	default:
		return apierror.New(apierror.ErrBadRequest, "direction is required to be [outbound|enbound]", nil)
	}

	return nil
}

func (e *Ec2) RevokeSecurityGroup(ctx context.Context, direction, sg string, permissions []*ec2.IpPermission) error {
	if direction == "" || sg == "" || permissions == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("Revoking security group %s for %s", direction, sg)

	switch direction {
	case "outbound":
		out, err := e.Service.RevokeSecurityGroupEgressWithContext(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       aws.String(sg),
			IpPermissions: permissions,
		})
		if err != nil {
			return ErrCode("failed revoking egress", err)
		}

		log.Debugf("got output authorizing security group egress: %+v", out)

		if !aws.BoolValue(out.Return) {
			return apierror.New(apierror.ErrBadRequest, "security group revoke rule failed", nil)
		}

	case "inbound":
		out, err := e.Service.RevokeSecurityGroupIngressWithContext(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       aws.String(sg),
			IpPermissions: permissions,
		})
		if err != nil {
			return ErrCode("failed revoking egress", err)
		}

		log.Debugf("got output authorizing security group ingress: %+v", out)

		if !aws.BoolValue(out.Return) {
			return apierror.New(apierror.ErrBadRequest, "security group revoke rule failed", nil)
		}
	default:
		return apierror.New(apierror.ErrBadRequest, "direction is required to be [outbound|enbound]", nil)
	}

	return nil
}
