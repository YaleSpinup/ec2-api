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

func (s *server) SnapshotListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	orch, err := s.newEc2Orchestrator(r.Context(), &sessionParams{
		role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		policyArns: []string{
			"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
		},
	})
	if err != nil {
		handleError(w, err)
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

	out, next, err := orch.listSnapshots(r.Context(), int64(perPage), pageToken)
	if err != nil {
		handleError(w, err)
		return
	}
	list := make([]map[string]*string, len(out))
	for i, s := range out {
		list[i] = map[string]*string{
			"id": s.SnapshotId,
		}
	}

	w.Header().Set("X-Items", strconv.Itoa(len(list)))
	if next != nil {
		w.Header().Set("X-Per-Page", strconv.Itoa(perPage))
		w.Header().Set("X-Next-Token", aws.StringValue(next))
	}

	handleResponseOk(w, list)
}

func (s *server) SnapshotSyncTagHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{ResponseWriter: w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	policy, err := tagCreatePolicy()
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

	out, err := SnapshotsWithoutCOA(r.Context(), orch)
	if err != nil {
		handleError(w, err)
		return
	}

	stats, err := UpdateSnapshotTags(r.Context(), orch, out)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, stats)
}

func (s *server) SnapshotGetHandler(w http.ResponseWriter, r *http.Request) {
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

	out, err := service.GetSnapshot(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	if len(out) == 0 {
		handleError(w, apierror.New(apierror.ErrNotFound, "not found", nil))
		return
	}

	if len(out) > 1 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "unexpected snapshot count returned", nil))
		return
	}

	handleResponseOk(w, toEC2SnapshotResponse(out[0]))
}

func (s *server) SnapshotCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	req := &Ec2SnapshotCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		msg := fmt.Sprintf("cannot decode body into update volume input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}
	if req.VolumeId == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required fields: volume_id", nil))
		return
	}
	if req.CopyTags == nil {
		req.CopyTags = aws.Bool(true)
	}

	policy, err := generatePolicy([]string{"ec2:CreateSnapshot", "ec2:CreateTags"})
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

	out, err := orch.createSnapshot(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

func (s *server) SnapshotDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	id := vars["id"]

	policy, err := generatePolicy([]string{"ec2:DeleteSnapshot"})
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

	err = orch.deleteSnapshot(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, nil)
}
