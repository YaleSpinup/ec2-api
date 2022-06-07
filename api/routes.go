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
	api := s.router.PathPrefix("/v2/ec2").Subrouter()
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

	api.HandleFunc("/{account}/instances/{id}/ssm/command", s.InstanceGetCommandHandler).Methods(http.MethodGet).Queries("command_id", "{cid}")
	api.HandleFunc("/{account}/instances/{id}/ssm/association", s.DescribeAssociationHandler).Methods(http.MethodGet).Queries("document", "{doc}")

	api.HandleFunc("/{account}/sgs", s.SecurityGroupListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/sgs/{id}", s.SecurityGroupGetHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes", s.VolumeListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes/{id}", s.VolumeGetHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes/{id}/modifications", s.VolumeListModificationsHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/volumes/{id}/snapshots", s.VolumeListSnapshotsHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/snapshots", s.SnapshotListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/snapshots/{id}", s.SnapshotGetHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/subnets", s.SubnetsListHandler).Methods(http.MethodGet).Queries("vpc", "{vpc}")
	api.HandleFunc("/{account}/subnets", s.SubnetsListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/images", s.ImageListHandler).Methods(http.MethodGet).Queries("name", "{name}")
	api.HandleFunc("/{account}/images", s.ImageListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/images/{id}", s.ImageGetHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/vpcs", s.VpcListHandler).Methods(http.MethodGet)
	api.HandleFunc("/{account}/vpcs/{id}", s.VpcShowHandler).Methods(http.MethodGet)

	api.HandleFunc("/{account}/instances", s.InstanceCreateHandler).Methods(http.MethodPost)
	api.HandleFunc("/{account}/instances/{id}/volumes", s.VolumeAttachHandler).Methods(http.MethodPost)
	api.HandleFunc("/{account}/sgs", s.SecurityGroupCreateHandler).Methods(http.MethodPost)
	api.HandleFunc("/{account}/volumes", s.VolumeCreateHandler).Methods(http.MethodPost)
	api.HandleFunc("/{account}/snapshots", s.ProxyRequestHandler).Methods(http.MethodPost)
	api.HandleFunc("/{account}/images", s.ImageCreateHandler).Methods(http.MethodPost)

	api.HandleFunc("/{account}/images/{id}/tags", s.ImageUpdateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}", s.NotImplementedHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/power", s.InstanceStateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/ssm/command", s.InstanceSendCommandHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/ssm/association", s.InstanceSSMAssociationHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/tags", s.InstanceUpdateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/instances/{id}/attribute", s.InstanceUpdateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/sgs/{id}", s.SecurityGroupUpdateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/sgs/{id}/tags", s.SecurityGroupUpdateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/volumes/{id}", s.VolumeUpdateHandler).Methods(http.MethodPut)
	api.HandleFunc("/{account}/volumes/{id}/tags", s.VolumeUpdateHandler).Methods(http.MethodPut)

	api.HandleFunc("/{account}/instances/{id}", s.InstanceDeleteHandler).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/instances/{id}/volumes/{vid}", s.ProxyRequestHandler).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/instanceprofiles/{name}", s.ProxyRequestHandler).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/sgs/{id}", s.SecurityGroupDeleteHandler).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/volumes/{id}", s.VolumeDeleteHandler).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/snapshots/{id}", s.ProxyRequestHandler).Methods(http.MethodDelete)
	api.HandleFunc("/{account}/images/{id}", s.ProxyRequestHandler).Methods(http.MethodDelete)
}
