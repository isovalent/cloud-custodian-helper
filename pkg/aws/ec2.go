package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"c7n-helper/pkg/dto"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func ParseEC2(region string, content []byte) ([]dto.Resource, error) {
	var vms []struct {
		InstanceId   string     `json:"InstanceId"`
		LaunchTime   time.Time  `json:"LaunchTime"`
		InstanceType string     `json:"InstanceType"`
		Tags         []keyValue `json:"Tags"`
	}
	if err := json.Unmarshal(content, &vms); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(vms))
	for _, vm := range vms {
		owner := ""
		for _, tag := range vm.Tags {
			if tag.Key == "owner" {
				owner = tag.Value
				break
			}
		}
		result = append(result, dto.Resource{
			Name:     fmt.Sprintf("%s [%s]", vm.InstanceId, vm.InstanceType),
			Location: region,
			Owner:    owner,
			Created:  vm.LaunchTime,
		})
	}
	return result, nil
}

func listReservations(ctx context.Context, client *ec2.Client, vpcId string) ([]types.Reservation, error) {
	input := ec2.DescribeInstancesInput{
		Filters: ec2VpcFilter(vpcId),
	}
	var reservations []types.Reservation
	for {
		output, err := client.DescribeInstances(ctx, &input)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, output.Reservations...)
		if output.NextToken == nil {
			return reservations, nil
		}
		input.NextToken = output.NextToken
	}
}

func terminateInstancesInReservations(ctx context.Context, client *ec2.Client, reservations []types.Reservation) error {
	// Find all non-terminated Instances.
	var nonTerminatedInstanceIds []string
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId == nil {
				continue
			}
			if instance.State != nil && instance.State.Name == types.InstanceStateNameTerminated {
				continue
			}
			nonTerminatedInstanceIds = append(nonTerminatedInstanceIds, *instance.InstanceId)
		}
	}

	// If all Instances are terminated then we are done.
	if len(nonTerminatedInstanceIds) == 0 {
		return nil
	}

	// Terminate all non-terminated Instances.
	terminateInstancesOutput, err := client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: nonTerminatedInstanceIds,
	})
	if err != nil {
		return err
	}

	// Find all terminating Instances.
	var terminatingInstanceIds []string
	for _, terminatingInstance := range terminateInstancesOutput.TerminatingInstances {
		if terminatingInstance.InstanceId == nil {
			continue
		}
		if terminatingInstance.CurrentState != nil && terminatingInstance.CurrentState.Name == types.InstanceStateNameTerminated {
			continue
		}
		terminatingInstanceIds = append(terminatingInstanceIds, *terminatingInstance.InstanceId)
	}

	// If there are no terminating Instances then we are done.
	if len(terminatingInstanceIds) == 0 {
		return nil
	}

	// Wait for all terminating Instances to terminate.
	instanceTerminatedWaiter := ec2.NewInstanceTerminatedWaiter(client)
	err = instanceTerminatedWaiter.Wait(
		ctx,
		&ec2.DescribeInstancesInput{InstanceIds: terminatingInstanceIds},
		instanceTerminatedWaiterMaxDuration,
		func(options *ec2.InstanceTerminatedWaiterOptions) {
			options.MinDelay = instanceTerminatedRetryMinDelay
			options.MaxDelay = instanceTerminatedRetryMaxDelay
		})
	if err != nil {
		return err
	}

	return nil
}
