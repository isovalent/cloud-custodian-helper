package aws

import (
	"c7n-helper/pkg/dto"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"log"
)

type clients struct {
	Autoscaling *autoscaling.Client
	CF          *cloudformation.Client
	EC2         *ec2.Client
	ELB         *elasticloadbalancing.Client
	EKS         *eks.Client
}

var clientsMap = map[string]*clients{}

func InitClientsMap(ctx context.Context, accounts []dto.Account) error {
	for _, account := range accounts {
		for _, resource := range account.Resources {
			key := clientKey(account.Name, resource.Location)
			if _, ok := clientsMap[key]; !ok {
				log.Printf("Initializing AWS clients for: %s\n", key)
				cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(account.Name), config.WithRegion(resource.Location))
				if err != nil {
					return err
				}
				clientsMap[key] = &clients{
					Autoscaling: autoscaling.NewFromConfig(cfg),
					CF:          cloudformation.NewFromConfig(cfg),
					EC2:         ec2.NewFromConfig(cfg),
					ELB:         elasticloadbalancing.NewFromConfig(cfg),
					EKS:         eks.NewFromConfig(cfg),
				}
			}
		}
	}
	return nil
}

func clientKey(account, region string) string {
	return fmt.Sprintf("%s:%s", account, region)
}
