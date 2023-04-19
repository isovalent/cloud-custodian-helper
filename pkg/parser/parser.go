package parser

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"c7n-helper/pkg/aws"
	"c7n-helper/pkg/azure"
	"c7n-helper/pkg/dto"
	"c7n-helper/pkg/gcp"
	"c7n-helper/pkg/log"
)

var resourceParsers = map[string]func(region string, content []byte) ([]dto.Resource, error){
	"eks": aws.ParseEKS,
	"ec2": aws.ParseEC2,
	"s3":  aws.ParseS3,
	"gke": gcp.GKE,
	"gce": gcp.GCE,
	"arg": azure.RG,
}

func Parse(ctx context.Context, resourceType, c7nDir, policy, outFile string) error {
	logger := log.FromContext(ctx)
	logger.Info("processing c7n report directory...")
	files, err := resourceFiles(c7nDir, policy)
	if err != nil {
		return err
	}
	logger.Info("parsing c7n resource files...")
	report, err := reportFromFiles(ctx, files, resourceType, policy)
	if err != nil {
		return err
	}
	logger.Info("sorting resources...")
	sortResources(report.Accounts)
	logger.Infof("saving %s...", report.String())
	return persistReport(report, outFile)
}

func resourceFiles(c7nDir, policy string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(c7nDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "/"+policy+"/resources.json") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func reportFromFiles(ctx context.Context, files []string, resourceType, policy string) (dto.PolicyReport, error) {
	accountMap := make(map[string]dto.Account)
	for _, file := range files {
		accName, region := accountRegion(file)
		resources, err := resourcesFromFile(ctx, resourceType, region, file)
		if err != nil {
			return dto.PolicyReport{}, err
		}
		if len(resources) == 0 {
			continue
		}
		account, ok := accountMap[accName]
		if !ok {
			account = dto.Account{Name: accName, Resources: make([]dto.Resource, 0)}
		}
		account.Resources = append(account.Resources, resources...)
		accountMap[accName] = account
	}
	return dto.PolicyReport{
		Type:     resourceType,
		Policy:   policy,
		Accounts: accountsFromMap(accountMap),
	}, nil
}

func sortResources(accounts []dto.Account) {
	for i := range accounts {
		acc := accounts[i]
		sort.Slice(acc.Resources, func(i, j int) bool {
			return acc.Resources[i].Created.Before(acc.Resources[j].Created)
		})
	}
}

func resourcesFromFile(ctx context.Context, resourceType, region, file string) ([]dto.Resource, error) {
	parser, ok := resourceParsers[resourceType]
	if !ok {
		return nil, errors.New("unsupported resource type")
	}
	content, err := jsonToBytes(ctx, file)
	if err != nil {
		return nil, err
	}
	return parser(region, content)
}

func accountsFromMap(accountMap map[string]dto.Account) []dto.Account {
	accounts := make([]dto.Account, 0, len(accountMap))
	for _, accRegion := range accountMap {
		accounts = append(accounts, accRegion)
	}
	return accounts
}

func persistReport(report dto.PolicyReport, outFile string) error {
	file, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(outFile, file, 0644)
}

// Parses C7N report path: `.../<account-name|project|subscription>/<region-name|global>/<policy-name>/resources.json`
func accountRegion(file string) (string, string) {
	parts := strings.Split(file, "/")
	l := len(parts)
	return parts[l-4] /* account */, parts[l-3] /* region */
}

func jsonToBytes(ctx context.Context, file string) ([]byte, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			log.FromContext(ctx).Errorf("unable to close json file: %s", err.Error())
		}
	}()
	return io.ReadAll(jsonFile)
}
