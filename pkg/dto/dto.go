package dto

import (
	"encoding/json"
	"os"
	"time"
)

type PolicyReport struct {
	ResourceType string    `json:"resourceType"`
	C7NPolicy    string    `json:"policyName"`
	Accounts     []Account `json:"accounts"`
}

type Account struct {
	Name            string                `json:"name"`
	RegionResources map[string][]Resource `json:"regionResources"`
}

type Resource struct {
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
}

func (r *PolicyReport) ReadFromFile(reportFile string) error {
	file, err := os.ReadFile(reportFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(file, r); err != nil {
		return err
	}
	return nil
}
