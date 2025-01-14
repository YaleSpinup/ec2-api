// In ssm/managedinstance.go

package ssm

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

// ManagedInstance represents the response structure for managed instances
type ManagedInstance struct {
	InstanceId   *string `json:"instanceId"`
	Name         *string `json:"name"`
	PingStatus   *string `json:"pingStatus"`
	PlatformType *string `json:"platformType"`
	ResourceType *string `json:"resourceType"`
	ComputerName *string `json:"computerName"`
	IPAddress    *string `json:"ipAddress"`
	AgentVersion *string `json:"agentVersion"`
}

// ListManagedInstances lists hybrid instances from SSM Fleet Manager
func (s *SSM) ListManagedInstances(ctx context.Context, per int64, next *string) ([]*ManagedInstance, *string, error) {
	if per < 1 || per > 50 {
		return nil, nil, apierror.New(apierror.ErrBadRequest, "per page must be between 1 and 50", nil)
	}

	log.Infof("listing managed instances from SSM")

	input := &ssm.DescribeInstanceInformationInput{
		MaxResults: aws.Int64(per),
		NextToken:  next,
		Filters: []*ssm.InstanceInformationStringFilter{
			{
				Key: aws.String("ResourceType"),
				Values: []*string{
					aws.String("ManagedInstance"),
				},
			},
		},
	}

	out, err := s.Service.DescribeInstanceInformationWithContext(ctx, input)
	if err != nil {
		return nil, nil, common.ErrCode("listing managed instances", err)
	}

	log.Debugf("got output from managed instance list: %+v", out)

	instances := make([]*ManagedInstance, 0, len(out.InstanceInformationList))
	for _, info := range out.InstanceInformationList {
		instance := &ManagedInstance{
			InstanceId:   info.InstanceId,
			Name:         info.Name,
			PingStatus:   info.PingStatus,
			PlatformType: info.PlatformType,
			ResourceType: info.ResourceType,
			ComputerName: info.ComputerName,
			IPAddress:    info.IPAddress,
			AgentVersion: info.AgentVersion,
		}
		instances = append(instances, instance)
	}

	return instances, out.NextToken, nil
}

// GetManagedInstance gets details about a specific managed instance
func (s *SSM) GetManagedInstance(ctx context.Context, instanceId string) (*ManagedInstance, error) {
	if instanceId == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "instance id is required", nil)
	}

	log.Infof("getting managed instance details for %s", instanceId)

	input := &ssm.DescribeInstanceInformationInput{
		Filters: []*ssm.InstanceInformationStringFilter{
			{
				Key: aws.String("InstanceIds"),
				Values: []*string{
					aws.String(instanceId),
				},
			},
			{
				Key: aws.String("ResourceType"),
				Values: []*string{
					aws.String("ManagedInstance"),
				},
			},
		},
	}

	out, err := s.Service.DescribeInstanceInformationWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("getting managed instance", err)
	}

	log.Debugf("got output from managed instance get: %+v", out)

	if len(out.InstanceInformationList) == 0 {
		return nil, apierror.New(apierror.ErrNotFound, "managed instance not found", nil)
	}

	if len(out.InstanceInformationList) > 1 {
		return nil, apierror.New(apierror.ErrBadRequest, "multiple instances found", nil)
	}

	info := out.InstanceInformationList[0]
	instance := &ManagedInstance{
		InstanceId:   info.InstanceId,
		Name:         info.Name,
		PingStatus:   info.PingStatus,
		PlatformType: info.PlatformType,
		ResourceType: info.ResourceType,
		ComputerName: info.ComputerName,
		IPAddress:    info.IPAddress,
		AgentVersion: info.AgentVersion,
	}

	return instance, nil
}
