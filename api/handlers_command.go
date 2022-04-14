package api

import (
	"fmt"
	"net/http"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ssm"
	"github.com/gorilla/mux"
)

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
