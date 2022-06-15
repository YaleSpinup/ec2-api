package ec2

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

var (
	inpIds  = []string{"id-234"}
	inpTags = map[string]string{"foo": "bar"}
	expTags = []*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}
)

func (m *mockEC2Client) CreateTagsWithContext(ctx context.Context, input *ec2.CreateTagsInput, opts ...request.Option) (*ec2.CreateTagsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !reflect.DeepEqual(input.Resources, aws.StringSlice(inpIds)) || !reflect.DeepEqual(input.Tags, expTags) {
		return nil, errors.New("input does not match")
	}
	return &ec2.CreateTagsOutput{}, nil
}

func (m *mockEC2Client) DescribeVolumesWithContext(aws aws.Context, inp *ec2.DescribeVolumesInput, opt ...request.Option) (*ec2.DescribeVolumesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DescribeVolumesOutput{}, nil
}

func TestEc2_UpdateTags(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx  context.Context
		tags map[string]string
		ids  []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), tags: inpTags, ids: inpIds},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), tags: inpTags, ids: inpIds},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "no tags",
			fields:  fields{Service: newmockEC2Client(t, nil)},
			args:    args{ctx: context.TODO(), tags: nil, ids: inpIds},
			wantErr: true,
		},
		{
			name:    "no ids",
			fields:  fields{Service: newmockEC2Client(t, nil)},
			args:    args{ctx: context.TODO(), tags: inpTags, ids: []string{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			err := e.UpdateTags(tt.args.ctx, tt.args.tags, tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.UpdateTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func TestEc2_UpdateInstanceTags(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.CreateTagsInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(), input: &ec2.CreateTagsInput{
				Resources: aws.StringSlice(inpIds),
				Tags:      expTags}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: false,
		},
		{
			name: "aws error",
			args: args{ctx: context.TODO(), input: &ec2.CreateTagsInput{
				Resources: aws.StringSlice(inpIds),
				Tags:      expTags}},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:   "no tags",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{ctx: context.TODO(), input: &ec2.CreateTagsInput{
				Resources: aws.StringSlice(inpIds),
				Tags:      nil}},
			wantErr: true,
		},
		{
			name:   "no ids",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{ctx: context.TODO(), input: &ec2.CreateTagsInput{
				Resources: aws.StringSlice([]string{}),
				Tags:      expTags}},
			wantErr: true,
		},
		{
			name:    "no input",
			fields:  fields{Service: newmockEC2Client(t, nil)},
			args:    args{ctx: context.TODO(), input: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			if err := e.UpdateInstanceTags(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Ec2.UpdateTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
