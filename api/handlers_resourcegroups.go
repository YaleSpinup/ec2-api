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
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/YaleSpinup/apierror"
	rg2 "github.com/YaleSpinup/ec2-api/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
)

// ResourceGroupsCreateHandler handles the creation of a new resource group
func (s *server) ResourceGroupsCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	account := vars["account"]

	log.Debugf("Creating resource group in account: %s", account)

	// Initialize the resource groups service for this account
	if err := s.setResourceGroupsService(account); err != nil {
		log.Errorf("Failed to initialize resource groups service: %v", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to initialize resource groups service", err))
		return
	}

	// Updated input struct to match AWS SDK expectations
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Query       struct {
			Type  string `json:"Type"`
			Query string `json:"Query"`
		} `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "invalid json input", err))
		return
	}

	if input.Name == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "name is required", nil))
		return
	}

	// Convert the input to AWS SDK format
	createInput := rg2.CreateGroupInput{
		Name:        input.Name,
		Description: input.Description,
		ResourceQuery: &resourcegroups.ResourceQuery{
			Type:  &input.Query.Type,
			Query: &input.Query.Query,
		},
	}

	group, err := s.resourceGroups.CreateGroup(createInput)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to create resource group", err))
		return
	}

	handleResponseOk(w, group)
}

// ResourceGroupsListHandler lists all resource groups
func (s *server) ResourceGroupsListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	account := vars["account"]

	log.Debugf("Listing resource groups in account: %s", account)

	if err := s.setResourceGroupsService(account); err != nil {
		log.Errorf("Failed to initialize resource groups service: %v", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to initialize resource groups service", err))
		return
	}

	groups, err := s.resourceGroups.ListGroups()
	if err != nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to list resource groups", err))
		return
	}

	log.Infof("Successfully listed resource groups for account: %s", account)
	handleResponseOk(w, groups)
}

// ResourceGroupsGetHandler gets details of a specific resource group and its resources
func (s *server) ResourceGroupsGetHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	account := vars["account"]
	groupName := vars["id"]

	log.Debugf("Getting resource group %s in account: %s", groupName, account)

	if err := s.setResourceGroupsService(account); err != nil {
		log.Errorf("Failed to initialize resource groups service: %v", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to initialize resource groups service", err))
		return
	}

	// Get group details
	group, err := s.resourceGroups.GetGroup(groupName)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to get resource group", err))
		return
	}

	// Get resources in the group
	resources, err := s.resourceGroups.ListGroupResources(groupName)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to list group resources", err))
		return
	}

	// Combine group details and resources
	response := struct {
		Group     *resourcegroups.Group `json:"Group"`
		Resources interface{}           `json:"Resources"`
	}{
		Group:     group,
		Resources: resources,
	}

	log.Infof("Successfully retrieved resource group: %s", groupName)
	handleResponseOk(w, response)
}

// ResourceGroupsDeleteHandler handles deletion of a resource group
func (s *server) ResourceGroupsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	account := vars["account"]
	groupName := vars["id"]

	log.Debugf("Deleting resource group %s in account: %s", groupName, account)

	if err := s.setResourceGroupsService(account); err != nil {
		log.Errorf("Failed to initialize resource groups service: %v", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to initialize resource groups service", err))
		return
	}

	if err := s.resourceGroups.DeleteGroup(groupName); err != nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "failed to delete resource group", err))
		return
	}

	log.Infof("Successfully deleted resource group: %s", groupName)
	handleResponseOk(w, nil)
}
