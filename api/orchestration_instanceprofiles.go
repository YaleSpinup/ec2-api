package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/YaleSpinup/ec2-api/common"
	pEc2 "github.com/YaleSpinup/ec2-api/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func getAssociatedBucketTagValue(instance *ec2.Instance) *string {
	var out string

	for _, tag := range instance.Tags {
		if *tag.Key == "AssociatedBucket" {
			out = *tag.Value
		}
	}

	return &out
}

// PolicyStatement is an individual IAM Policy statement
type PolicyStatement struct {
	Effect    string
	Action    []string
	Resource  []string            `json:",omitempty"`
	Principal map[string][]string `json:",omitempty"`
}

// PolicyDoc collects the policy statements
type PolicyDoc struct {
	Version   string
	Statement []PolicyStatement
}

// assumeRolePolicy defines the IAM policy for assuming a role
func (o *iamOrchestrator) assumeRolePolicy() ([]byte, error) {
	policyDoc, err := json.Marshal(PolicyDoc{
		Version: "2012-10-17",
		Statement: []PolicyStatement{
			PolicyStatement{
				Effect: "Allow",
				Action: []string{"sts:AssumeRole"},
				Principal: map[string][]string{
					"Service": {"ec2.amazonaws.com"},
				},
			},
		}})
	if err != nil {
		log.Errorf("failed to generate assume role policy: %s", err)
		return []byte{}, err
	}

	log.Debugf("generated assume role policy with document %s", string(policyDoc))

	return policyDoc, nil
}

// createRole creates a role and instance profile for an instance to access a data set
// returns a slice of functions to perform rollback of its actions
func (o *iamOrchestrator) createRole(ctx context.Context, roleName, instanceID string) ([]func() error, error) {
	var rollBackTasks []func() error

	log.Debugf("creating role %s", roleName)

	roleDoc, err := o.assumeRolePolicy()
	if err != nil {
		return rollBackTasks, common.ErrCode("failed to generate IAM assume role policy", err)
	}

	var roleOutput *iam.CreateRoleOutput
	if roleOutput, err = o.iamClient.Service.CreateRoleWithContext(ctx, &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(roleDoc)),
		Description:              aws.String(fmt.Sprintf("Role for instance %s", instanceID)),
		Path:                     aws.String("/"),
		RoleName:                 aws.String(roleName),
	}); err != nil {
		return rollBackTasks, common.ErrCode("failed to create IAM role "+roleName, err)
	}

	// append role delete to rollback tasks
	rollBackTasks = append(rollBackTasks, func() error {
		return func() error {
			log.Debug("DeleteRoleWithContext")
			if _, err := o.iamClient.Service.DeleteRoleWithContext(ctx, &iam.DeleteRoleInput{RoleName: roleOutput.Role.RoleName}); err != nil {
				return err
			}
			return nil
		}()
	})

	log.Debugf("creating instance profile %s", roleName)

	var instanceProfileOutput *iam.CreateInstanceProfileOutput
	if instanceProfileOutput, err = o.iamClient.Service.CreateInstanceProfileWithContext(ctx, &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(roleName),
		Path:                aws.String("/"),
	}); err != nil {
		return rollBackTasks, common.ErrCode("failed to create instance profile "+roleName, err)
	}

	// append instance profile delete to rollback tasks
	rollBackTasks = append(rollBackTasks, func() error {
		return func() error {
			log.Debug("DeleteInstanceProfileWithContext")
			if _, err := o.iamClient.Service.DeleteInstanceProfileWithContext(ctx, &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String(roleName)}); err != nil {
				return err
			}
			return nil
		}()
	})

	log.Debugf("adding role to instance profile %s", roleName)

	if _, err = o.iamClient.Service.AddRoleToInstanceProfileWithContext(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(roleName),
		RoleName:            aws.String(roleName),
	}); err != nil {
		return rollBackTasks, common.ErrCode("failed to add role to instance profile "+roleName, err)
	}

	// append role removal from instance profile to rollback tasks
	rollBackTasks = append(rollBackTasks, func() error {
		return func() error {
			log.Debug("RemoveRoleFromInstanceProfileWithContext")
			if _, err := o.iamClient.Service.RemoveRoleFromInstanceProfileWithContext(ctx, &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: aws.String(roleName),
				RoleName:            aws.String(roleName),
			}); err != nil {
				return err
			}
			return nil
		}()
	})

	log.Debugf("created instance profile: %s", aws.StringValue(instanceProfileOutput.InstanceProfile.Arn))

	return rollBackTasks, nil
}

