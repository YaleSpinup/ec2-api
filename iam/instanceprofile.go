package iam

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)


func (i *Iam) GetInstanceProfile(ctx context.Context, name string) (?????, error) {
	inp := iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}

	out, err:= i.Service.GetInstanceProfile(&inp)
	// TODO: continue from here

}
