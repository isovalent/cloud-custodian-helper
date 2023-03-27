package cleaner

import (
	"c7n-helper/pkg/dto"
	"log"
	"strings"
)

func Clean(resourceFile string) error {
	log.Println("Reading resource file...")
	var report dto.PolicyReport
	if err := report.ReadFromFile(resourceFile); err != nil {
		return err
	}
	for _, account := range report.Accounts {
		for region, resources := range account.RegionResources {
			log.Printf("Cleaning %s [%d] in %s [%s] ...\n", strings.ToUpper(report.ResourceType), len(resources), account.Name, region)
			//TODO: implement me
		}
	}
	return nil
}
