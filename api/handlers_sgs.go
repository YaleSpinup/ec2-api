package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bytes"
	"io"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (s *server) SecurityGroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)
	policy, err := sgCreatePolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	req := &Ec2SecurityGroupRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		handleError(w, err)
		return
	}

	orch, err := s.newEc2Orchestrator(r.Context(), &sessionParams{
		inlinePolicy: policy,
		role:         role,
		policyArns: []string{
			"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
		},
	})
	if err != nil {
		handleError(w, err)
		return
	}

	out, err := orch.createSecurityGroup(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

func (s *server) SecurityGroupUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	// Log raw request body
	body, _ := io.ReadAll(r.Body)
	log.Debugf("Raw request body: %s", string(body))
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for later reading

	req := &Ec2SecurityGroupRuleRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		handleError(w, err)
		return
	}

	log.Debugf("SecurityGroupUpdateHandler received request: %+v", awsutil.Prettify(req))
	if (req.Tags != nil) == (req.RuleType != nil) {
		handleError(w, apierror.New(apierror.ErrBadRequest, "request should either update tags or modify security group", nil))
		return
	}

	var policy string
	var err error

	if req.Tags != nil {
		policy, err = tagCreatePolicy()
	} else {
		policy, err = sgUpdatePolicy(id)
	}

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

	if req.Tags != nil {
		if err := orch.ec2Client.UpdateRawTags(r.Context(), *req.Tags, id); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		if err := orch.updateSecurityGroup(r.Context(), id, req); err != nil {
			handleError(w, err)
			return
		}
		handleResponseOk(w, nil)
	}
}

func (s *server) SecurityGroupListHandler(w http.ResponseWriter, r *http.Request) {
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

	service := ec2.New(
		ec2.WithSession(session.Session),
		ec2.WithOrg(s.org),
	)

	out, err := service.ListSecurityGroups(r.Context(), "")
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("X-Items", strconv.Itoa(len(out)))

	handleResponseOk(w, out)
}

func (s *server) SecurityGroupGetHandler(w http.ResponseWriter, r *http.Request) {
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

	out, err := service.GetSecurityGroup(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	if len(out) == 0 {
		handleError(w, apierror.New(apierror.ErrNotFound, "not found", nil))
		return
	}

	if len(out) > 1 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "unexpected security group count returned", nil))
		return
	}

	handleResponseOk(w, toEc2SecurityGroupResponse(out[0]))
}

func (s *server) SecurityGroupDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)
	policy, err := sgDeletePolicy(id)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to generate policy", err))
		return
	}

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		policy,
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

	if err := service.DeleteSecurityGroup(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, "OK")
}
