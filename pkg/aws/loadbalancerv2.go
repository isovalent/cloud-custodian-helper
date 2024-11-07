package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"go.uber.org/multierr"
)

func deleteLoadBalancersV2(ctx context.Context, client *elasticloadbalancingv2.Client, loadBalancers []types.LoadBalancer) (errs error) {
	for _, loadBalancer := range loadBalancers {
		if loadBalancer.LoadBalancerArn == nil {
			continue
		}
		_, err := client.DeleteLoadBalancer(ctx, &elasticloadbalancingv2.DeleteLoadBalancerInput{
			LoadBalancerArn: loadBalancer.LoadBalancerArn,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listLoadBalancersV2(ctx context.Context, client *elasticloadbalancingv2.Client, vpcId string) ([]types.LoadBalancer, error) {
	input := elasticloadbalancingv2.DescribeLoadBalancersInput{}
	var loadBalancers []types.LoadBalancer
	for {
		output, err := client.DescribeLoadBalancers(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, loadBalancer := range output.LoadBalancers {
			if loadBalancer.VpcId == nil || *loadBalancer.VpcId != vpcId {
				continue
			}
			fmt.Printf("\nfound LB: %s\n", *loadBalancer.LoadBalancerName)
			loadBalancers = append(loadBalancers, loadBalancer)
		}
		if output.NextMarker == nil {
			return loadBalancers, nil
		}
		input.Marker = output.NextMarker
	}
}
