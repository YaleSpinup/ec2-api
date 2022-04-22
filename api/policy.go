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
	log.Debugf("generating tag crete policy document")

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

func volumeCreatePolicy() (string, error) {
	log.Debugf("generating volume crete policy document")

	policy := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:CreateVolume",
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
