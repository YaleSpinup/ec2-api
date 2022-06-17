package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
)

func (s *server) ImageListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

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

	// TODO only return images from our org, current EC2-API returns all (needed for managed)
	out, err := service.ListImages(r.Context(), "", name)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("X-Items", strconv.Itoa(len(out)))

	handleResponseOk(w, out)
}

func (s *server) ImageGetHandler(w http.ResponseWriter, r *http.Request) {
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

	out, err := service.GetImage(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	if len(out) == 0 {
		handleError(w, apierror.New(apierror.ErrNotFound, "not found", nil))
		return
	}

	if len(out) > 1 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "unexpected image count returned", nil))
		return
	}

	handleResponseOk(w, toEc2ImageResponse(out[0]))
}

func (s *server) ImageUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	req := &Ec2ImageUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into update image input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if len(req.Tags) == 0 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: tags", nil))
		return
	}

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)
	policy, err := tagCreatePolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		policy,
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

	if err := service.UpdateRawTags(r.Context(), req.Tags, id); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) ImageCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	req := &Ec2ImageCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into create image input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.InstanceId == nil || req.Name == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required fields: instance_id, name", nil))
		return
	}
	if req.ForceReboot == nil {
		req.ForceReboot = aws.Bool(false)
	}
	if req.CopyTags == nil {
		req.CopyTags = aws.Bool(true)
	}

	policy, err := generatePolicy([]string{"ec2:CreateImage", "ec2:CreateTags"})
	if err != nil {
		handleError(w, err)
		return
	}

	orch, err := s.newEc2Orchestrator(r.Context(), &sessionParams{
		role:         fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		inlinePolicy: policy,
		policyArns: []string{
			"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
		},
	})
	if err != nil {
		handleError(w, err)
		return
	}

	out, err := orch.createImage(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}
