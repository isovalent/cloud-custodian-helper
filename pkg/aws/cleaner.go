package aws

import (
	"context"
	"errors"
	"time"

	"c7n-helper/pkg/dto"
	"c7n-helper/pkg/log"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/multierr"
)

func DeleteResources(ctx context.Context, accounts []dto.Account, tries int, retryInterval time.Duration) error {
	wg := multierror.Group{}
	for _, account := range accounts {
		for _, resource := range account.Resources {
			key := clientKey(account.Name, resource.Location)
			cls := clientsMap[key]
			clusterName := resource.Name
			wg.Go(func() error {
				ctx, logger := log.UpdateContext(ctx, "account:region", key, "eks", clusterName)
				logger.Info("finding cluster and vpc")
				cluster, err := listEKS(ctx, cls.EKS, clusterName)
				if err != nil {
					if errors.As(err, &eksNotFoundErr) {
						logger.Info("cluster not found, probably it was deleted previously")
						return nil
					}
					return err
				}
				vpcID := *cluster.ResourcesVpcConfig.VpcId
				ctx, logger = log.UpdateContext(ctx, "vpc", vpcID)
				for try := 1; try <= tries; try++ {
					if try > 1 {
						logger.Warnf("delete failed, will retry after sleep: %s", err.Error())
						time.Sleep(retryInterval)
					}
					logger.Infof("starting delete process [attempt: %d]", try)
					if err = deleteVpcAndEks(ctx, cls, vpcID, clusterName); err == nil {
						break
					}
				}
				return err
			})
		}
	}
	return wg.Wait().ErrorOrNil()
}

func deleteVpcAndEks(ctx context.Context, clients *clients, vpcID, clusterName string) error {
	logger := log.FromContext(ctx)
	var errs error

	logger.Info("deleting cluster")
	if err := deleteEKS(ctx, clients.EKS, clusterName); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing vpc peering connections")
	connections, err := listVpcPeeringConnections(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting vpc peering connections: %d", len(connections))
	if err := deleteVpcPeeringConnections(ctx, clients.EC2, vpcID, connections); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing load balancers")
	balancers, err := listLoadBalancers(ctx, clients.ELB, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting load balancers: %d", len(balancers))

	if err := deleteLoadBalancers(ctx, clients.ELB, balancers); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing load balancers V2")
	balancersV2, err := listLoadBalancersV2(ctx, clients.ELBv2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting load balancers V2: %d", len(balancersV2))

	if err := deleteLoadBalancersV2(ctx, clients.ELBv2, balancersV2); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing autoscaling groups")
	scalingGroups, err := listAutoScalingGroups(ctx, clients.ASG, clusterName)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting autoscaling groups: %d", len(scalingGroups))
	if err := deleteAutoScalingGroups(ctx, clients.ASG, clients.EC2, scalingGroups); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing reservation")
	reservations, err := listReservations(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting reservation: %d", len(reservations))
	if err := terminateInstancesInReservations(ctx, clients.EC2, reservations); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing network acl")
	acls, err := listNonDefaultNetworkAcls(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting network acl: %d", len(acls))
	if err := deleteNetworkAcls(ctx, clients.EC2, vpcID, acls); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing elastic ips")
	ips, err := listElasticIps(ctx, clients.EC2, clusterName)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting elastic ips: %d", len(ips))
	if err := releaseElasticIps(ctx, clients.EC2, ips); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing nat gateways")
	nats, err := listNatGateways(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting nat gateways: %d", len(nats))
	if err := deleteNatGateways(ctx, clients.EC2, nats); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing internet gateways")
	gws, err := listInternetGateways(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting internet gateways: %d", len(gws))
	if err := deleteInternetGateways(ctx, clients.EC2, vpcID, gws); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing network interfaces")
	interfaces, err := listNetworkInterfaces(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting network interfaces: %d", len(interfaces))
	if err := deleteNetworkInterfaces(ctx, clients.EC2, interfaces); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing subnets")
	subnets, err := listSubnets(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting subnets: %d", len(subnets))
	if err := deleteSubnets(ctx, clients.EC2, vpcID, subnets); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing security groups")
	secGroups, err := listNonDefaultSecurityGroups(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting security groups: %d", len(secGroups))
	if err := deleteSecurityGroups(ctx, clients.EC2, vpcID, secGroups); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing vpn gateways")
	vpns, err := listVpnGateways(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting vpn gateways: %d", len(vpns))
	if err := deleteVpnGateways(ctx, clients.EC2, vpcID, vpns); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("listing route tables")
	routes, err := listRouteTables(ctx, clients.EC2, vpcID)
	if err != nil {
		return multierr.Append(errs, err)
	}
	logger.Infof("deleting route tables: %d", len(routes))
	if err := deleteRouteTables(ctx, clients.EC2, vpcID, routes); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("deleting cluster node groups")
	if err := deleteClusterNodeGroups(ctx, clients.EKS, clusterName); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("deleting vpc")
	if err := deleteVpc(ctx, clients.EC2, vpcID); err != nil {
		errs = multierr.Append(errs, err)
	}

	logger.Info("deleting cloud formation")
	if _, err := listCloudFormationStacks(ctx, clients.CF, clusterName); err != nil {
		return multierr.Append(errs, err)
	}
	if err := deleteCloudFormation(ctx, clients.CF, clusterName); err != nil {
		errs = multierr.Append(errs, err)
	}
	return errs
}
