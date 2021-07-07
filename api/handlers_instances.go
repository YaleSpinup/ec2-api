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

func (s *server) InstanceListHandler(w http.ResponseWriter, r *http.Request) {
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

	perPage := 500
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

	// TODO an api should be for one org, currently we need to suppor the entire acocunt
	out, next, err := service.ListInstances(r.Context(), "", int64(perPage), pageToken)
	// out, next, err := service.ListInstances(r.Context(), s.org, int64(perPage), pageToken)
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

func (s *server) InstanceGetHandler(w http.ResponseWriter, r *http.Request) {
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

	out, err := service.GetInstance(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, toEc2InstanceResponse(out))
}

func (s *server) InstanceVolumesHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]
	vid := vars["vid"]

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

	if vid == "" {
		out, err := service.ListInstanceVolumes(r.Context(), id)
		if err != nil {
			handleError(w, err)
			return
		}
		handleResponseOk(w, out)
	} else {
		out, err := service.GetInstanceVolume(r.Context(), id, vid)
		if err != nil {
			handleError(w, err)
			return
		}

		handleResponseOk(w, toEc2VolumeResponse(out))
	}
}

func (s *server) InstanceListSnapshotsHandler(w http.ResponseWriter, r *http.Request) {
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

	out, err := service.ListInstanceSnapshots(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	list := make([]map[string]string, 0, len(out))
	for _, s := range out {
		list = append(list, map[string]string{"id": s})
	}

	handleResponseOk(w, list)
}
