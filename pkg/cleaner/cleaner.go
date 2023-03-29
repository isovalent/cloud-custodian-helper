package cleaner

import (
	"c7n-helper/pkg/aws"
	"c7n-helper/pkg/dto"
	"context"
	"errors"
	"log"
	"strings"
	"time"
)

func Clean(resourceFile string, tries int, retryInterval time.Duration) error {
	log.Println("Reading resource file...")
	var report dto.PolicyReport
	if err := report.ReadFromFile(resourceFile); err != nil {
		return err
	}
	if strings.ToLower(report.Type) != "eks" {
		return errors.New("unsupported resource type")
	}
	ctx := context.Background()
	log.Println("Preparing AWS clients...")
	if err := aws.InitClientsMap(ctx, report.Accounts); err != nil {
		return err
	}
	log.Println("Starting resources deletion...")
	if err := aws.DeleteClusters(ctx, report.Accounts, tries, retryInterval); err != nil {
		return err
	}
	log.Println("Finished successful!")
	return nil
}