// copyInstanceProfile copy the role policies from a given role on an instance, and add the required spinup plus bucket policies
func (o *iamOrchestrator) copyInstanceProfile(ctx context.Context, ec2Service *pEc2.Ec2, instanceID string, roleName string, account string) (*Ec2InstanceProfile, error) {
	var instanceRoleAssociated bool
	var rollBackTasks []func() error
	var newRoleName string

	prof, err := o.getInstanceProfile(ctx, roleName)
	if err != nil {
		return nil, common.ErrCode("failed to get information about instance profile "+roleName, err)
	}

	// we describe the given instance so we can
	// 1) make sure it exists, and 2) see if it already has an instance profile association
	instanceInfo, instanceErr := ec2Service.GetInstance(ctx, instanceID)
	if instanceErr != nil {
		return nil, instanceErr
	}

	// setup rollback function list and defer execution
	defer func() error {
		if err != nil {
			log.Errorf("recovering from error granting access to s3datarepository: %s, executing %d rollback tasks", err, len(rollBackTasks))
			rollBackE(&rollBackTasks)
		}
		return nil
	}()

	// if the instance already has some other instance profile, we copy all policies
	// from the existing profile into the new one and then disassociate the old profile
	if instanceInfo.IamInstanceProfile != nil {
		bucket := getAssociatedBucketTagValue(instanceInfo)
		if bucket == nil {
			return nil, common.ErrCode("failed to get associated bucket tag value", nil)
		}
		// the instance role name is programmatically determined
		// it is equivalent to the instance profile name
		newRoleName = fmt.Sprintf("%s_%s", roleName, *bucket)
		log.Debugf("new role name: %s", newRoleName)

		// we only have the instance profile arn, so let's extract the name
		ipArns := strings.Split(aws.StringValue(instanceInfo.IamInstanceProfile.Arn), "/")
		currentInstanceProfileName := ipArns[len(ipArns)-1]
		if currentInstanceProfileName == newRoleName {
			instanceRoleAssociated = true
		}
		log.Printf("ip arns: %v+", ipArns)
		log.Printf("current instance profile: %v", currentInstanceProfileName)
		log.Printf("instance role associated: %v", instanceRoleAssociated)

		bucketPolicy, inlineBucketPolicyErr := inlineBucketAccessPolicy(*bucket)
		bucketPolicyName := fmt.Sprintf("%s-%s-policy", roleName, *bucket)
		if inlineBucketPolicyErr != nil {
			return nil, common.ErrCode("failed to get inline bucket access policy "+*bucket, err)
		}
		log.Printf("bucket policy: %v", bucketPolicy)
		log.Printf("bucket policy name: %v", bucketPolicyName)

		var neededPolicy iam.Policy
		hasBucketPolicy, hasBucketPolicyErr := o.iamClient.Service.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(fmt.Sprintf("arn:aws:iam::%s:policy/%s", account, bucketPolicyName)),
		})
		log.Debugf("bucket policy: %v", hasBucketPolicy)
		log.Debugf("bucket policy err: %v", hasBucketPolicyErr)
		if hasBucketPolicyErr != nil || hasBucketPolicy == nil {
			createdBucketPolicy, createPolicyErr := o.iamClient.Service.CreatePolicy(&iam.CreatePolicyInput{
				PolicyName:     aws.String(bucketPolicyName),
				PolicyDocument: aws.String(bucketPolicy),
				Path:           aws.String("/"),
			})
			if createPolicyErr != nil {
				return nil, common.ErrCode("failed to create bucket policy", createPolicyErr)
			}
			log.Printf("created bucket policy: %v", *createdBucketPolicy.Policy)
			neededPolicy = *createdBucketPolicy.Policy
		} else {
			neededPolicy = *hasBucketPolicy.Policy
		}

		roleExists := true
		if _, roleExistsErr := o.iamClient.Service.GetRoleWithContext(ctx, &iam.GetRoleInput{RoleName: aws.String(newRoleName)}); roleExistsErr != nil {
			var aerr awserr.Error
			if errors.As(roleExistsErr, &aerr) {
				if aerr.Code() == iam.ErrCodeNoSuchEntityException {
					roleExists = false
					log.Debugf("role %s does not exist", newRoleName)
				} else {
					return nil, common.ErrCode("failed to get IAM role "+newRoleName, roleExistsErr)
				}
			}
		} else {
			log.Debugf("role %s already exists", newRoleName)
		}

		// if there's no existing role for this instance we'll create one and associate it with an
		// instance profile with the same name
		// we also assume that if the role exists, its corresponding instance profile already exists,
		// which will be true unless manually modified outside of this api
		log.Debugf("role exists: %v", roleExists)
		if !roleExists {
			createRoleRollback, err := o.createRole(ctx, newRoleName, instanceID)
			rollBackTasks = append(rollBackTasks, createRoleRollback...)
			if err != nil {
				return nil, err
			}
		}

		log.Debugf("attaching policy %s to role %s", aws.StringValue(neededPolicy.Arn), newRoleName)

		_, err = o.iamClient.Service.AttachRolePolicyWithContext(ctx, &iam.AttachRolePolicyInput{
			PolicyArn: neededPolicy.Arn,
			RoleName:  aws.String(newRoleName),
		})
		if err != nil {
			return nil, common.ErrCode("failed to attach policy "+*neededPolicy.Arn+" to role "+newRoleName, err)
		}

		// append policy detach from role to rollback tasks
		rollBackTasks = append(rollBackTasks, func() error {
			return func() error {
				log.Debug("DetachRolePolicyWithContext")
				if _, err := o.iamClient.Service.DetachRolePolicyWithContext(ctx, &iam.DetachRolePolicyInput{
					PolicyArn: neededPolicy.Arn,
					RoleName:  aws.String(newRoleName),
				}); err != nil {
					return err
				}
				return nil
			}()
		})

		if !instanceRoleAssociated {
			log.Infof("instance %s already has instance profile %s, will try to migrate existing policies", instanceID, currentInstanceProfileName)

			// find out what role(s) correspond to this instance profile and what policies are attached to them
			var ipOut *iam.GetInstanceProfileOutput
			if ipOut, err = o.iamClient.Service.GetInstanceProfileWithContext(ctx, &iam.GetInstanceProfileInput{
				InstanceProfileName: aws.String(currentInstanceProfileName),
			}); err != nil {
				return nil, common.ErrCode("failed to get information about current instance profile "+currentInstanceProfileName, err)
			}

			// TODO: we are _not_ considering role inline policies at this point, should we?
			var currentPoliciesArn []string
			for _, r := range ipOut.InstanceProfile.Roles {
				log.Debugf("listing attached policies for role %s", aws.StringValue(r.RoleName))

				var attachedRolePoliciesOut *iam.ListAttachedRolePoliciesOutput
				if attachedRolePoliciesOut, err = o.iamClient.Service.ListAttachedRolePoliciesWithContext(ctx, &iam.ListAttachedRolePoliciesInput{
					RoleName: r.RoleName,
				}); err != nil {
					return nil, common.ErrCode("failed to list attached policies for role "+aws.StringValue(r.RoleName), err)
				}

				if attachedRolePoliciesOut.AttachedPolicies == nil {
					log.Warnf("no attached policies found for current role %s, there may be inline policies", aws.StringValue(r.RoleName))
				}

				for _, p := range attachedRolePoliciesOut.AttachedPolicies {
					currentPoliciesArn = append(currentPoliciesArn, aws.StringValue(p.PolicyArn))
				}
			}

			log.Infof("policies attached to the current instance profile: %s", currentPoliciesArn)

			//attach current policies to our new role
			for _, p := range currentPoliciesArn {
				log.Debugf("attaching pre-existing policy %s to role %s", p, newRoleName)

				_, err = o.iamClient.Service.AttachRolePolicyWithContext(ctx, &iam.AttachRolePolicyInput{
					PolicyArn: aws.String(p),
					RoleName:  aws.String(newRoleName),
				})
				if err != nil {
					return nil, common.ErrCode("failed to attach policy "+p+" to role "+newRoleName, err)
				}
			}

			// append policy detach from role to rollback tasks
			rollBackTasks = append(rollBackTasks, func() error {
				return func() error {
					for _, p := range currentPoliciesArn {
						log.Debugf("DetachRolePolicyWithContext: %s (%s)", p, newRoleName)
						err = retry(3, 3*time.Second, func() error {
							_, err = o.iamClient.Service.DetachRolePolicyWithContext(ctx, &iam.DetachRolePolicyInput{
								PolicyArn: aws.String(p),
								RoleName:  aws.String(newRoleName),
							})
							if err != nil {
								log.Debugf("retrying, got error: %s", err)
								return err
							}
							return nil
						})
						if err != nil {
							log.Warnf("failed to detach policy "+p+" from role "+newRoleName, err)
							continue
						}
					}
					return nil
				}()
			})

			// find out the association id for the currently associated instance profile
			var ipAssociationsOut *ec2.DescribeIamInstanceProfileAssociationsOutput
			if ipAssociationsOut, err = ec2Service.Service.DescribeIamInstanceProfileAssociationsWithContext(ctx, &ec2.DescribeIamInstanceProfileAssociationsInput{
				Filters: []*ec2.Filter{
					{
						Name:   aws.String("instance-id"),
						Values: []*string{aws.String(instanceID)},
					},
					{
						Name:   aws.String("state"),
						Values: []*string{aws.String("associated")},
					},
				},
			}); err != nil {
				return nil, common.ErrCode("failed to describe instance profile associations for instance "+instanceID, err)
			}

			log.Debugf("got associations: %+v", ipAssociationsOut.IamInstanceProfileAssociations)

			if len(ipAssociationsOut.IamInstanceProfileAssociations) != 1 {
				return nil, common.ErrCode("did not find exactly 1 instance profile association for instance "+instanceID, nil)
			}

			log.Debugf("disassociating association id %s", aws.StringValue(ipAssociationsOut.IamInstanceProfileAssociations[0].AssociationId))

			// retry the instance profile disassociation
			err = retry(5, 3*time.Second, func() error {
				_, err = ec2Service.Service.DisassociateIamInstanceProfileWithContext(ctx, &ec2.DisassociateIamInstanceProfileInput{
					AssociationId: ipAssociationsOut.IamInstanceProfileAssociations[0].AssociationId,
				})
				if err != nil {
					log.Debugf("retrying, got error: %s", err)
					return err
				}
				return nil
			})
			if err != nil {
				return nil, common.ErrCode("failed to disassociate current instance profile from instance "+instanceID, err)
			}

			// append original instance profile association to rollback tasks
			rollBackTasks = append(rollBackTasks, func() error {
				return func() error {
					log.Debug("AssociateIamInstanceProfileWithContext")
					err = retry(5, 3*time.Second, func() error {
						_, err = ec2Service.Service.AssociateIamInstanceProfileWithContext(ctx, &ec2.AssociateIamInstanceProfileInput{
							IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
								Arn: instanceInfo.IamInstanceProfile.Arn,
							},
							InstanceId: aws.String(instanceID),
						})
						if err != nil {
							log.Debugf("retrying, got error: %s", err)
							return err
						}
						return nil
					})
					if err != nil {
						return err
					}
					return nil
				}()
			})
		}
	}

	// we associate the new instance profile with the instance, unless it's already associated
	if !instanceRoleAssociated {
		log.Infof("associating instance profile %s with instance %s", newRoleName, instanceID)

		// retry the instance profile association as it takes a while to show up
		err = retry(5, 3*time.Second, func() error {
			_, err = ec2Service.Service.AssociateIamInstanceProfileWithContext(ctx, &ec2.AssociateIamInstanceProfileInput{
				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Name: aws.String(newRoleName),
				},
				InstanceId: aws.String(instanceID),
			})
			if err != nil {
				log.Debugf("retrying, got error: %s", err)
				return err
			}
			return nil
		})
		if err != nil {
			return nil, common.ErrCode("failed to associate instance profile with instance "+instanceID, err)
		}

		log.Debugf("associated instance profile %s with instance %s", newRoleName, instanceID)
	}

	return prof, nil
}

