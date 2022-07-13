package ec2

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

var images = []*ec2.Image{
	{
		ImageId: aws.String("i-00000001"),
		Name:    aws.String("Image_00000001"),
		OwnerId: aws.String("self"),
		Public:  aws.Bool(false),
		State:   aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("dev"),
			},
		},
	},
	{
		ImageId: aws.String("i-00000002"),
		Name:    aws.String("Image_00000002"),
		OwnerId: aws.String("self"),
		Public:  aws.Bool(false),
		State:   aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("dev"),
			},
		},
	},
	{
		ImageId: aws.String("i-00000003"),
		Name:    aws.String("Image_00000003"),
		OwnerId: aws.String("self"),
		Public:  aws.Bool(false),
		State:   aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("tst"),
			},
		},
	},
	{
		ImageId: aws.String("i-00000004"),
		Name:    aws.String("Image_00000004"),
		OwnerId: aws.String("self"),
		Public:  aws.Bool(false),
		State:   aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("tst"),
			},
		},
	},
	{
		ImageId: aws.String("i-00000005"),
		Name:    aws.String("Image_00000005"),
		OwnerId: aws.String("aws"),
		Public:  aws.Bool(false),
		State:   aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("dev"),
			},
		},
	},
	{
		ImageId: aws.String("i-00000006"),
		Name:    aws.String("Image_00000006"),
		OwnerId: aws.String("self"),
		Public:  aws.Bool(true),
		State:   aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("dev"),
			},
		},
	},
	{
		ImageId: aws.String("i-00000007"),
		Name:    aws.String("Image_00000007"),
		OwnerId: aws.String("self"),
		Public:  aws.Bool(false),
		State:   aws.String("pending"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("yale:org"),
				Value: aws.String("dev"),
			},
		},
	},
}

func (m mockEC2Client) DescribeImagesWithContext(ctx context.Context, input *ec2.DescribeImagesInput, opts ...request.Option) (*ec2.DescribeImagesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	imageList := []*ec2.Image{}
	for _, i := range images {
		m.t.Logf("testing image %s against filters", aws.StringValue(i.ImageId))

		if input.ImageIds != nil {
			m.t.Logf("checking passed image ids %+v", input.ImageIds)

			var match bool
			for _, id := range input.ImageIds {
				if aws.StringValue(id) == aws.StringValue(i.ImageId) {
					match = true
					break
				}
			}

			if !match {
				continue
			}
		}

		if input.Owners != nil {
			m.t.Logf("checking passed image owners %+v", input.Owners)

			var match bool
			for _, o := range input.Owners {
				if aws.StringValue(o) == aws.StringValue(i.OwnerId) {
					match = true
					break
				}
			}

			if !match {
				continue
			}
		}

		if input.Filters != nil {
			match := true

			for _, f := range input.Filters {
				m.t.Logf("checking passed image filter %+v", f)

				switch aws.StringValue(f.Name) {
				case "name":
					var filterMatch bool
					for _, v := range f.Values {
						if aws.StringValue(v) == aws.StringValue(i.Name) {
							filterMatch = true
							break
						}
					}

					if !filterMatch {
						match = false
					}
				case "is-public":
					var filterMatch bool
					for _, v := range f.Values {
						b, err := strconv.ParseBool(aws.StringValue(v))
						if err != nil {
							m.t.Logf("failed to parse bool value: %s", aws.StringValue(v))
							continue
						}

						if b == aws.BoolValue(i.Public) {
							filterMatch = true
							break
						}
					}

					if !filterMatch {
						match = false
					}
				case "state":
					var filterMatch bool
					for _, v := range f.Values {
						if aws.StringValue(v) == aws.StringValue(i.State) {
							filterMatch = true
							break
						}
					}

					if !filterMatch {
						match = false
					}
				case "tag:yale:org":
					var filterMatch bool
					var tagValue string

					for _, tag := range i.Tags {
						if aws.StringValue(tag.Key) == "yale:org" {
							tagValue = aws.StringValue(tag.Value)
						}
					}

					for _, v := range f.Values {
						if aws.StringValue(v) == tagValue {
							filterMatch = true
							break
						}
					}

					if !filterMatch {
						match = false
					}
				}
			}

			if !match {
				continue
			}
		}

		m.t.Logf("image %s matches filters", aws.StringValue(i.ImageId))

		imageList = append(imageList, i)
	}

	return &ec2.DescribeImagesOutput{Images: imageList}, nil
}

func (m mockEC2Client) CreateImageWithContext(ctx aws.Context, input *ec2.CreateImageInput, opts ...request.Option) (*ec2.CreateImageOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.CreateImageOutput{ImageId: aws.String("1234")}, nil
}

func (m mockEC2Client) DeregisterImageWithContext(ctx aws.Context, input *ec2.DeregisterImageInput, opts ...request.Option) (*ec2.DeregisterImageOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DeregisterImageOutput{}, nil
}

