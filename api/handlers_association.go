package api

import (
	"fmt"
	"net/http"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ssm"
	"github.com/gorilla/mux"
)

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
