package cleaner

import (
	"context"
	"errors"
	"strings"
	"time"

	"c7n-helper/pkg/aws"
	"c7n-helper/pkg/dto"
	"c7n-helper/pkg/log"
)

func Clean(ctx context.Context, resourceFile string, tries int, retryInterval time.Duration) error {
	logger := log.FromContext(ctx)
	logger.Info("reading resource file...")
	var report dto.PolicyReport
	if err := report.ReadFromFile(resourceFile); err != nil {
		return err
	}
	if strings.ToLower(report.Type) != "eks" {
		return errors.New("unsupported resource type")
	}
	logger.Info("preparing aws clients...")
	if err := aws.InitClientsMap(ctx, report.Accounts); err != nil {
		return err
	}
	logger.Info("starting resources cleanup...")
	if err := aws.DeleteResources(ctx, report.Accounts, tries, retryInterval); err != nil {
		return err
	}
	logger.Info("finished successful")
	return nil
}
