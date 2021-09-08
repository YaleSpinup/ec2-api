package api

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
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
	return t.UTC().Format("2006-01-02 15:04:05 MST")
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

type Ec2SecurityGroupUserIdGroupPair struct {
	Description          string `json:"description"`
	GroupId              string `json:"group_id"`
	GroupName            string `json:"group_name"`
	PeeringStatus        string `json:"peering_status"`
	UserId               string `json:"user_id"`
	VpcId                string `json:"vpc_id"`
	VpcPeeringConnection string `json:"vpc_peering_connection"`
}

type Ec2SecurityGroupPrefixListId struct {
	Description  string `json:"description"`
	PrefixListId string `json:"prefix_list_id"`
}

type Ec2SecurityGroupIpv6Range struct {
	CidrIpv6    string `json:"cidr_ipv_6"`
	Description string `json:"description"`
}

type Ec2SecurityGroupIpRange struct {
	CidrIp      string `json:"cidr_ip"`
	Description string `json:"description"`
}

type Ec2SecurityGroupIpPermission struct {
	FromPort         int64                              `json:"from_port"`
	IpProtocol       string                             `json:"ip_protocol"`
	IpRanges         []*Ec2SecurityGroupIpRange         `json:"ip_ranges"`
	Ipv6Ranges       []*Ec2SecurityGroupIpv6Range       `json:"ipv_6_ranges"`
	PrefixListIds    []*Ec2SecurityGroupPrefixListId    `json:"prefix_list_ids"`
	ToPort           int64                              `json:"to_port"`
	UserIdGroupPairs []*Ec2SecurityGroupUserIdGroupPair `json:"user_id_group_pairs"`
}

type Ec2SecurityGroupResponse struct {
	Description   string                          `json:"description"`
	GroupId       string                          `json:"group_id"`
	GroupName     string                          `json:"group_name"`
	IncomingRules []*Ec2SecurityGroupIpPermission `json:"incoming_rules"`
	OutgoingRules []*Ec2SecurityGroupIpPermission `json:"outgoing_rules"`
	OwnerId       string                          `json:"owner_id"`
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
	return &Ec2SecurityGroupResponse{
		Description:   aws.StringValue(sg.Description),
		IncomingRules: toEc2SecurityGroupIpPermissions(sg.IpPermissions),
		OutgoingRules: toEc2SecurityGroupIpPermissions(sg.IpPermissionsEgress),
		GroupId:       aws.StringValue(sg.GroupId),
		GroupName:     aws.StringValue(sg.GroupName),
		OwnerId:       aws.StringValue(sg.OwnerId),
		VpcId:         aws.StringValue(sg.VpcId),
	}
}
