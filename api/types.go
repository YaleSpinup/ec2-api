package api

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

// timeFormat returns the standardized time format used by most of the returned objects
func timeFormat(t *time.Time) string {
	if t == nil {
		return ""
	}

	// TODO convert to standard with timezone, requires changes in reaper, indexer, ui, etc
	// return t.UTC().UTC().Format(time.RFC3339)
	return t.UTC().Format("2006/01/02 15:04:05")
}

// tzTimeFormat returns the time format with a TZ used in a few places
// TODO get rid of these places
func tzTimeFormat(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.UTC().Format("2006-01-02 15:04:05 MST")
}

type Ec2InstanceCreateRequest struct {
	Type            *string          `json:"type"`
	Image           *string          `json:"image"`
	Subnet          *string          `json:"subnet"`
	Sgs             []*string        `json:"sgs"`
	CpuCredits      *string          `json:"cpu_credits"`
	InstanceProfile *string          `json:"instanceprofile"`
	Key             *string          `json:"key"`
	Userdata64      *string          `json:"userdata64"`
	BlockDevices    []Ec2BlockDevice `json:"block_devices"`
}

type Ec2BlockDevice struct {
	DeviceName *string       `json:"device_name"`
	Ebs        *Ec2EbsVolume `json:"ebs"`
}

type Ec2VolumeCreateRequest struct {
	Type       *string `json:"type"`
	Size       *int64  `json:"size"`
	Iops       *int64  `json:"iops"`
	AZ         *string `json:"az"`
	SnapshotId *string `json:"snapshot_id"`
	KmsKeyId   *string `json:"kms_key_id"`
	Encrypted  *bool   `json:"encrypted"`
}

type Ec2EbsVolume struct {
	Encrypted  *bool   `json:"encrypted"`
	VolumeSize *int64  `json:"volume_size"`
	VolumeType *string `json:"volume_type"`
}

type Volume struct {
	AttachTime          string `json:"attach_time"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
	Status              string `json:"status"`
	DeviceName          string `json:"device_name"`
}

type Ec2InstanceResponse struct {
	Az        string              `json:"az"`
	CreatedAt string              `json:"created_at"`
	CreatedBy string              `json:"created_by"`
	ID        string              `json:"id"`
	Image     string              `json:"image"`
	Ip        string              `json:"ip"`
	Key       string              `json:"key"`
	Name      string              `json:"name"`
	Platform  string              `json:"platform"`
	Sgs       []map[string]string `json:"sgs"`
	State     string              `json:"state"`
	Subnet    string              `json:"subnet"`
	Tags      []map[string]string `json:"tags"`
	Type      string              `json:"type"`
	Volumes   map[string]*Volume  `json:"volumes"`
}

func toEc2InstanceResponse(instance *ec2.Instance) *Ec2InstanceResponse {
	if instance == nil {
		log.Warn("returning nil response for nil instance")
		return nil
	}

	log.Debugf("mapping ec2 instance %s", awsutil.Prettify(instance))

	tags := make(map[string]string, len(instance.Tags))
	for _, t := range instance.Tags {
		tags[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	name, ok := tags["Name"]
	if !ok {
		log.Warnf("instance %s doesn't have a Name tag", aws.StringValue(instance.InstanceId))
	}

	platform := aws.StringValue(instance.Platform)
	if platform == "" {
		platform = "linux"
	}

	createdBy, ok := tags["yale:created_by"]
	if !ok {
		createdBy = tags["CreatedBy"]
	}

	tagsList := make([]map[string]string, 0, len(instance.Tags))
	for _, t := range instance.Tags {
		tagsList = append(tagsList, map[string]string{
			aws.StringValue(t.Key): aws.StringValue(t.Value),
		})
	}

	sgs := make([]map[string]string, 0, len(instance.SecurityGroups))
	for _, sg := range instance.SecurityGroups {
		log.Debugf("mapping security group %+v", sg)
		sgs = append(sgs, map[string]string{
			aws.StringValue(sg.GroupId): aws.StringValue(sg.GroupName),
		})
	}

	volumes := make(map[string]*Volume)
	for _, v := range instance.BlockDeviceMappings {
		volumes[aws.StringValue(v.Ebs.VolumeId)] = &Volume{
			AttachTime:          tzTimeFormat(v.Ebs.AttachTime),
			DeleteOnTermination: aws.BoolValue(v.Ebs.DeleteOnTermination),
			Status:              aws.StringValue(v.Ebs.Status),
			DeviceName:          aws.StringValue(v.DeviceName),
		}
	}

	var az string
	if instance.Placement != nil {
		az = aws.StringValue(instance.Placement.AvailabilityZone)
	}

	var state string
	if instance.State != nil {
		state = aws.StringValue(instance.State.Name)
	}

	// TODO pull createdat from tags
	response := Ec2InstanceResponse{
		Az:        az,
		CreatedAt: timeFormat(instance.LaunchTime),
		CreatedBy: createdBy,
		ID:        aws.StringValue(instance.InstanceId),
		Image:     aws.StringValue(instance.ImageId),
		Ip:        aws.StringValue(instance.PrivateIpAddress),
		Key:       aws.StringValue(instance.KeyName),
		Name:      name,
		Platform:  platform,
		Sgs:       sgs,
		State:     state,
		Subnet:    aws.StringValue(instance.SubnetId),
		Tags:      tagsList,
		Type:      aws.StringValue(instance.InstanceType),
		Volumes:   volumes,
	}

	return &response
}

type Ec2VolumeAttachment struct {
	AttachTime          string `json:"attach_time"`
	Device              string `json:"device"`
	InstanceID          string `json:"instance_id"`
	State               string `json:"state"`
	VolumeID            string `json:"volume_id"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
}

