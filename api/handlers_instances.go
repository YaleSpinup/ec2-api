package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/YaleSpinup/ec2-api/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
)

func (s *server) InstanceCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	req := Ec2InstanceCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := fmt.Sprintf("cannot decode body into create instance input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.Type == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: type", nil))
		return
	}

	if req.Image == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: image", nil))
		return
	}

	if req.Subnet == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: subnet", nil))
		return
	}

	if req.Sgs == nil || len(req.Sgs) < 1 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: sgs", nil))
		return
	}

	if req.CpuCredits != nil && *req.CpuCredits != "standard" && *req.CpuCredits != "unlimited" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "invalid value for cpu_credits: must be standard or unlimited", nil))
		return
	}

	policy, err := instanceCreatePolicy()
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

	out, err := orch.createInstance(r.Context(), &req)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

func (s *server) InstanceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	policy, err := instanceDeletePolicy(id)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to generate policy", err))
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

	if err := orch.deleteInstance(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, nil)
}

func (s *server) InstanceListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
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

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
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

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
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

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
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

func (s *server) InstanceGetCommandHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	instance_id := vars["id"]
	cmd_id := vars["cid"]

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

	out, err := service.GetCommandInvocation(r.Context(), instance_id, cmd_id)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, toSSMGetCommandInvocationOutput(out))
}

func (s *server) DescribeAssociationHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	instanceId := vars["id"]
	doc := vars["doc"]

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

	out, err := service.DescribeAssociation(r.Context(), instanceId, doc)
	if err != nil {
		handleError(w, err)
		return
	}
	handleResponseOk(w, toSSMAssociationDescription(out))
}

func (s *server) InstanceStateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	req := &Ec2InstanceStateChangeRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into change power input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.State == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: state", nil))
		return
	}

	policy, err := changeInstanceStatePolicy()
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

	if err := orch.instancesState(r.Context(), req.State, id); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) InstanceSendCommandHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	req := SsmCommandRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := fmt.Sprintf("cannot decode body into ssm send command input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.DocumentName == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "DocumentName is required", nil))
		return

	}

	if len(req.Parameters) == 0 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "Parameters are required", nil))
		return
	}
	policy, err := sendCommandPolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	orch, err := s.newSSMOrchestrator(r.Context(), &sessionParams{
		role:         fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		inlinePolicy: policy,
		policyArns: []string{
			"arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess",
		},
	})
	if err != nil {
		handleError(w, err)
		return
	}

	out, err := orch.sendInstancesCommand(r.Context(), &req, id)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)

}

func (s *server) NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *server) InstanceSSMAssociationHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	instanceId := vars["id"]

	req := &SSMAssociationRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into ssm create input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.Document == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "Document is mandatory", nil))
		return
	}

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)
	policy, err := ssmAssociationPolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		policy,
		"arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess",
		"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
	)
	if err != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, apierror.New(apierror.ErrForbidden, msg, err))
		return
	}

	service := ssm.New(
		ssm.WithSession(session.Session),
	)

	out, err := service.CreateAssociation(r.Context(), instanceId, req.Document)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, struct{ AssociationId string }{AssociationId: *out.AssociationDescription.AssociationId})
}

func (s *server) InstanceUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	instanceId := vars["id"]

	req := &Ec2InstanceUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into update image input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if len(req.Tags) == 0 && len(req.InstanceType) == 0 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required fields: tags or instance_type", nil))
		return
	} else if len(req.Tags) > 0 && len(req.InstanceType) > 0 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "only one of these fields should be provided: tags or instance_type", nil))
		return
	}

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)
	policy, err := instanceUpdatePolicy()
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

	if len(req.Tags) > 0 {
		if err := service.UpdateTags(r.Context(), req.Tags, instanceId); err != nil {
			handleError(w, err)
			return
		}
	} else if len(req.InstanceType) > 0 {
		if err := service.UpdateAttributes(r.Context(), req.InstanceType["value"], instanceId); err != nil {
			handleError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
