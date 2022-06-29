package ec2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func (m *mockEC2Client) ModifyInstanceAttributeWithContext(ctx aws.Context, inp *ec2.ModifyInstanceAttributeInput, opt ...request.Option) (*ec2.ModifyInstanceAttributeOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.ModifyInstanceAttributeOutput{}, nil

}
func TestEc2_UpdateAttributes(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.ModifyInstanceAttributeInput
	}
	tests := []struct {
		name    string
		fields  fields
		e       *Ec2
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(), input: &ec2.ModifyInstanceAttributeInput{
				InstanceType: &ec2.AttributeValue{Value: aws.String("Type1")},
				InstanceId:   aws.String("i-123"),
			}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: false,
		},
		{
			name: "aws error",
			args: args{ctx: context.TODO(), input: &ec2.ModifyInstanceAttributeInput{
				InstanceType: &ec2.AttributeValue{Value: aws.String("Type1")},
				InstanceId:   aws.String("i-123"),
			}},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "invalid input, input is empty",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			if err := e.UpdateAttributes(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Ec2.UpdateAttributes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
