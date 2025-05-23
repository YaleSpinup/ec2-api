package api

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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

	// TODO an api should be for one org, currently we need to support the entire account
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

func GetQueryArrayValues(r *http.Request, key string) ([]string, error) {
	var values []string

	queryValues := r.URL.Query()

	v, ok := queryValues["azs"]
	if !ok {
		log.Error("Missing 'azs' parameter")
		return []string{}, errors.New("missing 'azs' parameter")
	}

	// loop through the values
	for i, value := range v {
		log.Debug(fmt.Printf("Value %d: %s\n", i, value))
		values = append(values, value)
	}

	return values, nil
}

func (s *server) InstanceListTypeOfferings(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	azs, err := GetQueryArrayValues(r, "azs")
	if err != nil {
		handleError(w, err)
	}

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

	perPage := 1000
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

	out, next, err := service.ListInstanceTypeOfferings(r.Context(), azs, int64(perPage), pageToken)
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

	handleResponseOk(w, out)
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

	policy, err := generatePolicy([]string{"ec2:CreateTags", "ec2:ModifyInstanceAttribute"})
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

	if len(req.Tags) > 0 {
		if err := orch.updateInstanceTags(r.Context(), req.Tags, instanceId); err != nil {
			handleError(w, err)
			return
		}
	} else if len(req.InstanceType) > 0 {
		if _, ok := req.InstanceType["value"]; !ok {
			handleError(w, apierror.New(apierror.ErrBadRequest, "missing instance_type value", nil))
			return
		}
		if err := orch.updateInstanceType(r.Context(), req.InstanceType["value"], instanceId); err != nil {
			handleError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) VolumeDetachHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	instance_id := vars["id"]
	volume_id := vars["vid"]
	var force bool
	if r.URL.Query().Has("force") {
		var err error
		force, err = strconv.ParseBool(r.URL.Query().Get("force"))
		if err != nil {
			handleError(w, apierror.New(apierror.ErrBadRequest, "invalid value for force parameter", nil))
			return
		}
	}

	policy, err := generatePolicy([]string{"ec2:DetachVolume"})
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

	out, err := orch.detachVolume(r.Context(), instance_id, volume_id, force)
	if err != nil {
		handleError(w, err)
		return
	}
	handleResponseOk(w, out)
}

func (s *server) VolumeAttachHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	req := &Ec2VolumeAttachmentRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into attach volume input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.Device == nil || req.VolumeID == nil || req.DeleteOnTermination == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required fields: device, volume_id and delete_on_termination", nil))
		return
	}

	policy, err := generatePolicy([]string{"ec2:AttachVolume", "ec2:ModifyInstanceAttribute"})
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

	out, err := orch.attachVolume(r.Context(), req, id)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

func (s *server) InstanceProfileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

	policy, err := generatePolicy([]string{
		"iam:GetInstanceProfile",
		"iam:ListAttachedRolePolicies",
		"iam:DetachRolePolicy",
		"iam:ListRolePolicies",
		"iam:DeleteRolePolicy",
		"iam:RemoveRoleFromInstanceProfile",
		"iam:DeleteRole",
		"iam:DeleteInstanceProfile",
	})
	if err != nil {
		handleError(w, err)
		return
	}

	orch, err := s.newIAMOrchestrator(r.Context(), &sessionParams{
		role:         fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		inlinePolicy: policy,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	err = orch.deleteInstanceProfile(r.Context(), name)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, nil)
}

type Ec2InstanceProfileCopyRequest struct {
	InstanceID string
}

func (s *server) InstanceProfileCopyHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

	policy, err := instanceProfileCopyPolicy()

	req := Ec2InstanceProfileCopyRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := fmt.Sprintf("cannot decode body into iam profile copy input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	ec2Orch, err := s.newEc2Orchestrator(r.Context(), &sessionParams{
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

	iamOrch, err := s.newIAMOrchestrator(r.Context(), &sessionParams{
		role:         fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		inlinePolicy: policy,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	ip, ipErr := iamOrch.copyInstanceProfile(r.Context(), ec2Orch.ec2Client, req.InstanceID, name, account)
	if ipErr != nil {
		handleError(w, ipErr)
		return
	}

	handleResponseOk(w, ip)
}

func (s *server) InstanceProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

	policy, err := generatePolicy([]string{
		"iam:GetInstanceProfile",
		"iam:ListAttachedRolePolicies",
		"iam:ListRolePolicies",
	})
	if err != nil {
		handleError(w, err)
		return
	}

	orch, err := s.newIAMOrchestrator(r.Context(), &sessionParams{
		role:         fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		inlinePolicy: policy,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	ip, ipErr := orch.getInstanceProfile(r.Context(), name)
	if ipErr != nil {
		handleError(w, ipErr)
		return
	}

	handleResponseOk(w, ip)
}
