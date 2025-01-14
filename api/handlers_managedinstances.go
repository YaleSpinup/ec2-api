package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ListManagedInstancesHandler lists all hybrid (managed) instances in an account
func (s *server) ListManagedInstancesHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	log.Infof("listing managed instances in account: %s", account)

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

	service := ssm.New(
		ssm.WithSession(session.Session),
	)

	perPage := int64(10) // default value
	var pageToken *string
	for name, values := range r.URL.Query() {
		if name == "next" {
			pageToken = aws.String(values[0])
		}

		if name == "limit" {
			limit, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				handleError(w, apierror.New(apierror.ErrBadRequest, "invalid value for limit parameter", err))
				return
			}
			perPage = limit
		}
	}

	instances, next, err := service.ListManagedInstances(r.Context(), perPage, pageToken)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("X-Items", strconv.Itoa(len(instances)))
	if next != nil {
		w.Header().Set("X-Per-Page", strconv.FormatInt(perPage, 10))
		w.Header().Set("X-Next-Token", aws.StringValue(next))
	}

	handleResponseOk(w, instances)
}

// GetManagedInstanceHandler gets details about a specific hybrid (managed) instance
func (s *server) GetManagedInstanceHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	identifier := vars["id"]

	log.Infof("getting managed instance with identifier %s in account: %s", identifier, account)

	if identifier == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "identifier (instance_id or computer_name) is required", nil))
		return
	}

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

	service := ssm.New(
		ssm.WithSession(session.Session),
	)

	instance, err := service.GetManagedInstance(r.Context(), identifier)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, instance)
}
