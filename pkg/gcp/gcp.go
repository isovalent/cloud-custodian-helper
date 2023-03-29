package gcp

import (
	"c7n-helper/pkg/dto"
	"encoding/json"
	"strings"
	"time"
)

func GKE(_ string, content []byte) ([]dto.Resource, error) {
	var clusters []struct {
		Name      string    `json:"name"`
		Location  string    `json:"location"`
		CreatedAt time.Time `json:"createTime"`
	}
	if err := json.Unmarshal(content, &clusters); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, dto.Resource{
			Name:     cluster.Name,
			Location: cluster.Location,
			Created:  cluster.CreatedAt,
		})
	}
	return result, nil
}

func GCE(_ string, content []byte) ([]dto.Resource, error) {
	var vms []struct {
		Name       string    `json:"name"`
		Zone       string    `json:"zone"`
		LaunchTime time.Time `json:"creationTimestamp"`
	}
	if err := json.Unmarshal(content, &vms); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(vms))
	for _, vm := range vms {
		result = append(result, dto.Resource{
			Name:     vm.Name,
			Location: normalizeZone(vm.Zone),
			Created:  vm.LaunchTime,
		})
	}
	return result, nil
}

// Zone value: `https://www.googleapis.com/compute/v1/projects/<project-name>/zones/us-central1-a`
func normalizeZone(zone string) string {
	parts := strings.Split(zone, "/")
	if len(parts) < 2 {
		return zone
	}
	return parts[len(parts)-1]
}
