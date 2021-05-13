package api

import (
	"context"
	"net/url"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/aws-sdk-go/aws"

	log "github.com/sirupsen/logrus"
)

func Test_parseFilterQueries(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	type args struct {
		ctx context.Context
		q   url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    selector.Filters
		wantErr bool
	}{
		{
			name: "nil input",
			args: args{
				ctx: context.TODO(),
				q:   nil,
			},
			want: selector.Filters{},
		},
		{
			name: "one availabilityzone",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"AvailabilityZones": {"az1"},
				},
			},
			want: selector.Filters{AvailabilityZones: &[]string{"az1"}},
		},
		{
			name: "multiple availabilityzones",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"AvailabilityZones": {"az1", "az2", "az3"},
				},
			},
			want: selector.Filters{AvailabilityZones: &[]string{"az1", "az2", "az3"}},
		},
		{
			name: "allow list",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"AllowList": {"g3.*"},
				},
			},
			want: selector.Filters{AllowList: regexp.MustCompile("g3.*")},
		},
		{
			name: "deny list",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"DenyList": {"g3.*"},
				},
			},
			want: selector.Filters{DenyList: regexp.MustCompile("g3.*")},
		},
		{
			name: "integer example",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"MaxResults": {"17"},
				},
			},
			want: selector.Filters{MaxResults: aws.Int(17)},
		},
		{
			name: "float64 example",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"VCpusToMemoryRatio": {"1.17"},
				},
			},
			want: selector.Filters{VCpusToMemoryRatio: aws.Float64(1.17)},
		},
		{
			name: "boolean example true",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"BareMetal": {"true"},
				},
			},
			want: selector.Filters{BareMetal: aws.Bool(true)},
		},
		{
			name: "boolean example false",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"BareMetal": {"false"},
				},
			},
			want: selector.Filters{BareMetal: aws.Bool(false)},
		},
		{
			name: "range integers",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"GpusRange": {"1-500"},
				},
			},
			want: selector.Filters{GpusRange: &selector.IntRangeFilter{
				UpperBound: 500,
				LowerBound: 1,
			}},
		},
		{
			name: "invalid range integers",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"GpusRange": {"1-500thousand"},
				},
			},
			want: selector.Filters{},
		},
		{
			name: "range bytes",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"GpuMemoryRange": {"1MB-5GB"},
				},
			},
			want: selector.Filters{GpuMemoryRange: &selector.ByteQuantityRangeFilter{
				UpperBound: bytequantity.FromGiB(5),
				LowerBound: bytequantity.FromMiB(1),
			}},
		},
		{
			name: "mixed int and bytes, default to GB",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"GpuMemoryRange": {"1-5GB"},
				},
			},
			want: selector.Filters{GpuMemoryRange: &selector.ByteQuantityRangeFilter{
				UpperBound: bytequantity.FromGiB(5),
				LowerBound: bytequantity.FromGiB(1),
			}},
		},
		{
			name: "invalid byte size",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"GpuMemoryRange": {"1-10gigabytes"},
				},
			},
			want: selector.Filters{},
		},
		{
			name: "string value",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"CPUArchitecture": {"x86_64"},
				},
			},
			want: selector.Filters{CPUArchitecture: aws.String("x86_64")},
		},
		{
			name: "filter value type mismatch",
			args: args{
				ctx: context.TODO(),
				q: url.Values{
					"CPUArchitecture": {"1234"},
				},
			},
			want:    selector.Filters{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilterQueries(tt.args.ctx, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFilterQueries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFilterQueries() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
