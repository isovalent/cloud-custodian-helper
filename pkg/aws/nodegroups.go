package aws

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

func deleteClusterNodeGroups(ctx context.Context, client *eks.Client, clusterName string) error {
	nodeGroups, err := listClusterNodeGroups(ctx, client, clusterName)
	if err != nil {
		return err
	}
	for _, group := range nodeGroups {
		_, err = client.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: &group,
		})
		if err != nil && !errors.As(err, &eksNotFoundErr) {
			return err
		}
	}
	return err
}

func listClusterNodeGroups(ctx context.Context, client *eks.Client, clusterName string) ([]string, error) {
	var nextToken *string
	result := make([]string, 0)
	for {
		groups, err := client.ListNodegroups(ctx, &eks.ListNodegroupsInput{
			ClusterName: aws.String(clusterName),
			NextToken:   nextToken,
		})
		if err != nil {
			if errors.As(err, &eksNotFoundErr) {
				return nil, nil
			}
			return nil, err
		}
		result = append(result, groups.Nodegroups...)
		if len(groups.Nodegroups) == 0 || groups.NextToken == nil {
			break
		}
		nextToken = groups.NextToken
	}
	return result, nil
}
