package aws

import (
	"c7n-helper/pkg/dto"
	"encoding/json"
	"fmt"
	"time"
)

func EKS(region string, content []byte) ([]dto.Resource, error) {
	var clusters []struct {
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}
	if err := json.Unmarshal(content, &clusters); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, dto.Resource{
			Name:     cluster.Name,
			Location: region,
			Created:  cluster.CreatedAt,
		})
	}
	return result, nil
}

func EC2(region string, content []byte) ([]dto.Resource, error) {
	var vms []struct {
		InstanceId   string    `json:"InstanceId"`
		LaunchTime   time.Time `json:"LaunchTime"`
		InstanceType string    `json:"InstanceType"`
	}
	if err := json.Unmarshal(content, &vms); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(vms))
	for _, ec2 := range vms {
		result = append(result, dto.Resource{
			Name:     fmt.Sprintf("%s [%s]", ec2.InstanceId, ec2.InstanceType),
			Location: region,
			Created:  ec2.LaunchTime,
		})
	}
	return result, nil
}
