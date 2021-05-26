/*
Copyright © 2021 Yale University

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *server) routes() {
	api := s.router.PathPrefix("/v1/ec2").Subrouter()
	api.HandleFunc("/ping", s.PingHandler).Methods(http.MethodGet)
	api.HandleFunc("/version", s.VersionHandler).Methods(http.MethodGet)
	api.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	api.HandleFunc("/", s.AccountsHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/select", s.InstanceSelectorHandler).Methods(http.MethodGet)

	// instance endpoints
	api.HandleFunc("/{account}/instances", s.InstanceListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/instances/{id}", s.InstanceGetHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/instances/{id}/volumes", s.InstanceVolumesHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/instances/{id}/volumes/{vid}", s.InstanceVolumesHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/instances/{id}/snapshots", s.InstanceListSnapshotsHandler).Methods(http.MethodGet)

	api.HandleFunc("/{account}/instances/{id}/ssm/command", NotImplemented).Methods(http.MethodGet).Queries("command_id", "{cid}")
	api.HandleFunc("/{account}/instances/{id}/ssm/association", NotImplemented).Methods(http.MethodGet)

	api.HandleFunc("/{account}/sgs", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/sgs/{id}", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes/{id}", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes/{id}/modifications", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes/{id}/snapshots", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/snapshots", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/snapshots/{id}", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/subnets", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/images", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/images/{id}", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/vpcs", NotImplemented).Methods(http.MethodGet)
	api.HandleFunc("/{account}/vpcs/{id}", NotImplemented).Methods(http.MethodGet)

	api.HandleFunc("/{account}/instances", NotImplemented).Methods(http.MethodPost)
	api.HandleFunc("/{account}/instances/{id}/volumes", NotImplemented).Methods(http.MethodPost)
	api.HandleFunc("/{account}/sgs", NotImplemented).Methods(http.MethodPost)
	api.HandleFunc("/{account}/volumes", NotImplemented).Methods(http.MethodPost)
	api.HandleFunc("/{account}/snapshots", NotImplemented).Methods(http.MethodPost)
	api.HandleFunc("/{account}/images", NotImplemented).Methods(http.MethodPost)

	api.HandleFunc("/{account}/instances/{id}", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/power", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/ssm/command", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/ssm/association", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/tags", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/attribute", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/sgs/{id}", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/sgs/{id}/tags", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/volumes/{id}", NotImplemented).Methods(http.MethodPut)
	api.HandleFunc("/{account}/volumes/{id}/tags", NotImplemented).Methods(http.MethodPut)

	api.HandleFunc("/{account}/instances/{id}", NotImplemented).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/instances/{id}/volumes/{vid}", NotImplemented).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/sgs/{id}", NotImplemented).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/volumes/{id}", NotImplemented).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/snapshots/{id}", NotImplemented).Methods(http.MethodDelete)
}
