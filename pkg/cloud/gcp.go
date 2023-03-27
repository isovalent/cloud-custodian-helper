package cloud

import (
	"c7n-helper/pkg/converter"
	"c7n-helper/pkg/dto"
	"encoding/json"
	"time"
)

type GKE struct {
	Name      string    `converter:"name"`
	CreatedAt time.Time `converter:"createTime"`
}

type GCE struct {
	Name       string    `converter:"name"`
	LaunchTime time.Time `converter:"creationTimestamp"`
}

func GkeFromFile(file string) ([]dto.Resource, error) {
	content, err := converter.JsonToBytes(file)
	if err != nil {
		return nil, err
	}
	var clusters []GKE
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

func GceFromFile(file string) ([]dto.Resource, error) {
	content, err := converter.JsonToBytes(file)
	if err != nil {
		return nil, err
	}
	var vms []GCE
	if err := json.Unmarshal(content, &vms); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(vms))
	for _, ec2 := range vms {
		result = append(result, dto.Resource{
			Name:    ec2.Name,
			Created: ec2.LaunchTime,
		})
	}
	return result, nil
}
