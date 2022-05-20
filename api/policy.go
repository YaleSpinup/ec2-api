package api

import (
	"encoding/json"
	"fmt"

	"github.com/YaleSpinup/aws-go/services/iam"

	log "github.com/sirupsen/logrus"
)

// orgTagAccessPolicy generates the org tag conditional policy to be passed inline when assuming a role
func orgTagAccessPolicy(org string) (string, error) {
	log.Debugf("generating org policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   []string{"*"},
				Resource: []string{"*"},
				Condition: iam.Condition{
					"StringEquals": iam.ConditionStatement{
						"aws:ResourceTag/spinup:org": []string{org},
					},
				},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func instanceCreatePolicy() (string, error) {
	log.Debugf("generating instance crete policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:RunInstances",
					"iam:PassRole",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func instanceDeletePolicy(id string) (string, error) {
	log.Debugf("generating instance delete policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:TerminateInstances",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func sgDeletePolicy(id string) (string, error) {
	log.Debugf("generating sg delete policy document")

	sgResource := fmt.Sprintf("arn:aws:ec2:*:*:security-group/%s", id)

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:DeleteSecurityGroup",
				},
				Resource: []string{sgResource},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func sgCreatePolicy() (string, error) {
	log.Debugf("generating sg crete policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:CreateSecurityGroup",
					"ec2:CreateTags",
					"ec2:ModifySecurityGroupRules",
					"ec2:DeleteSecurityGroup",
					"ec2:AuthorizeSecurityGroupEgress",
					"ec2:AuthorizeSecurityGroupIngress",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func sgUpdatePolicy(id string) (string, error) {
	log.Debugf("generating sg crete policy document")

	sgResource := fmt.Sprintf("arn:aws:ec2:*:*:security-group/%s", id)

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:ModifySecurityGroupRules",
					"ec2:AuthorizeSecurityGroupEgress",
					"ec2:AuthorizeSecurityGroupIngress",
					"ec2:RevokeSecurityGroupEgress",
					"ec2:RevokeSecurityGroupIngress",
				},
				Resource: []string{sgResource},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func tagCreatePolicy() (string, error) {
	log.Debugf("generating tag create policy document")
	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:CreateTags",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func generatePolicy(actions []string) (string, error) {
	log.Debugf("generating %v policy document", actions)

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   actions,
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func volumeDeletePolicy(id string) (string, error) {
	log.Debugf("generating volume delete policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:DeleteVolume",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func changeInstanceStatePolicy() (string, error) {
	log.Debugf("generating power update policy document")
	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:StartInstances",
					"ec2:StopInstances",
					"ec2:RebootInstances",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func sendCommandPolicy() (string, error) {
	log.Debugf("generating send command policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ssm:SendCommand",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func instanceUpdatePolicy() (string, error) {
	log.Debugf("generating tag create policy document")
	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:CreateTags",
					"ec2:ModifyInstanceAttribute",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func ssmAssociationPolicy() (string, error) {
	log.Debugf("generating tag create policy document")
	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ssm:CreateAssociation",
					"ssm:UpdateAssociation",
				},
				Resource: []string{"*"},
			},
		},
	}

	j, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(j), nil
}
