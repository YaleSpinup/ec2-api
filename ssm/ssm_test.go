package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type mockSSMClient struct {
	ssmiface.SSMAPI
	t   *testing.T
	err error
}

func newMockSSMClient(t *testing.T, err error) ssmiface.SSMAPI {
	return &mockSSMClient{
		t:   t,
		err: err,
	}
}
