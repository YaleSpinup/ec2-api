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
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/YaleSpinup/ec2-api/resourcegroups"
	log "github.com/sirupsen/logrus"
)

func (s *server) resourceGroupsServiceForAccount(account string) (*resourcegroups.ResourceGroups, error) {
	log.Debugf("getting resourcegroups service for account %s", account)

	accountNumber := s.mapAccountNumber(account)
	if accountNumber == "" {
		return nil, fmt.Errorf("account number not found for %s", account)
	}

	// Construct the role ARN for the target account
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/SpinupPlusXAManagementRoleTst", accountNumber)
	log.Debugf("assuming role: %s", roleArn)

	// Use the existing session to create STS client
	stsClient := sts.New(s.session.Session)

	// Create the assume role input
	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("resource-groups-session"),
	}

	// Add ExternalId if it exists
	if s.session.ExternalID != "" {
		assumeRoleInput.ExternalId = aws.String(s.session.ExternalID)
	}

	// Assume the role
	roleOutput, err := stsClient.AssumeRole(assumeRoleInput)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %v", err)
	}

	// Create a new session with the assumed role credentials
	assumedSession := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			*roleOutput.Credentials.AccessKeyId,
			*roleOutput.Credentials.SecretAccessKey,
			*roleOutput.Credentials.SessionToken,
		),
		Region: aws.String("us-east-1"), // or get this from config
	}))

	return resourcegroups.New(resourcegroups.WithSession(assumedSession)), nil
}

// setResourceGroupsService sets the resourcegroups service in the server
func (s *server) setResourceGroupsService(account string) error {
	rg, err := s.resourceGroupsServiceForAccount(account)
	if err != nil {
		return err
	}

	s.resourceGroups = rg
	return nil
}
