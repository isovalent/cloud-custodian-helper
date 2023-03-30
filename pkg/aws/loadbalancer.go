package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"go.uber.org/multierr"
)

func deleteLoadBalancers(ctx context.Context, client *elasticloadbalancing.Client, loadBalancerDescriptions []types.LoadBalancerDescription) (errs error) {
	for _, loadBalancerDescription := range loadBalancerDescriptions {
		if loadBalancerDescription.LoadBalancerName == nil {
			continue
		}
		_, err := client.DeleteLoadBalancer(ctx, &elasticloadbalancing.DeleteLoadBalancerInput{
			LoadBalancerName: loadBalancerDescription.LoadBalancerName,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listLoadBalancers(ctx context.Context, client *elasticloadbalancing.Client, vpcId string) ([]types.LoadBalancerDescription, error) {
	input := elasticloadbalancing.DescribeLoadBalancersInput{}
	var loadBalancerDescriptions []types.LoadBalancerDescription
	for {
		output, err := client.DescribeLoadBalancers(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, loadBalancerDescription := range output.LoadBalancerDescriptions {
			if loadBalancerDescription.VPCId == nil || *loadBalancerDescription.VPCId != vpcId {
				continue
			}
			loadBalancerDescriptions = append(loadBalancerDescriptions, loadBalancerDescription)
		}
		if output.NextMarker == nil {
			return loadBalancerDescriptions, nil
		}
		input.Marker = output.NextMarker
	}
}