type Ec2VolumeResponse struct {
	CreatedAt   string                 `json:"created_at"`
	ID          string                 `json:"id"`
	Size        int64                  `json:"size"`
	Iops        int64                  `json:"iops"`
	Encrypted   bool                   `json:"encrypted"`
	State       string                 `json:"state"`
	Tags        []map[string]string    `json:"tags"`
	Type        string                 `json:"type"`
	Attachments []*Ec2VolumeAttachment `json:"attachments"`
}

func toEc2VolumeResponse(volume *ec2.Volume) *Ec2VolumeResponse {
	if volume == nil {
		log.Warn("returning nil response for nil volume")
		return nil
	}

	log.Debugf("mapping ec2 instance volume %s", awsutil.Prettify(volume))

	tagsList := make([]map[string]string, 0, len(volume.Tags))
	for _, t := range volume.Tags {
		tagsList = append(tagsList, map[string]string{
			aws.StringValue(t.Key): aws.StringValue(t.Value),
		})
	}

	attachments := []*Ec2VolumeAttachment{}
	for _, a := range volume.Attachments {
		attachments = append(attachments, &Ec2VolumeAttachment{
			AttachTime:          tzTimeFormat(a.AttachTime),
			Device:              aws.StringValue(a.Device),
			InstanceID:          aws.StringValue(a.InstanceId),
			State:               aws.StringValue(a.State),
			VolumeID:            aws.StringValue(a.VolumeId),
			DeleteOnTermination: aws.BoolValue(a.DeleteOnTermination),
		})
	}

	response := Ec2VolumeResponse{
		CreatedAt:   timeFormat(volume.CreateTime),
		ID:          aws.StringValue(volume.VolumeId),
		Size:        aws.Int64Value(volume.Size),
		Iops:        aws.Int64Value(volume.Iops),
		Encrypted:   aws.BoolValue(volume.Encrypted),
		State:       aws.StringValue(volume.State),
		Tags:        tagsList,
		Type:        aws.StringValue(volume.VolumeType),
		Attachments: attachments,
	}
	return &response
}

