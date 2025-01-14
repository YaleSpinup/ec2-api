package ssm

import (
	"context"
	"regexp"

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

// managedInstanceFilter creates a StringFilter for managed instances
func managedInstanceFilter() *ssm.InstanceInformationStringFilter {
	return &ssm.InstanceInformationStringFilter{
		Key: aws.String("ResourceType"),
		Values: []*string{
			aws.String("ManagedInstance"),
		},
	}
}

// ListManagedInstances lists hybrid instances from SSM Fleet Manager
func (s *SSM) ListManagedInstances(ctx context.Context, per int64, next *string) ([]*ManagedInstance, *string, error) {
	if per < 1 || per > 50 {
		return nil, nil, apierror.New(apierror.ErrBadRequest, "per page must be between 1 and 50", nil)
	}

	log.Info("listing managed instances from SSM")

	input := &ssm.DescribeInstanceInformationInput{
		MaxResults: aws.Int64(per),
		NextToken:  next,
		Filters: []*ssm.InstanceInformationStringFilter{
			managedInstanceFilter(),
		},
	}

	out, err := s.Service.DescribeInstanceInformationWithContext(ctx, input)
	if err != nil {
		return nil, nil, common.ErrCode("listing managed instances", err)
	}

	log.Debugf("got output from managed instance list: %+v", out)

	instances := make([]*ManagedInstance, 0, len(out.InstanceInformationList))
	for _, info := range out.InstanceInformationList {
		instances = append(instances, convertToManagedInstance(info))
	}

	return instances, out.NextToken, nil
}

// GetManagedInstance gets details about a specific managed instance by ID or computer name
func (s *SSM) GetManagedInstance(ctx context.Context, identifier string) (*ManagedInstance, error) {
	if identifier == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "identifier (instance id or computer name) is required", nil)
	}

	log.Infof("getting managed instance details for identifier: %s", identifier)

	// Check if the identifier matches managed instance ID pattern (mi-xxxxxxxxxxxxxxxxx)
	if matched, _ := regexp.MatchString(`^mi-\w{17}$`, identifier); matched {
		// If it's an instance ID, use direct filter
		filters := []*ssm.InstanceInformationStringFilter{
			managedInstanceFilter(),
			{
				Key: aws.String("InstanceId"),
				Values: []*string{
					aws.String(identifier),
				},
			},
		}

		input := &ssm.DescribeInstanceInformationInput{
			Filters: filters,
		}

		out, err := s.Service.DescribeInstanceInformationWithContext(ctx, input)
		if err != nil {
			return nil, common.ErrCode("getting managed instance", err)
		}

		if len(out.InstanceInformationList) == 0 {
			return nil, apierror.New(apierror.ErrNotFound, "managed instance not found", nil)
		}

		if len(out.InstanceInformationList) > 1 {
			return nil, apierror.New(apierror.ErrBadRequest, "multiple instances found", nil)
		}

		return convertToManagedInstance(out.InstanceInformationList[0]), nil
	}

	// If not an instance ID, assume it's a computer name and list all instances
	instances, _, err := s.ListManagedInstances(ctx, 50, nil)
	if err != nil {
		return nil, err
	}

	var matches []*ManagedInstance
	for _, instance := range instances {
		if aws.StringValue(instance.ComputerName) == identifier {
			matches = append(matches, instance)
		}
	}

	if len(matches) == 0 {
		return nil, apierror.New(apierror.ErrNotFound, "managed instance not found", nil)
	}

	if len(matches) > 1 {
		return nil, apierror.New(apierror.ErrBadRequest, "multiple instances found with same computer name", nil)
	}

	return matches[0], nil
}

// Helper function to convert SSM instance info to our ManagedInstance type
func convertToManagedInstance(info *ssm.InstanceInformation) *ManagedInstance {
	return &ManagedInstance{
		InstanceId:   info.InstanceId,
		Name:         info.Name,
		PingStatus:   info.PingStatus,
		PlatformType: info.PlatformType,
		ResourceType: info.ResourceType,
		ComputerName: info.ComputerName,
		IPAddress:    info.IPAddress,
		AgentVersion: info.AgentVersion,
	}
}
