package iam

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

// mockIAMClient is a fake iam client
type mockIAMClient struct {
	iamiface.IAMAPI
	t   *testing.T
	err error
}

func newmockIAMClient(t *testing.T, err error) iamiface.IAMAPI {
	return &mockIAMClient{
		t:   t,
		err: err,
	}
}

func TestNewSession(t *testing.T) {
	client := New()
	to := reflect.TypeOf(client).String()
	if to != "*iam.IAM" {
		t.Errorf("expected type to be '*iam.IAM', got %s", to)
	}
}