type Ec2VolumeModification struct {
	EndTime                    string `json:"end_time"`
	ModificationState          string `json:"modification_state"`
	OriginalIops               int64  `json:"original_iops"`
	OriginalMultiAttachEnabled bool   `json:"original_multi_attach_enabled"`
	OriginalSize               int64  `json:"original_size"`
	OriginalThroughput         int64  `json:"original_throughput"`
	OriginalVolumeType         string `json:"original_volume_type"`
	Progress                   int64  `json:"progress"`
	StartTime                  string `json:"start_time"`
	StatusMessage              string `json:"status_message"`
	TargetIops                 int64  `json:"target_iops"`
	TargetMultiAttachEnabled   bool   `json:"target_multi_attach_enabled"`
	TargetSize                 int64  `json:"target_size"`
	TargetThroughput           int64  `json:"target_throughput"`
	TargetVolumeType           string `json:"target_volume_type"`
	VolumeId                   string `json:"volume_id"`
}

func toEc2VolumeModificationsResponse(modifications []*ec2.VolumeModification) []Ec2VolumeModification {
	if modifications == nil {
		return nil
	}

	log.Debugf("mapping ec2 volume modifications %s", awsutil.Prettify(modifications))

	response := []Ec2VolumeModification{}

	for _, m := range modifications {
		response = append(response, Ec2VolumeModification{
			EndTime:                    tzTimeFormat(m.EndTime),
			ModificationState:          aws.StringValue(m.ModificationState),
			OriginalIops:               aws.Int64Value(m.OriginalIops),
			OriginalMultiAttachEnabled: aws.BoolValue(m.OriginalMultiAttachEnabled),
			OriginalSize:               aws.Int64Value(m.OriginalSize),
			OriginalThroughput:         aws.Int64Value(m.OriginalThroughput),
			OriginalVolumeType:         aws.StringValue(m.OriginalVolumeType),
			Progress:                   aws.Int64Value(m.Progress),
			StartTime:                  tzTimeFormat(m.StartTime),
			StatusMessage:              aws.StringValue(m.StatusMessage),
			TargetIops:                 aws.Int64Value(m.TargetIops),
			TargetMultiAttachEnabled:   aws.BoolValue(m.TargetMultiAttachEnabled),
			TargetSize:                 aws.Int64Value(m.TargetSize),
			TargetThroughput:           aws.Int64Value(m.TargetThroughput),
			TargetVolumeType:           aws.StringValue(m.TargetVolumeType),
			VolumeId:                   aws.StringValue(m.VolumeId),
		})
	}

	return response
}

type Ec2SnapshotResponse struct {
	CreatedAt   string              `json:"created_at"`
	Description string              `json:"description"`
	Encrypted   bool                `json:"encrypted"`
	ID          string              `json:"id"`
	Owner       string              `json:"owner"`
	Progress    string              `json:"progress"`
	Size        int64               `json:"size"`
	State       string              `json:"state"`
	Tags        []map[string]string `json:"tags"`
	VolumeID    string              `json:"volume_id"`
}

