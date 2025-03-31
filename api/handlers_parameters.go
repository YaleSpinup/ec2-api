package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ParameterCreateHandler handles creating SSM parameters
func (s *server) ParameterCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	log.Infof("creating parameter in account %s", account)

	req := SSMParameterCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := fmt.Sprintf("cannot decode body into create parameter input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	// Validate required fields
	if req.Name == nil || aws.StringValue(req.Name) == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: name", nil))
		return
	}

	if req.Type == nil || aws.StringValue(req.Type) == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: type", nil))
		return
	}

	if req.Value == nil || aws.StringValue(req.Value) == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: value", nil))
		return
	}

	// Validate parameter type
	paramType := aws.StringValue(req.Type)
	if paramType != "String" && paramType != "StringList" && paramType != "SecureString" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "type must be one of: String, StringList, SecureString", nil))
		return
	}

	// Setup policy and get orchestrator
	policy, err := ssmParameterCreatePolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	// Create inline policy with limited scope
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	o, err := s.newSSMOrchestrator(r.Context(), &sessionParams{
		role:         role,
		inlinePolicy: policy,
	})
	if err != nil {
		msg := fmt.Sprintf("failed to get ssm orchestrator for account: %s", account)
		handleError(w, apierror.New(apierror.ErrInternalError, msg, err))
		return
	}

	// Check if parameter should be force updated (overwritten)
	forceUpdate := false
	if req.Overwrite != nil && *req.Overwrite {
		forceUpdate = true
	}

	var out *SSMParameterResponse

	if forceUpdate {
		// Use update orchestrator for forced update
		out, err = o.updateParameter(r.Context(), &req)
	} else {
		// Use create orchestrator (will fail if parameter exists)
		out, err = o.createParameter(r.Context(), &req)
	}

	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

// ParameterUpdateHandler handles updating SSM parameters
func (s *server) ParameterUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

	log.Infof("updating parameter %s in account %s", name, account)

	req := SSMParameterCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := fmt.Sprintf("cannot decode body into update parameter input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	// Override name from URL path
	req.Name = aws.String(name)

	// Validate required fields
	if req.Type == nil || aws.StringValue(req.Type) == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: type", nil))
		return
	}

	if req.Value == nil || aws.StringValue(req.Value) == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "missing required field: value", nil))
		return
	}

	// Validate parameter type
	paramType := aws.StringValue(req.Type)
	if paramType != "String" && paramType != "StringList" && paramType != "SecureString" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "type must be one of: String, StringList, SecureString", nil))
		return
	}

	// Setup policy and get orchestrator
	policy, err := ssmParameterUpdatePolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	// Create inline policy with limited scope
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	o, err := s.newSSMOrchestrator(r.Context(), &sessionParams{
		role:         role,
		inlinePolicy: policy,
	})
	if err != nil {
		msg := fmt.Sprintf("failed to get ssm orchestrator for account: %s", account)
		handleError(w, apierror.New(apierror.ErrInternalError, msg, err))
		return
	}

	// Update parameter through orchestrator
	out, err := o.updateParameter(r.Context(), &req)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

// ParameterGetHandler handles retrieving SSM parameters
func (s *server) ParameterGetHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

	log.Infof("getting parameter %s in account %s", name, account)

	// Get withDecryption query parameter
	decrypt := false
	decryptStr := r.URL.Query().Get("decrypt")
	if decryptStr != "" {
		var err error
		decrypt, err = strconv.ParseBool(decryptStr)
		if err != nil {
			handleError(w, apierror.New(apierror.ErrBadRequest, "invalid decrypt parameter, must be true or false", err))
			return
		}
	}

	// Validate input
	if name == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "parameter name is required", nil))
		return
	}

	// Setup policy and get orchestrator
	policy, err := ssmParameterReadPolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	// If decryption is requested, add KMS permissions
	if decrypt {
		policy, err = ssmParameterReadDecryptPolicy()
		if err != nil {
			handleError(w, err)
			return
		}
	}

	// Create inline policy with limited scope
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	o, err := s.newSSMOrchestrator(r.Context(), &sessionParams{
		role:         role,
		inlinePolicy: policy,
	})
	if err != nil {
		msg := fmt.Sprintf("failed to get ssm orchestrator for account: %s", account)
		handleError(w, apierror.New(apierror.ErrInternalError, msg, err))
		return
	}

	// Get parameter through orchestrator
	out, err := o.getParameter(r.Context(), name, decrypt)
	if err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, out)
}

