package dto

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type PolicyReport struct {
	Type     string    `json:"type"`
	Policy   string    `json:"policy"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	Name      string     `json:"name"`
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Name     string    `json:"name"`
	Location string    `json:"location"`
	Owner    string    `json:"owner"`
	Created  time.Time `json:"created"`
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

func (r *PolicyReport) String() string {
	return fmt.Sprintf("%s report with %d accounts", r.Type, len(r.Accounts))
}
