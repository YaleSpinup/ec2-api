package api

import (
	"fmt"
	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/ec2"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (s *server) SecurityGroupListHandler(w http.ResponseWriter, r *http.Request) {
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

	out, err := service.ListSecurityGroups(r.Context(), "", name)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("X-Items", strconv.Itoa(len(out)))

	handleResponseOk(w, out)
}
