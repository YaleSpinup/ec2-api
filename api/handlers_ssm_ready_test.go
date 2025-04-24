package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/gorilla/mux"
)

// TestInstanceSSMReadyCheckHandler tests the instance SSM ready check handler
func TestInstanceSSMReadyCheckHandler(t *testing.T) {
	// Test case 1: Missing instance ID
	req, err := http.NewRequest("GET", "/v2/ec2/123456789012/instances//ssm/ready", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add path parameters to the request context
	vars := map[string]string{
		"account": "123456789012",
		"id":      "",
	}
	req = mux.SetURLVars(req, vars)
	
	rr := httptest.NewRecorder()
	s := server{}
	
	handler := http.HandlerFunc(s.InstanceSSMReadyCheckHandler)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// mockSSMService is a mock for the SSM service
type mockSSMService struct {
	instances []*ssm.InstanceInformation
	err       error
}

func (m *mockSSMService) GetInstanceInformationWithFilters(ctx context.Context, filters map[string]string) ([]*ssm.InstanceInformation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.instances, nil
}