package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"
)

func deleteNetworkAcls(ctx context.Context, client *ec2.Client, vpcId string, networkAcls []types.NetworkAcl) (errs error) {
	for _, networkAcl := range networkAcls {
		if networkAcl.NetworkAclId == nil {
			continue
		}
		if networkAcl.VpcId == nil || *networkAcl.VpcId != vpcId {
			continue
		}

		_, err := client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: networkAcl.NetworkAclId,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listNonDefaultNetworkAcls(ctx context.Context, client *ec2.Client, vpcId string) ([]types.NetworkAcl, error) {
	input := ec2.DescribeNetworkAclsInput{
		Filters: ec2VpcFilter(vpcId),
	}
	var networkAcls []types.NetworkAcl
	for {
		output, err := client.DescribeNetworkAcls(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, networkAcl := range output.NetworkAcls {
			if networkAcl.IsDefault == nil || !*networkAcl.IsDefault {
				networkAcls = append(networkAcls, networkAcl)
			}
		}
		if output.NextToken == nil {
			return networkAcls, nil
		}
		input.NextToken = output.NextToken
	}
}
