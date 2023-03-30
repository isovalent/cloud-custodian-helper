package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"
)

func deleteInternetGateways(ctx context.Context, client *ec2.Client, vpcId string, internetGateways []types.InternetGateway) (errs error) {
	for _, internetGateway := range internetGateways {
		if internetGateway.InternetGatewayId == nil {
			continue
		}
		// Detach the InternetGateway from the VPC.
		var internetGatewayErrs error
		for _, internetGatewayAttachment := range internetGateway.Attachments {
			state := internetGatewayAttachment.State
			if state == types.AttachmentStatusDetaching || state == types.AttachmentStatusDetached {
				continue
			}
			if internetGatewayAttachment.VpcId == nil || *internetGatewayAttachment.VpcId != vpcId {
				continue
			}
			_, err := client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
				InternetGatewayId: internetGateway.InternetGatewayId,
				VpcId:             internetGatewayAttachment.VpcId,
			})
			internetGatewayErrs = multierr.Append(internetGatewayErrs, err)
		}
		errs = multierr.Append(errs, internetGatewayErrs)
		if internetGatewayErrs != nil {
			continue
		}
		// Delete the InternetGateway.
		_, err := client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: internetGateway.InternetGatewayId,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listInternetGateways(ctx context.Context, client *ec2.Client, vpcId string) ([]types.InternetGateway, error) {
	input := ec2.DescribeInternetGatewaysInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []string{vpcId},
			},
		},
	}
	var internetGateways []types.InternetGateway
	for {
		output, err := client.DescribeInternetGateways(ctx, &input)
		if err != nil {
			return nil, err
		}
		internetGateways = append(internetGateways, output.InternetGateways...)
		if output.NextToken == nil {
			return internetGateways, nil
		}
		input.NextToken = output.NextToken
	}
}
