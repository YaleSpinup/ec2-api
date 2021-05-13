package api

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// timeFormat returns the standardized time format
func timeFormat(t *time.Time) string {
	return t.UTC().UTC().Format(time.RFC3339)
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
	log.Debugf("mapping ec2 instance %s", awsutil.Prettify(instance))

	tags := make(map[string]string, len(instance.Tags))
	for _, t := range instance.Tags {
		tags[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	name, ok := tags["Name"]
	if !ok {
		log.Warnf("instance %s doesn't have a Name tag", aws.StringValue(instance.InstanceId))
	}

	tagsList := make([]map[string]string, 0, len(tags))
	for k, v := range tags {
		tagsList = append(tagsList, map[string]string{k: v})
	}

	platform := aws.StringValue(instance.Platform)
	if platform == "" {
		platform = "linux"
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
			AttachTime:          timeFormat(v.Ebs.AttachTime),
			DeleteOnTermination: aws.BoolValue(v.Ebs.DeleteOnTermination),
			Status:              aws.StringValue(v.Ebs.Status),
			DeviceName:          aws.StringValue(v.DeviceName),
		}
	}

	response := Ec2InstanceResponse{
		Az: aws.StringValue(instance.Placement.AvailabilityZone),
		// TODO from tags
		CreatedAt: timeFormat(instance.LaunchTime),
		CreatedBy: tags["CreatedBy"],
		ID:        aws.StringValue(instance.InstanceId),
		Image:     aws.StringValue(instance.ImageId),
		Ip:        aws.StringValue(instance.PrivateIpAddress),
		Key:       aws.StringValue(instance.KeyName),
		Name:      name,
		Platform:  platform,
		Sgs:       sgs,
		State:     aws.StringValue(instance.State.Name),
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
	log.Debugf("mapping ec2 instance volume %s", awsutil.Prettify(volume))

	tags := make(map[string]string, len(volume.Tags))
	for _, t := range volume.Tags {
		tags[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	tagsList := make([]map[string]string, 0, len(tags))
	for k, v := range tags {
		tagsList = append(tagsList, map[string]string{k: v})
	}

	attachments := []*Ec2VolumeAttachment{}
	for _, a := range volume.Attachments {
		attachments = append(attachments, &Ec2VolumeAttachment{
			AttachTime:          timeFormat(a.AttachTime),
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
