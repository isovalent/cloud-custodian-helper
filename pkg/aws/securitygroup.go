package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"
)

func deleteSecurityGroups(ctx context.Context, client *ec2.Client, vpcId string, securityGroups []types.SecurityGroup) (errs error) {
	for _, securityGroup := range securityGroups {
		if securityGroup.GroupId == nil {
			continue
		}
		if securityGroup.VpcId == nil || *securityGroup.VpcId != vpcId {
			continue
		}
		groupId := *securityGroup.GroupId
		securityGroupRules, err := listSecurityGroupRules(ctx, client, groupId)
		if err == nil && len(securityGroupRules) > 0 {
			if err := deleteSecurityGroupRules(ctx, client, groupId, securityGroupRules); err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
		}
		_, err = client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: securityGroup.GroupId,
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listNonDefaultSecurityGroups(ctx context.Context, client *ec2.Client, vpcId string) ([]types.SecurityGroup, error) {
	input := ec2.DescribeSecurityGroupsInput{
		Filters: ec2VpcFilter(vpcId),
	}
	var securityGroups []types.SecurityGroup
	for {
		output, err := client.DescribeSecurityGroups(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, securityGroup := range output.SecurityGroups {
			if securityGroup.GroupName != nil && *securityGroup.GroupName == "default" {
				continue
			}
			securityGroups = append(securityGroups, securityGroup)
		}
		if output.NextToken == nil {
			return securityGroups, nil
		}
		input.NextToken = output.NextToken
	}
}
