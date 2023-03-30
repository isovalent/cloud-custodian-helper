package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"
)

func deleteSecurityGroupRules(ctx context.Context, client *ec2.Client, groupId string, securityGroupRules []types.SecurityGroupRule) (errs error) {
	var egressSecurityGroupRules []types.SecurityGroupRule
	var ingressSecurityGroupRules []types.SecurityGroupRule
	for _, securityGroupRule := range securityGroupRules {
		if securityGroupRule.SecurityGroupRuleId == nil {
			continue
		}
		if securityGroupRule.GroupId == nil || *securityGroupRule.GroupId != groupId {
			continue
		}
		if securityGroupRule.IsEgress == nil || !*securityGroupRule.IsEgress {
			ingressSecurityGroupRules = append(ingressSecurityGroupRules, securityGroupRule)
		} else {
			egressSecurityGroupRules = append(egressSecurityGroupRules, securityGroupRule)
		}
	}
	if len(ingressSecurityGroupRules) > 0 {
		_, err := client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:              aws.String(groupId),
			SecurityGroupRuleIds: securityGroupRuleIds(ingressSecurityGroupRules),
		})
		errs = multierr.Append(errs, err)
	}
	if len(egressSecurityGroupRules) > 0 {
		_, err := client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:              aws.String(groupId),
			SecurityGroupRuleIds: securityGroupRuleIds(egressSecurityGroupRules),
		})
		errs = multierr.Append(errs, err)
	}
	return
}

func listSecurityGroupRules(ctx context.Context, client *ec2.Client, groupId string) ([]types.SecurityGroupRule, error) {
	input := ec2.DescribeSecurityGroupRulesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []string{groupId},
			},
		},
	}
	var securityGroupRules []types.SecurityGroupRule
	for {
		output, err := client.DescribeSecurityGroupRules(ctx, &input)
		if err != nil {
			return nil, err
		}
		securityGroupRules = append(securityGroupRules, output.SecurityGroupRules...)
		if output.NextToken == nil {
			return securityGroupRules, nil
		}
		input.NextToken = output.NextToken
	}
}

func securityGroupRuleIds(securityGroupRules []types.SecurityGroupRule) []string {
	securityGroupRuleIds := make([]string, 0, len(securityGroupRules))
	for _, securityGroupRule := range securityGroupRules {
		if securityGroupRule.SecurityGroupRuleId != nil {
			securityGroupRuleIds = append(securityGroupRuleIds, *securityGroupRule.SecurityGroupRuleId)
		}
	}
	return securityGroupRuleIds
}
