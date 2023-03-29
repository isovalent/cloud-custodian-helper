package aws

import (
	"c7n-helper/pkg/dto"
	"context"
	"github.com/hashicorp/go-multierror"
	"log"
	"time"
)

func DeleteClusters(ctx context.Context, accounts []dto.Account, tries int, retryInterval time.Duration) error {
	wg := multierror.Group{}
	for _, account := range accounts {
		for _, resource := range account.Resources {
			cl := clientsMap[clientKey(account.Name, resource.Location)]
			clusterName := resource.Name
			wg.Go(func() error {
				var err error
				for try := 0; try < tries; try++ {
					if err = deleteEKS(ctx, cl, clusterName); err == nil {
						break
					}
					log.Printf("EKS %s delete failed, will retry after sleep...\n", clusterName)
					time.Sleep(retryInterval)
					log.Printf("Retrying EKS %s deletion...\n", clusterName)
				}
				return err

			})
		}
	}
	return wg.Wait().ErrorOrNil()
}

func deleteEKS(ctx context.Context, clients *clients, clusterName string) error {
	log.Printf("Deleting EKS %s...\n", clusterName)
	//TODO: implement me
	return nil
}
