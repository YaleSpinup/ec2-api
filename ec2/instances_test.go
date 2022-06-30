package ec2

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func (m *mockEC2Client) DescribeInstancesWithContext(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...request.Option) (*ec2.DescribeInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	if len(input.InstanceIds) != 0 && aws.StringValue(input.InstanceIds[0]) == "i-notfound" {
		return &ec2.DescribeInstancesOutput{}, nil
	} else if len(input.InstanceIds) != 0 && aws.StringValue(input.InstanceIds[0]) == "i-multiple" {
		return &ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{},
		}, nil
	}

	return nil, nil
}

func (m mockEC2Client) RunInstancesWithContext(ctx context.Context, input *ec2.RunInstancesInput, opts ...request.Option) (*ec2.Reservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	// return multiple Instances (unexpected)
	if aws.StringValue(input.InstanceType) == "t3.weird" {
		return &ec2.Reservation{
			Instances: []*ec2.Instance{
				{InstanceId: aws.String("i-0123456789abcdef0")},
				{InstanceId: aws.String("i-0123456789abcdef1")},
				{InstanceId: aws.String("i-0123456789abcdef2")},
			},
		}, nil
	}

	return &ec2.Reservation{
		Instances: []*ec2.Instance{
			{InstanceId: aws.String("i-0123456789abcdef0")},
		},
	}, nil
}

func (m mockEC2Client) TerminateInstancesWithContext(ctx context.Context, input *ec2.TerminateInstancesInput, opts ...request.Option) (*ec2.TerminateInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.TerminateInstancesOutput{}, nil
}

func (m mockEC2Client) StartInstancesWithContext(ctx context.Context, input *ec2.StartInstancesInput, opts ...request.Option) (*ec2.StartInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.StartInstancesOutput{}, nil
}

func (m mockEC2Client) StopInstancesWithContext(ctx context.Context, input *ec2.StopInstancesInput, opts ...request.Option) (*ec2.StopInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.StopInstancesOutput{}, nil
}

func (m mockEC2Client) RebootInstancesWithContext(ctx context.Context, input *ec2.RebootInstancesInput, opts ...request.Option) (*ec2.RebootInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.RebootInstancesOutput{}, nil
}

func TestEc2_CreateInstance(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx   context.Context
		input *ec2.RunInstancesInput
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ec2.Instance
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "unexpected output",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				input: &ec2.RunInstancesInput{
					InstanceType: aws.String("t3.weird"),
				},
			},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				input: &ec2.RunInstancesInput{
					InstanceType: aws.String("t3.nano"),
				},
			},
			want: &ec2.Instance{
				InstanceId: aws.String("i-0123456789abcdef0"),
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			got, err := e.CreateInstance(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.CreateInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.CreateInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_DeleteInstance(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx   context.Context
		input string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx:   context.TODO(),
				input: "i-0123456789abcdef0",
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			err := e.DeleteInstance(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.DeleteInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEc2_GetInstance(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ec2.Instance
		wantErr bool
	}{
		{
			name:   "empty id",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				id:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			got, err := e.GetInstance(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.GetInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.GetInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_StartInstance(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx context.Context
		ids []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-0123456789abcdef0"},
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-0123456789abcdef0"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			err := e.StartInstance(tt.args.ctx, tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.StartInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEc2_StopInstance(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx   context.Context
		force bool
		ids   []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx:   context.TODO(),
				force: true,
				ids:   []string{"i-0123456789abcdef0"},
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-0123456789abcdef0"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			err := e.StopInstance(tt.args.ctx, tt.args.force, tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.StopInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEc2_RebootInstance(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx context.Context
		ids []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-0123456789abcdef0"},
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-0123456789abcdef0"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			err := e.RebootInstance(tt.args.ctx, tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.RebootInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
