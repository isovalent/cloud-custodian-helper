package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"
)

func releaseElasticIps(ctx context.Context, client *ec2.Client, addresses []types.Address) (errs error) {
	for _, address := range addresses {
		_, err := client.ReleaseAddress(ctx, &ec2.ReleaseAddressInput{
			AllocationId: address.AllocationId,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listElasticIps(ctx context.Context, client *ec2.Client, clusterName string) ([]types.Address, error) {
	filters := []types.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []string{clusterName + "*"},
		},
	}
	output, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}
	return output.Addresses, nil
}
