package resourcegroups

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

func TestWithSession(t *testing.T) {
	sess := session.Must(session.NewSession())
	rg := New(WithSession(sess))

	if rg.session != sess {
		t.Error("WithSession option did not set the session correctly")
	}

	if rg.Service == nil {
		t.Error("Service was not initialized with session")
	}
}

func TestWithCredentials(t *testing.T) {
	cases := []struct {
		name   string
		key    string
		secret string
		token  string
		region string
	}{
		{
			name:   "with all credentials",
			key:    "test-key",
			secret: "test-secret",
			token:  "test-token",
			region: "us-east-1",
		},
		{
			name:   "without token",
			key:    "test-key",
			secret: "test-secret",
			token:  "",
			region: "us-west-2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rg := New(WithCredentials(tc.key, tc.secret, tc.token, tc.region))

			if rg.session == nil {
				t.Error("session was not created")
			}

			if rg.Service == nil {
				t.Error("Service was not initialized")
			}
		})
	}
}

func TestNewWithMultipleOptions(t *testing.T) {
	cases := []struct {
		name    string
		opts    []Option
		wantNil bool
	}{
		{
			name:    "no options",
			opts:    []Option{},
			wantNil: true,
		},
		{
			name: "with credentials",
			opts: []Option{
				WithCredentials("key", "secret", "", "us-east-1"),
			},
			wantNil: false,
		},
		{
			name: "with session",
			opts: []Option{
				WithSession(session.Must(session.NewSession())),
			},
			wantNil: false,
		},
		{
			name: "with both options",
			opts: []Option{
				WithSession(session.Must(session.NewSession())),
				WithCredentials("key", "secret", "", "us-east-1"),
			},
			wantNil: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rg := New(tc.opts...)

			if tc.wantNil {
				if rg.Service != nil {
					t.Error("Service should be nil when no options are provided")
				}
			} else {
				if rg.Service == nil {
					t.Error("Service should not be nil when options are provided")
				}
			}
		})
	}
}
