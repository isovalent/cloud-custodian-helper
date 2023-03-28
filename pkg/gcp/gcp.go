package gcp

import (
	"c7n-helper/pkg/dto"
	"encoding/json"
	"time"
)

type gke struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createTime"`
}

type gce struct {
	Name       string    `json:"name"`
	LaunchTime time.Time `json:"creationTimestamp"`
}

func GKE(content []byte) ([]dto.Resource, error) {
	var clusters []gke
	if err := json.Unmarshal(content, &clusters); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, dto.Resource{
			Name:    cluster.Name,
			Created: cluster.CreatedAt,
		})
	}
	return result, nil
}

func GCE(content []byte) ([]dto.Resource, error) {
	var vms []gce
	if err := json.Unmarshal(content, &vms); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(vms))
	for _, vm := range vms {
		result = append(result, dto.Resource{
			Name:    vm.Name,
			Created: vm.LaunchTime,
		})
	}
	return result, nil
}