func toEC2SnapshotResponse(snapshot *ec2.Snapshot) *Ec2SnapshotResponse {
	if snapshot == nil {
		log.Warn("returning nil response for nil snapshot")
		return nil
	}

	tags := make(map[string]string, len(snapshot.Tags))
	for _, t := range snapshot.Tags {
		tags[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	tagsList := make([]map[string]string, 0, len(tags))
	for k, v := range tags {
		tagsList = append(tagsList, map[string]string{k: v})
	}

	return &Ec2SnapshotResponse{
		CreatedAt:   timeFormat(snapshot.StartTime),
		Description: aws.StringValue(snapshot.Description),
		Encrypted:   aws.BoolValue(snapshot.Encrypted),
		ID:          aws.StringValue(snapshot.SnapshotId),
		Owner:       aws.StringValue(snapshot.OwnerId),
		Progress:    aws.StringValue(snapshot.Progress),
		Size:        aws.Int64Value(snapshot.VolumeSize),
		State:       aws.StringValue(snapshot.State),
		Tags:        tagsList,
		VolumeID:    aws.StringValue(snapshot.VolumeId),
	}
}

type Ec2ImageResponse struct {
	Architecture   string              `json:"architecture"`
	CreatedAt      string              `json:"created_at"`
	CreatedBy      string              `json:"created_by"`
	Description    string              `json:"description"`
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Public         bool                `json:"public"`
	RootDeviceName string              `json:"root_device_name"`
	RootDeviceType string              `json:"root_device_type"`
	State          string              `json:"state"`
	Tags           []map[string]string `json:"tags"`
	Type           string              `json:"type"`
	Volumes        Ec2ImageVolumeMap   `json:"volumes"`
}

type Ec2ImageVolumeMap map[string]*Ec2ImageVolumeResponse
type Ec2ImageVolumeResponse struct {
	DeleteOnTermination bool   `json:"delete_on_termination"`
	SnapshotId          string `json:"snapshot_id"`
	VolumeSize          int64  `json:"volume_size"`
	VolumeType          string `json:"volume_type"`
	Encrypted           bool   `json:"encrypted"`
}

func toEc2ImageResponse(image *ec2.Image) *Ec2ImageResponse {
	tags := make(map[string]string, len(image.Tags))
	for _, t := range image.Tags {
		tags[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// TODO check if we should use tags instead
	// name, ok := tags["Name"]
	// if !ok {
	// 	log.Warnf("image %s doesn't have a Name tag", aws.StringValue(image.ImageId))
	// }

	createdBy, ok := tags["yale:created_by"]
	if !ok {
		createdBy = tags["CreatedBy"]
	}

	tagsList := make([]map[string]string, 0, len(image.Tags))
	for _, t := range image.Tags {
		tagsList = append(tagsList, map[string]string{
			aws.StringValue(t.Key): aws.StringValue(t.Value),
		})
	}

	volumes := make(map[string]*Ec2ImageVolumeResponse)
	for _, v := range image.BlockDeviceMappings {
		if v.Ebs == nil {
			volumes[aws.StringValue(v.DeviceName)] = nil
			continue
		}

		volumes[aws.StringValue(v.DeviceName)] = &Ec2ImageVolumeResponse{
			DeleteOnTermination: aws.BoolValue(v.Ebs.DeleteOnTermination),
			SnapshotId:          aws.StringValue(v.Ebs.SnapshotId),
			VolumeSize:          aws.Int64Value(v.Ebs.VolumeSize),
			VolumeType:          aws.StringValue(v.Ebs.VolumeType),
			Encrypted:           aws.BoolValue(v.Ebs.Encrypted),
		}
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05.000Z", aws.StringValue(image.CreationDate))
	if err != nil {
		log.Errorf("failed to parse created at date %s", aws.StringValue(image.CreationDate))
	}

	return &Ec2ImageResponse{
		Architecture:   aws.StringValue(image.Architecture),
		CreatedAt:      timeFormat(&createdAt),
		CreatedBy:      createdBy,
		Description:    aws.StringValue(image.Description),
		ID:             aws.StringValue(image.ImageId),
		Name:           aws.StringValue(image.Name),
		Public:         aws.BoolValue(image.Public),
		RootDeviceName: aws.StringValue(image.RootDeviceName),
		RootDeviceType: aws.StringValue(image.RootDeviceType),
		State:          aws.StringValue(image.State),
		Tags:           tagsList,
		Type:           aws.StringValue(image.ImageType),
		Volumes:        volumes,
	}
}

// MarshalJSON for Ec2ImageVolumeMap to return an empty object (`{}`) for nil values (instead of `null`)
func (e *Ec2ImageVolumeMap) MarshalJSON() ([]byte, error) {
	output := map[string]interface{}{}
	for k, v := range *e {
		if v == nil {
			output[k] = struct{}{}
			continue
		}
		output[k] = v
	}

	return json.Marshal(&output)
}

type Ec2SecurityGroupRequest struct {
	Description string                         `json:"description"`
	GroupName   string                         `json:"group_name"`
	InitRules   []*Ec2SecurityGroupRuleRequest `json:"init_rules"`
	Tags        []map[string]string            `json:"tags"`
	VpcId       string                         `json:"vpc_id"`
}

type Ec2SecurityGroupRuleRequest struct {
	RuleType    *string `json:"rule_type"`   // Direction of traffic: [inbound|outbound]
	Action      *string `json:"action"`      // Adding or removing the rule: [add|remove]
	CidrIp      *string `json:"cidr_ip"`     // IPv4 CIDR address range to allow traffic to/from
	SgId        *string `json:"sg_id"`       // Security group to allow traffic to/from
	IpProtocol  *string `json:"ip_protocol"` // IP Protocol name [tcp|udp|icmp|-1]
	FromPort    *int64  `json:"from_port"`   // The starting port (not required if Protocol -1)
	ToPort      *int64  `json:"to_port"`     // The ending port (not required if Protocol -1)
	Description *string `json:"description"` // Optional description for this rule
}

type Ec2SecurityGroupUserIdGroupPair struct {
	Description          string `json:"description,omitempty"`
	GroupId              string `json:"group_id,omitempty"`
	GroupName            string `json:"group_name,omitempty"`
	PeeringStatus        string `json:"peering_status,omitempty"`
	UserId               string `json:"user_id,omitempty"`
	VpcId                string `json:"vpc_id,omitempty"`
	VpcPeeringConnection string `json:"vpc_peering_connection,omitempty"`
}

type Ec2SecurityGroupPrefixListId struct {
	Description  string `json:"description,omitempty"`
	PrefixListId string `json:"prefix_list_id,omitempty"`
}

type Ec2SecurityGroupIpv6Range struct {
	CidrIpv6    string `json:"cidr_ipv_6,omitempty"`
	Description string `json:"description,omitempty"`
}

type Ec2SecurityGroupIpRange struct {
	CidrIp      string `json:"cidr_ip,omitempty"`
	Description string `json:"description,omitempty"`
}

type Ec2SecurityGroupIpPermission struct {
	FromPort         int64                              `json:"from_port,omitempty"`
	IpProtocol       string                             `json:"ip_protocol,omitempty"`
	IpRanges         []*Ec2SecurityGroupIpRange         `json:"ip_ranges"`
	Ipv6Ranges       []*Ec2SecurityGroupIpv6Range       `json:"ipv_6_ranges"`
	PrefixListIds    []*Ec2SecurityGroupPrefixListId    `json:"prefix_list_ids"`
	ToPort           int64                              `json:"to_port,omitempty"`
	UserIdGroupPairs []*Ec2SecurityGroupUserIdGroupPair `json:"user_id_group_pairs"`
}

type Ec2SecurityGroupResponse struct {
	Description   string                          `json:"description"`
	GroupId       string                          `json:"group_id"`
	GroupName     string                          `json:"group_name"`
	IncomingRules []*Ec2SecurityGroupIpPermission `json:"incoming_rules"`
	OutgoingRules []*Ec2SecurityGroupIpPermission `json:"outgoing_rules"`
	Tags          []map[string]string             `json:"tags"`
	VpcId         string                          `json:"vpc_id"`
}

// toEc2SecurityGroupIpPermissions Convert a list of ipPermissions to the proper json response format
func toEc2SecurityGroupIpPermissions(ipPermissions []*ec2.IpPermission) []*Ec2SecurityGroupIpPermission {
	outIpPermissions := []*Ec2SecurityGroupIpPermission{}
	for _, outIpPermission := range ipPermissions {
		ipRanges := []*Ec2SecurityGroupIpRange{}
		for _, ipRange := range outIpPermission.IpRanges {
			ipRanges = append(ipRanges, &Ec2SecurityGroupIpRange{
				CidrIp:      aws.StringValue(ipRange.CidrIp),
				Description: aws.StringValue(ipRange.Description),
			})
		}

		ipv6Ranges := []*Ec2SecurityGroupIpv6Range{}
		for _, ipv6Range := range outIpPermission.Ipv6Ranges {
			ipv6Ranges = append(ipv6Ranges, &Ec2SecurityGroupIpv6Range{
				CidrIpv6:    aws.StringValue(ipv6Range.CidrIpv6),
				Description: aws.StringValue(ipv6Range.Description),
			})
		}

		prefixListIds := []*Ec2SecurityGroupPrefixListId{}
		for _, prefixListId := range outIpPermission.PrefixListIds {
			prefixListIds = append(prefixListIds, &Ec2SecurityGroupPrefixListId{
				Description:  aws.StringValue(prefixListId.Description),
				PrefixListId: aws.StringValue(prefixListId.PrefixListId),
			})
		}

		userIdGroupPairs := []*Ec2SecurityGroupUserIdGroupPair{}
		for _, userIdGroupPair := range outIpPermission.UserIdGroupPairs {
			userIdGroupPairs = append(userIdGroupPairs, &Ec2SecurityGroupUserIdGroupPair{
				Description:          aws.StringValue(userIdGroupPair.Description),
				GroupId:              aws.StringValue(userIdGroupPair.GroupId),
				GroupName:            aws.StringValue(userIdGroupPair.GroupName),
				PeeringStatus:        aws.StringValue(userIdGroupPair.PeeringStatus),
				UserId:               aws.StringValue(userIdGroupPair.UserId),
				VpcId:                aws.StringValue(userIdGroupPair.VpcId),
				VpcPeeringConnection: aws.StringValue(userIdGroupPair.VpcPeeringConnectionId),
			})
		}

		outIpPermissions = append(outIpPermissions, &Ec2SecurityGroupIpPermission{
			FromPort:         aws.Int64Value(outIpPermission.FromPort),
			IpProtocol:       aws.StringValue(outIpPermission.IpProtocol),
			IpRanges:         ipRanges,
			Ipv6Ranges:       ipv6Ranges,
			PrefixListIds:    prefixListIds,
			ToPort:           aws.Int64Value(outIpPermission.ToPort),
			UserIdGroupPairs: userIdGroupPairs,
		})
	}

	return outIpPermissions
}

// toEc2SecurityGroupResponse Convert a security group to the proper json response format
func toEc2SecurityGroupResponse(sg *ec2.SecurityGroup) *Ec2SecurityGroupResponse {
	tagsList := make([]map[string]string, 0, len(sg.Tags))
	for _, t := range sg.Tags {
		tagsList = append(tagsList, map[string]string{
			aws.StringValue(t.Key): aws.StringValue(t.Value),
		})
	}

	return &Ec2SecurityGroupResponse{
		Description:   aws.StringValue(sg.Description),
		IncomingRules: toEc2SecurityGroupIpPermissions(sg.IpPermissions),
		OutgoingRules: toEc2SecurityGroupIpPermissions(sg.IpPermissionsEgress),
		GroupId:       aws.StringValue(sg.GroupId),
		GroupName:     aws.StringValue(sg.GroupName),
		Tags:          tagsList,
		VpcId:         aws.StringValue(sg.VpcId),
	}
}

type Ec2VpcResponse struct {
	Id              string              `json:"id"`
	CIDRBlock       string              `json:"cidr_block"`
	DHCPOptionsId   string              `json:"dhcp_options_id"`
	State           string              `json:"state"`
	InstanceTenancy string              `json:"instance_tenancy"`
	IsDefault       bool                `json:"is_default"`
	Tags            []map[string]string `json:"tags"`
}

func toEc2VpcResponse(vpc *ec2.Vpc) *Ec2VpcResponse {
	return &Ec2VpcResponse{
		Id:              aws.StringValue(vpc.VpcId),
		CIDRBlock:       aws.StringValue(vpc.CidrBlock),
		DHCPOptionsId:   aws.StringValue(vpc.DhcpOptionsId),
		State:           aws.StringValue(vpc.State),
		InstanceTenancy: aws.StringValue(vpc.InstanceTenancy),
		IsDefault:       aws.BoolValue(vpc.IsDefault),
		Tags:            tagsList(vpc.Tags),
	}
}

func tagsList(tags []*ec2.Tag) (res []map[string]string) {
	for _, t := range tags {
		res = append(res, map[string]string{
			aws.StringValue(t.Key): aws.StringValue(t.Value),
		})
	}
	return res
}

type CloudWatchOutputConfig struct {
	CloudWatchLogGroupName  string `json:"cloud_watch_log_group_name"`
	CloudWatchOutputEnabled bool   `json:"cloud_watch_output_enabled"`
}

type SSMGetCommandInvocationOutput struct {
	CommandId              string                 `json:"command_id"`
	InstanceId             string                 `json:"instance_id"`
	Comment                string                 `json:"comment"`
	DocumentName           string                 `json:"document_name"`
	DocumentVersion        string                 `json:"document_version"`
	PluginName             string                 `json:"plugin_name"`
	ResponseCode           int                    `json:"response_code"`
	ExecutionStartDateTime string                 `json:"execution_start_date_time"`
	ExecutionElapsedTime   string                 `json:"execution_elapsed_time"`
	ExecutionEndDateTime   string                 `json:"execution_end_date_time"`
	Status                 string                 `json:"status"`
	StatusDetails          string                 `json:"status_details"`
	StandardOutputContent  string                 `json:"standard_output_content"`
	StandardOutputURL      string                 `json:"standard_output_url"`
	StandardErrorContent   string                 `json:"standard_error_content"`
	StandardErrorURL       string                 `json:"standard_error_url"`
	CloudWatchOutputConfig CloudWatchOutputConfig `json:"cloud_watch_output_config"`
}

func toSSMGetCommandInvocationOutput(rawOut *ssm.GetCommandInvocationOutput) *SSMGetCommandInvocationOutput {
	return &SSMGetCommandInvocationOutput{
		CommandId:              aws.StringValue(rawOut.CommandId),
		InstanceId:             aws.StringValue(rawOut.InstanceId),
		Comment:                aws.StringValue(rawOut.Comment),
		DocumentName:           aws.StringValue(rawOut.DocumentName),
		DocumentVersion:        aws.StringValue(rawOut.DocumentVersion),
		PluginName:             aws.StringValue(rawOut.PluginName),
		ResponseCode:           int(aws.Int64Value(rawOut.ResponseCode)),
		ExecutionStartDateTime: aws.StringValue(rawOut.ExecutionStartDateTime),
		ExecutionElapsedTime:   aws.StringValue(rawOut.ExecutionElapsedTime),
		ExecutionEndDateTime:   aws.StringValue(rawOut.ExecutionEndDateTime),
		Status:                 aws.StringValue(rawOut.Status),
		StatusDetails:          aws.StringValue(rawOut.StatusDetails),
		StandardOutputContent:  aws.StringValue(rawOut.StandardOutputContent),
		StandardOutputURL:      aws.StringValue(rawOut.StandardOutputUrl),
		StandardErrorContent:   aws.StringValue(rawOut.StandardErrorContent),
		StandardErrorURL:       aws.StringValue(rawOut.StandardErrorUrl),
		CloudWatchOutputConfig: CloudWatchOutputConfig{
			CloudWatchLogGroupName:  aws.StringValue(rawOut.CloudWatchOutputConfig.CloudWatchLogGroupName),
			CloudWatchOutputEnabled: aws.BoolValue(rawOut.CloudWatchOutputConfig.CloudWatchOutputEnabled),
		},
	}
}

type Ec2ImageUpdateRequest struct {
	Tags map[string]string
}

type Ec2InstancesAttributeUpdateRequest struct {
	InstanceType map[string]string `json:"instance_type,omitempty"`
}
type AssociationDescription struct {
	Name                        string              `json:"name"`
	InstanceId                  string              `json:"instance_id"`
	AssociationVersion          string              `json:"association_version"`
	Date                        string              `json:"date"`
	LastUpdateAssociationDate   string              `json:"last_update_association_date"`
	Status                      AssociationStatus   `json:"status"`
	Overview                    AssociationOverview `json:"overview"`
	DocumentVersion             string              `json:"document_version"`
	AssociationId               string              `json:"association_id"`
	Targets                     []AssociationTarget `json:"targets"`
	LastExecutionDate           string              `json:"last_execution_date"`
	LastSuccessfulExecutionDate string              `json:"last_successful_execution_date"`
	ApplyOnlyAtCronInterval     bool                `json:"apply_only_at_cron_interval"`
}
type AssociationStatus struct {
	Date    string `json:"date"`
	Name    string `json:"name"`
	Message string `json:"message"`
}
type AssociationOverview struct {
	Status         string `json:"status"`
	DetailedStatus string `json:"detailed_status"`
}

type AssociationTarget struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func toSSMAssociationDescription(rawDesc *ssm.DescribeAssociationOutput) *AssociationDescription {
	const dateLayout = "2006-01-02 15:04:05 +0000"

	return &AssociationDescription{
		Name:                      aws.StringValue(rawDesc.AssociationDescription.Name),
		InstanceId:                aws.StringValue(rawDesc.AssociationDescription.InstanceId),
		AssociationVersion:        aws.StringValue(rawDesc.AssociationDescription.AssociationVersion),
		Date:                      rawDesc.AssociationDescription.Date.Format(dateLayout),
		LastUpdateAssociationDate: rawDesc.AssociationDescription.LastUpdateAssociationDate.Format(dateLayout),
		Status: AssociationStatus{
			Date:    rawDesc.AssociationDescription.Status.Date.Format(dateLayout),
			Name:    aws.StringValue(rawDesc.AssociationDescription.Status.Name),
			Message: aws.StringValue(rawDesc.AssociationDescription.Status.Message),
		},
		Overview: AssociationOverview{
			Status:         aws.StringValue(rawDesc.AssociationDescription.Overview.Status),
			DetailedStatus: aws.StringValue(rawDesc.AssociationDescription.Overview.DetailedStatus),
		},
		DocumentVersion:             aws.StringValue(rawDesc.AssociationDescription.DocumentVersion),
		AssociationId:               aws.StringValue(rawDesc.AssociationDescription.AssociationId),
		Targets:                     parseAssociationTargets(rawDesc.AssociationDescription.Targets),
		LastExecutionDate:           rawDesc.AssociationDescription.LastExecutionDate.Format(dateLayout),
		LastSuccessfulExecutionDate: rawDesc.AssociationDescription.LastSuccessfulExecutionDate.Format(dateLayout),
		ApplyOnlyAtCronInterval:     aws.BoolValue(rawDesc.AssociationDescription.ApplyOnlyAtCronInterval),
	}
}

func parseAssociationTargets(rawTgts []*ssm.Target) (tgts []AssociationTarget) {
	for _, rt := range rawTgts {
		t := AssociationTarget{Key: aws.StringValue(rt.Key)}
		for _, rv := range rt.Values {
			t.Values = append(t.Values, aws.StringValue(rv))
		}
		tgts = append(tgts, t)
	}
	return tgts
}

type Ec2InstanceStateChangeRequest struct {
	State string
}

type SSMCreateRequest struct {
	Document string
}

type SsmCommandRequest struct {
	DocumentName   string               `json:"document_name"`
	Parameters     map[string][]*string `json:"parameters"`
	TimeoutSeconds *int64               `json:"timeout"`
}
