package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) UpdateAttributes(ctx context.Context, input *ec2.ModifyInstanceAttributeInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("updating attributes with input %+v", *input)

	if _, err := e.Service.ModifyInstanceAttributeWithContext(ctx, input); err != nil {
		return common.ErrCode("updating attributes", err)
	}

	return nil
}