// ParameterListHandler handles listing SSM parameters
func (s *server) ParameterListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])

	log.Infof("listing parameters in account %s", account)

	// Extract query parameters for filtering
	filters := make(map[string]string)

	// Name filter
	if name := r.URL.Query().Get("name"); name != "" {
		filters["Name"] = name
	}

	// Type filter
	if paramType := r.URL.Query().Get("type"); paramType != "" {
		filters["Type"] = paramType
	}

	// Path filter - allows filtering by parameter path
	if path := r.URL.Query().Get("path"); path != "" {
		filters["Path"] = path
	}

	// MaxResults parameter
	var maxResults int64
	if maxStr := r.URL.Query().Get("max_results"); maxStr != "" {
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil {
			handleError(w, apierror.New(apierror.ErrBadRequest, "invalid max_results parameter", err))
			return
		}
		maxResults = max
	}

	// NextToken parameter
	nextToken := r.URL.Query().Get("next_token")

	// Setup policy and get orchestrator
	policy, err := ssmParameterReadPolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	// Create inline policy with limited scope
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	o, err := s.newSSMOrchestrator(r.Context(), &sessionParams{
		role:         role,
		inlinePolicy: policy,
	})
	if err != nil {
		msg := fmt.Sprintf("failed to get ssm orchestrator for account: %s", account)
		handleError(w, apierror.New(apierror.ErrInternalError, msg, err))
		return
	}

	// List parameters through orchestrator
	parameters, token, err := o.listParameters(r.Context(), filters, maxResults, nextToken)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return parameters with next token if available
	response := map[string]interface{}{
		"parameters": parameters,
	}

	if token != "" {
		response["next_token"] = token
	}

	handleResponseOk(w, response)
}

// ParameterDeleteHandler handles deleting SSM parameters
func (s *server) ParameterDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := s.mapAccountNumber(vars["account"])
	name := vars["name"]

	log.Infof("deleting parameter %s in account %s", name, account)

	// Validate input
	if name == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "parameter name is required", nil))
		return
	}

	// Setup policy and get orchestrator
	policy, err := ssmParameterDeletePolicy()
	if err != nil {
		handleError(w, err)
		return
	}

	// Create inline policy with limited scope
	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	o, err := s.newSSMOrchestrator(r.Context(), &sessionParams{
		role:         role,
		inlinePolicy: policy,
	})
	if err != nil {
		msg := fmt.Sprintf("failed to get ssm orchestrator for account: %s", account)
		handleError(w, apierror.New(apierror.ErrInternalError, msg, err))
		return
	}

	// Delete parameter through orchestrator
	if err := o.deleteParameter(r.Context(), name); err != nil {
		handleError(w, err)
		return
	}

	handleResponseOk(w, map[string]interface{}{
		"message": fmt.Sprintf("parameter %s deleted", name),
	})
}

// Policy for creating SSM parameters
func ssmParameterCreatePolicy() (string, error) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ssm:PutParameter",
					"ssm:GetParameter",
					"ssm:DescribeParameters",
					"ssm:AddTagsToResource",
					"kms:GenerateDataKey",
					"kms:Decrypt"
				],
				"Resource": "*"
			}
		]
	}`

	return policy, nil
}

// Policy for updating SSM parameters
func ssmParameterUpdatePolicy() (string, error) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ssm:PutParameter",
					"ssm:GetParameter",
					"ssm:DescribeParameters",
					"ssm:AddTagsToResource",
					"ssm:RemoveTagsFromResource",
					"kms:GenerateDataKey",
					"kms:Decrypt"
				],
				"Resource": "*"
			}
		]
	}`

	return policy, nil
}

// Policy for reading SSM parameters
func ssmParameterReadPolicy() (string, error) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ssm:GetParameter",
					"ssm:GetParameters",
					"ssm:DescribeParameters",
					"ssm:ListTagsForResource"
				],
				"Resource": "*"
			}
		]
	}`

	return policy, nil
}

// Policy for reading and decrypting SSM parameters
func ssmParameterReadDecryptPolicy() (string, error) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ssm:GetParameter",
					"ssm:GetParameters",
					"ssm:DescribeParameters",
					"ssm:ListTagsForResource",
					"kms:Decrypt"
				],
				"Resource": "*"
			}
		]
	}`

	return policy, nil
}

// Policy for deleting SSM parameters
func ssmParameterDeletePolicy() (string, error) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ssm:DeleteParameter",
					"ssm:RemoveTagsFromResource"
				],
				"Resource": "*"
			}
		]
	}`

	return policy, nil
}
