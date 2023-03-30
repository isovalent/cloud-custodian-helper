package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"go.uber.org/multierr"
)

const (
	instanceTerminatedWaiterMaxDuration = time.Minute * 10
	instanceTerminatedRetryMinDelay     = time.Second * 5
	instanceTerminatedRetryMaxDelay     = time.Second * 15
)

func deleteAutoScalingGroups(ctx context.Context, client *autoscaling.Client, ec2Client *ec2.Client, autoScalingGroups []types.AutoScalingGroup) (errs error) {
	for _, autoScalingGroup := range autoScalingGroups {
		if autoScalingGroup.AutoScalingGroupName == nil {
			continue
		}
		// Resize the AutoScalingGroup to zero if not already zero.
		if (autoScalingGroup.DesiredCapacity != nil && *autoScalingGroup.DesiredCapacity != 0) ||
			(autoScalingGroup.MaxSize != nil && *autoScalingGroup.MaxSize != 0) ||
			(autoScalingGroup.MinSize != nil && *autoScalingGroup.MinSize != 0) {
			_, err := client.UpdateAutoScalingGroup(ctx, &autoscaling.UpdateAutoScalingGroupInput{
				AutoScalingGroupName: autoScalingGroup.AutoScalingGroupName,
				DesiredCapacity:      aws.Int32(0),
				MaxSize:              aws.Int32(0),
				MinSize:              aws.Int32(0),
			})
			errs = multierr.Append(errs, err)
		}
		// Wait for any Instances to terminate.
		instanceIds := make([]string, 0, len(autoScalingGroup.Instances))
		for _, instance := range autoScalingGroup.Instances {
			if instance.InstanceId != nil {
				instanceIds = append(instanceIds, *instance.InstanceId)
			}
		}
		if len(instanceIds) > 0 {
			instanceTerminatedWaiter := ec2.NewInstanceTerminatedWaiter(ec2Client)
			err := instanceTerminatedWaiter.Wait(
				ctx,
				&ec2.DescribeInstancesInput{InstanceIds: instanceIds},
				instanceTerminatedWaiterMaxDuration,
				func(options *ec2.InstanceTerminatedWaiterOptions) {
					options.MinDelay = instanceTerminatedRetryMinDelay
					options.MaxDelay = instanceTerminatedRetryMaxDelay
				})
			errs = multierr.Append(errs, err)
			if err != nil {
				continue
			}
		}

		// Delete the AutoScalingGroup.
		_, err := client.DeleteAutoScalingGroup(ctx, &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: autoScalingGroup.AutoScalingGroupName,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listAutoScalingGroups(ctx context.Context, client *autoscaling.Client, clusterName string) ([]types.AutoScalingGroup, error) {
	autoScalingGroups := make([]types.AutoScalingGroup, 0)
	groupNames := map[string]struct{}{}
	for _, filter := range autoScalingFilters(clusterName) {
		groups, err := describeAutoScalingGroups(ctx, client, filter)
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			if _, ok := groupNames[*group.AutoScalingGroupARN]; ok {
				continue
			}
			autoScalingGroups = append(autoScalingGroups, group)
			groupNames[*group.AutoScalingGroupARN] = struct{}{}
		}
	}
	return autoScalingGroups, nil
}

func autoScalingFilters(clusterName string) [][]types.Filter {
	filters := make([][]types.Filter, 0)
	if clusterName != "" {
		filters = append(filters, []types.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: []string{"k8s.io/cluster-autoscaler/" + clusterName, "kubernetes.io/cluster/" + clusterName, "k8s.io/cluster/" + clusterName},
			},
			{
				Name:   aws.String("tag-value"),
				Values: []string{"owned"},
			},
		})
		filters = append(filters, []types.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: []string{"eks:cluster-name"},
			},
			{
				Name:   aws.String("tag-value"),
				Values: []string{clusterName},
			},
		})
	}
	return filters
}

func describeAutoScalingGroups(ctx context.Context, client *autoscaling.Client, filters []types.Filter) ([]types.AutoScalingGroup, error) {
	var autoScalingGroups []types.AutoScalingGroup
	for {
		input := &autoscaling.DescribeAutoScalingGroupsInput{
			Filters: filters,
		}
		output, err := client.DescribeAutoScalingGroups(ctx, input)
		if err != nil {
			return nil, err
		}
		autoScalingGroups = append(autoScalingGroups, output.AutoScalingGroups...)
		if output.NextToken == nil {
			return autoScalingGroups, nil
		}
		input.NextToken = output.NextToken
	}
}
