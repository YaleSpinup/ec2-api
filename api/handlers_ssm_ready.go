package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ssm"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// InstanceSSMReadyCheckHandler checks if SSM is ready on an instance
func (s *server) InstanceSSMReadyCheckHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	instanceID := vars["id"]

	log.Infof("checking SSM readiness for instance %s in account %s", instanceID, account)

	// Validate inputs
	if instanceID == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "instance id is required", nil))
		return
	}

	// Assume the appropriate role for the account
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)
	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		"",
		"arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess",
	)
	if err != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, apierror.New(apierror.ErrForbidden, msg, err))
		return
	}

	// We already have our role and session from above, just reusing them here

	ssmSvc := ssm.New(
		ssm.WithSession(session.Session),
	)

	// Check if instance is managed by SSM
	isManagedInstance, err := checkInstanceSSMStatus(r.Context(), ssmSvc, instanceID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return the result
	response := map[string]interface{}{
		"instanceId": instanceID,
		"ready":      isManagedInstance,
	}

	handleResponseOk(w, response)
}

// checkInstanceSSMStatus checks if an instance is managed by SSM
func checkInstanceSSMStatus(ctx context.Context, ssmSvc *ssm.SSM, instanceID string) (bool, error) {
	log.Debugf("checking SSM status for instance: %s", instanceID)

	// Get the list of SSM managed instances to see if our instance is managed
	instances, err := ssmSvc.GetInstanceInformationWithFilters(ctx, map[string]string{
		"InstanceIds": instanceID,
	})
	if err != nil {
		return false, err
	}

	// If the instance is in the list and ping status is 'Online', it's ready
	for _, instance := range instances {
		if instance.InstanceId != nil && *instance.InstanceId == instanceID && 
		   instance.PingStatus != nil && *instance.PingStatus == "Online" {
			log.Infof("instance %s is managed by SSM and online", instanceID)
			return true, nil
		}
	}

	// Instance either not managed by SSM or not online
	log.Infof("instance %s is not yet managed by SSM or not online", instanceID)
	return false, nil
}