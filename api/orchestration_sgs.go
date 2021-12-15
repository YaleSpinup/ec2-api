package api

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (o *ec2Orchestrator) createSecurityGroup(ctx context.Context, req *Ec2SecurityGroupRequest) (string, error) {
	if req == nil {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	var err error
	var rollBackTasks []rollbackFunc
	defer func() {
		if err != nil {
			log.Errorf("recovering from error: %s, executing %d rollback tasks", err, len(rollBackTasks))
			rollBack(&rollBackTasks)
		}
	}()

	tags := []*ec2.Tag{}
	for _, tag := range req.Tags {
		for k, v := range tag {
			tags = append(tags, &ec2.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
	}

	out, err := o.client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		Description: aws.String(req.Description),
		GroupName:   aws.String(req.GroupName),
		VpcId:       aws.String(req.VpcId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("security-group"),
				Tags:         tags,
			},
		},
	})
	if err != nil {
		return "", err
	}

	err = o.client.WaitUntilSecurityGroupExists(ctx, aws.StringValue(out.GroupId))
	if err != nil {
		return "", err
	}

	rollBackTasks = append(rollBackTasks, func(ctx context.Context) error {
		log.Errorf("rollback: deleting security group: %s", aws.StringValue(out.GroupId))
		return o.client.DeleteSecurityGroup(ctx, aws.StringValue(out.GroupId))
	})

	if len(req.InitRules) > 0 {
		for _, r := range req.InitRules {
			log.Debugf("creating securitygrouprulerequest with %+v", r)

			if r.CidrIp == nil && r.SgId == nil {
				return "", apierror.New(apierror.ErrBadRequest, "cidr_ip or sg_id is required", nil)
			}

			ipPermissions := ipPermissionsFromRequest(r)

			err = o.client.AuthorizeSecurityGroup(ctx, *r.RuleType, aws.StringValue(out.GroupId), ipPermissions)
			if err != nil {
				return "", err
			}
		}
	}

	return aws.StringValue(out.GroupId), nil
}

func (o *ec2Orchestrator) updateSecurityGroup(ctx context.Context, id string, req *Ec2SecurityGroupRuleRequest) error {
	if id == "" || req == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	switch *req.Action {
	case "add":
		if err := o.client.AuthorizeSecurityGroup(ctx, *req.RuleType, id, ipPermissionsFromRequest(req)); err != nil {
			return err
		}
	case "remove":
		if err := o.client.RevokeSecurityGroup(ctx, *req.RuleType, id, ipPermissionsFromRequest(req)); err != nil {
			return err
		}
	default:
		return apierror.New(apierror.ErrBadRequest, "action should be [add|remove]", nil)
	}

	return nil
}

func ipPermissionsFromRequest(r *Ec2SecurityGroupRuleRequest) []*ec2.IpPermission {
	ipPermissions := []*ec2.IpPermission{}

	if r.CidrIp != nil {
		ipPermissions = append(ipPermissions, &ec2.IpPermission{
			IpProtocol: r.IpProtocol,
			FromPort:   r.FromPort,
			ToPort:     r.ToPort,
			IpRanges: []*ec2.IpRange{
				{
					CidrIp:      r.CidrIp,
					Description: r.Description,
				},
			},
		})
	}

	if r.SgId != nil {
		ipPermissions = append(ipPermissions, &ec2.IpPermission{
			IpProtocol: r.IpProtocol,
			FromPort:   r.FromPort,
			ToPort:     r.ToPort,
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupId:     r.SgId,
					Description: r.Description,
				},
			},
		})
	}

	return ipPermissions
}
