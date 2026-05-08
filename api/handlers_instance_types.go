package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// InstanceTypeCatalogHandler returns an enriched catalog of all EC2 instance types
// with metadata (vCPU, memory, architecture, GPU) and per-AZ availability.
// Results are cached for 24 hours per account.
func (s *server) InstanceTypeCatalogHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	cacheKey := fmt.Sprintf("instance-type-catalog_%s", account)
	if cached, found := s.dataCache.Get(cacheKey); found {
		log.Infof("returning cached instance type catalog for account %s", account)
		handleResponseOk(w, cached)
		return
	}

	log.Infof("cache miss for instance type catalog, fetching from AWS for account %s", account)

	orch, err := s.newEc2Orchestrator(r.Context(), &sessionParams{
		role:       fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
		policyArns: []string{"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess"},
	})
	if err != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, fmt.Errorf("%s: %w", msg, err))
		return
	}

	catalog, err := orch.getInstanceTypeCatalog(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	s.dataCache.Set(cacheKey, catalog, 24*time.Hour)
	log.Infof("cached instance type catalog for account %s (%d types)", account, len(catalog))

	handleResponseOk(w, catalog)
}
