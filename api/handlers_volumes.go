package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
)

func (s *server) VolumeListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		"",
		"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
	)
	if err != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, apierror.New(apierror.ErrForbidden, msg, err))
		return
	}

	perPage := 0
	var pageToken *string
	for name, values := range r.URL.Query() {
		if name == "next" {
			pageToken = aws.String(values[0])
		}

		if name == "limit" {
			limit, err := strconv.Atoi(values[0])
			if err != nil {
				handleError(w, fmt.Errorf("failed to parse limit parameter: %s", err))
				return
			}
			perPage = limit
		}
	}

	service := ec2.New(
		ec2.WithSession(session.Session),
		ec2.WithOrg(s.org),
	)

	out, next, err := service.ListVolumes(r.Context(), "", int64(perPage), pageToken)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("X-Items", strconv.Itoa(len(out)))
	if next != nil {
		w.Header().Set("X-Per-Page", strconv.Itoa(perPage))
		w.Header().Set("X-Next-Token", aws.StringValue(next))
	}

	handleResponseOk(w, out)
}

func (s *server) VolumeGetHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		"",
		"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
	)
	if err != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, apierror.New(apierror.ErrForbidden, msg, err))
		return
	}

	service := ec2.New(
		ec2.WithSession(session.Session),
		ec2.WithOrg(s.org),
	)

	out, err := service.GetVolume(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	if len(out) == 0 {
		handleError(w, apierror.New(apierror.ErrNotFound, "not found", nil))
		return
	}

	if len(out) > 1 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "unexpected volume count returned", nil))
		return
	}

	handleResponseOk(w, toEc2VolumeResponse(out[0]))
}