// getInstanceProfile get the instance profile for a given role name and its associated policies and inline policies
func (o *iamOrchestrator) getInstanceProfile(ctx context.Context, name string) (*Ec2InstanceProfile, error) {
	out := &Ec2InstanceProfile{}
	ip, err := o.iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: aws.String(name)})
	if err != nil {
		return nil, err // do not modify this error
	}

	out.Profile = ip

	// detach policies from role(s) and delete the role(s)
	for _, r := range ip.Roles {
		// Get attached policies
		rps, err := o.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: r.RoleName})
		if err != nil {
			return nil, common.ErrCode(fmt.Sprintf("failed to list attached policies for the role %s", aws.StringValue(r.RoleName)), err)
		}

		// assign attached policies to array
		out.AttachedPolicies = rps

		// get inline policies if they exist
		ps, psErr := o.iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{RoleName: r.RoleName})
		if psErr != nil {
			return nil, common.ErrCode("failed to list inline policies", psErr)
		}
		out.InlinePolicies = ps
	}

	// return the filled out struct
	return out, nil
}

// deleteInstanceProfile deletes the specified instance profile and associated role, if they exist
// any policies attached to the role will be detached and left intact
func (o *iamOrchestrator) deleteInstanceProfile(ctx context.Context, name string) error {
	ip, err := o.iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: aws.String(name)})
	if err != nil {
		return err // do not modify this error
	}

	// detach policies from role(s) and delete the role(s)
	for _, r := range ip.Roles {
		// detach all attached policies
		rps, err := o.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: r.RoleName})
		if err != nil {
			return common.ErrCode(fmt.Sprintf("failed to list attached policies for the role %s", aws.StringValue(r.RoleName)), err)
		}
		for _, rp := range rps {
			input := &iam.DetachRolePolicyInput{
				RoleName:  r.RoleName,
				PolicyArn: rp.PolicyArn,
			}
			if err := o.iamClient.DetachRolePolicy(ctx, input); err != nil {
				return common.ErrCode(fmt.Sprintf("failed to detach policy for the role %s", aws.StringValue(r.RoleName)), err)
			}
		}

		// delete all inline policies
		ps, err := o.iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{RoleName: r.RoleName})
		if err != nil {
			return common.ErrCode("failed to list inline policies", err)
		}
		for _, p := range ps {
			input := &iam.DeleteRolePolicyInput{
				RoleName:   r.RoleName,
				PolicyName: p,
			}
			if err := o.iamClient.DeleteRolePolicy(ctx, input); err != nil {
				return common.ErrCode("failed to delete inline policy", err)
			}
		}

		// remove the role from the instance profile
		input := &iam.RemoveRoleFromInstanceProfileInput{
			RoleName:            r.RoleName,
			InstanceProfileName: ip.InstanceProfileName,
		}
		if err := o.iamClient.RemoveRoleFromInstanceProfile(ctx, input); err != nil {
			return common.ErrCode("failed to remove role from instance profile", err)
		}

		// delete the role
		if err := o.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: r.RoleName}); err != nil {
			return common.ErrCode("failed to delete role", err)
		}
	}

	// delete the instance profile
	if err := o.iamClient.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String(aws.StringValue(ip.InstanceProfileName))}); err != nil {
		return common.ErrCode("failed to delete instance profile", err)
	}
	return nil
}
