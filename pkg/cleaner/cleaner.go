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
		for _, resource := range account.Resources {
			log.Printf("Deleteing %s %s in %s [%s] ...\n", strings.ToUpper(report.Type), resource.Name, account.Name, resource.Location)
			//TODO: implement me
		}
	}
	return nil
}
