package parser

import (
	"c7n-helper/pkg/cloud"
	"c7n-helper/pkg/dto"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var resourceParsers = map[string]func(file string) ([]dto.Resource, error){
	"eks": cloud.EksFromFile,
	"ec2": cloud.Ec2FromFile,
	"gke": cloud.GkeFromFile,
	"gce": cloud.GceFromFile,
}

func Parse(resourceType, c7nDir, policy, outFile string) error {
	log.Println("Processing C7N report directory...")
	files, err := resourceFiles(c7nDir, policy)
	if err != nil {
		return err
	}
	log.Println("Parsing C7N resource files...")
	report, err := reportFromFiles(files, resourceType, policy)
	if err != nil {
		return err
	}
	log.Println("Persisting JSON report...")
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

func reportFromFiles(files []string, resourceType, policy string) (dto.PolicyReport, error) {
	accountMap := make(map[string]dto.Account)
	for _, file := range files {
		resources, err := resourcesFromFile(resourceType, file)
		if err != nil {
			return dto.PolicyReport{}, err
		}
		if len(resources) == 0 {
			continue
		}
		accName, region := accountRegion(file)
		account, ok := accountMap[accName]
		if !ok {
			account = dto.Account{
				Name:            accName,
				RegionResources: make(map[string][]dto.Resource),
			}
			accountMap[accName] = account
		}
		if _, ok := account.RegionResources[region]; !ok {
			account.RegionResources[region] = make([]dto.Resource, 0)
		}
		account.RegionResources[region] = append(account.RegionResources[region], resources...)
	}
	return dto.PolicyReport{
		ResourceType: resourceType,
		C7NPolicy:    policy,
		Accounts:     accountsFromMap(accountMap),
	}, nil
}

func resourcesFromFile(resourceType, file string) ([]dto.Resource, error) {
	parser, ok := resourceParsers[resourceType]
	if !ok {
		return nil, errors.New("unsupported resource type")
	}
	return parser(file)
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
