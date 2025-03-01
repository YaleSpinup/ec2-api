package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/YaleSpinup/ec2-api/ssm"
	"github.com/gorilla/mux"
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

	req := &Ec2SecurityGroupRuleRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		handleError(w, err)
		return
	}

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

// SSMAssociationByTagHandler handler function for creating an ssm association in aws by tag targets
// 1. Get the vars and account id
// 2. Decode the request
// 3. Check for missing values in the request body
// 4. Construct the role arn to assume in the session
// 5. Get the policy required for creating the ssm association via the session
// 6. Create the session required to access aws
// 7. Create the ssm service that houses the repository calls for creating the association by tag targets
// 8. Call the underlying service
// 9. Return the created association id
func (s *server) SSMAssociationByTagHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	// decode the request
	req := &SSMAssociationByTagRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into ssm create input %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	// Check for missing values in request body
	errMsg := ""
	if req.Document == "" {
		errMsg = fmt.Sprintf("document is mandatory")
	}
	if len(req.TagFilters) == 0 {
		errMsg = fmt.Sprintf("tagFilters is mandatory")
	}
	for _, tagValues := range req.TagFilters {
		if len(tagValues) == 0 {
			errMsg = fmt.Sprintf("You are missing values for one of your tags")
		}
	}
	if errMsg != "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, errMsg, nil))
		return
	}

	// construct the role arn by name in the account
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	// get the ssm association policy for creating the association
	policy, pErr := ssmAssociationPolicy()
	if pErr != nil {
		handleError(w, pErr)
		return
	}

	// create the session
	session, sErr := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		policy,
		"arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess",
		"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
	)
	if sErr != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, apierror.New(apierror.ErrForbidden, msg, sErr))
		return
	}

	service := ssm.New(
		ssm.WithSession(session.Session),
	)

	out, cErr := service.CreateAssociationByTag(r.Context(), req.Name, req.Document, req.DocumentVersion, req.ScheduleExpression, req.ScheduleOffset, req.TagFilters, req.Parameters)
	if cErr != nil {
		handleError(w, cErr)
		return
	}

	handleResponseOk(w, out)
}