func TestEc2_ListImages(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx  context.Context
		org  string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []map[string]*string
		wantErr bool
	}{
		{
			name:   "empty org",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				org: "",
			},
			want: []map[string]*string{
				{
					"id":   aws.String("i-00000001"),
					"name": aws.String("Image_00000001"),
				},
				{
					"id":   aws.String("i-00000002"),
					"name": aws.String("Image_00000002"),
				},
				{
					"id":   aws.String("i-00000003"),
					"name": aws.String("Image_00000003"),
				},
				{
					"id":   aws.String("i-00000004"),
					"name": aws.String("Image_00000004"),
				},
			},
		},
		{
			name:   "dev org",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				org: "dev",
			},
			want: []map[string]*string{
				{
					"id":   aws.String("i-00000001"),
					"name": aws.String("Image_00000001"),
				},
				{
					"id":   aws.String("i-00000002"),
					"name": aws.String("Image_00000002"),
				},
			},
		},
		{
			name:   "name Image_00000002",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx:  context.TODO(),
				name: "Image_00000002",
			},
			want: []map[string]*string{
				{
					"id":   aws.String("i-00000002"),
					"name": aws.String("Image_00000002"),
				},
			},
		},
		{
			name:   "dev org with name Image_00000001",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx:  context.TODO(),
				org:  "dev",
				name: "Image_00000001",
			},
			want: []map[string]*string{
				{
					"id":   aws.String("i-00000001"),
					"name": aws.String("Image_00000001"),
				},
			},
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
			got, err := e.ListImages(tt.args.ctx, tt.args.org, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ListImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ListImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_GetImage(t *testing.T) {
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
		want    []*ec2.Image
		wantErr bool
	}{
		{
			name:    "nil ids",
			fields:  fields{Service: newmockEC2Client(t, nil)},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name:   "empty ids",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{},
			},
			wantErr: true,
		},
		{
			name:   "one id",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-00000007"},
			},
			want: []*ec2.Image{
				{
					ImageId: aws.String("i-00000007"),
					Name:    aws.String("Image_00000007"),
					OwnerId: aws.String("self"),
					Public:  aws.Bool(false),
					State:   aws.String("pending"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("dev"),
						},
					},
				},
			},
		},
		{
			name:   "one id not owned by self",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{"i-00000005"},
			},
			want: []*ec2.Image{
				{
					ImageId: aws.String("i-00000005"),
					Name:    aws.String("Image_00000005"),
					OwnerId: aws.String("aws"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("dev"),
						},
					},
				},
			},
		},
		{
			name:   "multiple ids",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{
					"i-00000002",
					"i-00000003",
					"i-00000004",
				},
			},
			want: []*ec2.Image{
				{
					ImageId: aws.String("i-00000002"),
					Name:    aws.String("Image_00000002"),
					OwnerId: aws.String("self"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("dev"),
						},
					},
				},
				{
					ImageId: aws.String("i-00000003"),
					Name:    aws.String("Image_00000003"),
					OwnerId: aws.String("self"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("tst"),
						},
					},
				},
				{
					ImageId: aws.String("i-00000004"),
					Name:    aws.String("Image_00000004"),
					OwnerId: aws.String("self"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("tst"),
						},
					},
				},
			},
		},
		{
			name:   "multiple ids with some not owned by self",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{
					"i-00000003",
					"i-00000004",
					"i-00000005",
				},
			},
			want: []*ec2.Image{
				{
					ImageId: aws.String("i-00000003"),
					Name:    aws.String("Image_00000003"),
					OwnerId: aws.String("self"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("tst"),
						},
					},
				},
				{
					ImageId: aws.String("i-00000004"),
					Name:    aws.String("Image_00000004"),
					OwnerId: aws.String("self"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("tst"),
						},
					},
				},
				{
					ImageId: aws.String("i-00000005"),
					Name:    aws.String("Image_00000005"),
					OwnerId: aws.String("aws"),
					Public:  aws.Bool(false),
					State:   aws.String("available"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("dev"),
						},
					},
				},
			},
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
			got, err := e.GetImage(tt.args.ctx, tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.GetImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.GetImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_CreateImage(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.CreateImageInput
	}
	tests := []struct {
		name    string
		fields  fields
		e       *Ec2
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.CreateImageInput{InstanceId: aws.String("1234"), Name: aws.String("Testcase1")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			want:    "1234",
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.CreateImageInput{InstanceId: aws.String("1234"), Name: aws.String("Testcase1")}},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			got, err := e.CreateImage(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.CreateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ec2.CreateImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_DeregisterImage(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.DeregisterImageInput
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: true,
		},
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.DeregisterImageInput{ImageId: aws.String("v-123")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.DeregisterImageInput{ImageId: aws.String("v-123")}},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			err := e.DeregisterImage(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.DeregisterImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
