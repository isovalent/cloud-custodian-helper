package aws

import (
	"c7n-helper/pkg/dto"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"time"
)

var eksNotFoundErr *types.ResourceNotFoundException

func ParseEKS(region string, content []byte) ([]dto.Resource, error) {
	var clusters []struct {
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}
	if err := json.Unmarshal(content, &clusters); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, dto.Resource{
			Name:     cluster.Name,
			Location: region,
			Created:  cluster.CreatedAt,
		})
	}
	return result, nil
}

func listEKS(ctx context.Context, client *eks.Client, clusterName string) (*types.Cluster, error) {
	res, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return nil, err
	}
	return res.Cluster, nil
}

func deleteEKS(ctx context.Context, client *eks.Client, clusterName string) error {
	_, err := client.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil && !errors.As(err, &eksNotFoundErr) {
		return err
	}
	return nil
}
