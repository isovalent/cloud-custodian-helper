package aws

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"go.uber.org/multierr"
)

func deleteVpc(ctx context.Context, client *ec2.Client, vpcId string) error {
	_, err := client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcId),
	})
	if apiErr := (*smithy.GenericAPIError)(nil); errors.As(err, &apiErr) {
		if apiErr.ErrorCode() == "InvalidVpcID.NotFound" {
			return nil
		}
	}
	return err
}

func deleteVpcPeeringConnections(ctx context.Context, client *ec2.Client, vpcId string, vpcPeeringConnections []types.VpcPeeringConnection) (errs error) {
	for _, vpcPeeringConnection := range vpcPeeringConnections {
		if vpcPeeringConnection.VpcPeeringConnectionId == nil {
			continue
		}
		isAccepter := vpcPeeringConnection.AccepterVpcInfo != nil &&
			vpcPeeringConnection.AccepterVpcInfo.VpcId != nil &&
			*vpcPeeringConnection.AccepterVpcInfo.VpcId == vpcId
		isRequester := vpcPeeringConnection.RequesterVpcInfo != nil &&
			vpcPeeringConnection.RequesterVpcInfo.VpcId != nil &&
			*vpcPeeringConnection.RequesterVpcInfo.VpcId == vpcId
		if !isAccepter && !isRequester {
			continue
		}

		_, err := client.DeleteVpcPeeringConnection(ctx, &ec2.DeleteVpcPeeringConnectionInput{
			VpcPeeringConnectionId: vpcPeeringConnection.VpcPeeringConnectionId,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listVpcPeeringConnections(ctx context.Context, client *ec2.Client, vpcId string) ([]types.VpcPeeringConnection, error) {
	var connections []types.VpcPeeringConnection
AccepterRequester:
	for _, name := range []string{"accepter-vpc-info.vpc-id", "requester-vpc-info.vpc-id"} {
		req := ec2.DescribeVpcPeeringConnectionsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String(name),
					Values: []string{vpcId},
				},
			},
		}
		for {
			res, err := client.DescribeVpcPeeringConnections(ctx, &req)
			if err != nil {
				return nil, err
			}
			connections = append(connections, res.VpcPeeringConnections...)
			if res.NextToken == nil {
				continue AccepterRequester
			}
			req.NextToken = res.NextToken
		}
	}
	return connections, nil
}

func ec2VpcFilter(vpcId string) []types.Filter {
	return []types.Filter{
		{
			Name:   aws.String("vpc-id"),
			Values: []string{vpcId},
		},
	}
}
