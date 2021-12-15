package ec2

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// mockEC2Client is a fake ec2 client
type mockEC2Client struct {
	ec2iface.EC2API
	t   *testing.T
	err error
}

func newmockEC2Client(t *testing.T, err error) ec2iface.EC2API {
	return &mockEC2Client{
		t:   t,
		err: err,
	}
}

func TestNewSession(t *testing.T) {
	client := New()
	to := reflect.TypeOf(client).String()
	if to != "*ec2.Ec2" {
		t.Errorf("expected type to be '*ec2.Ec2', got %s", to)
	}
}
