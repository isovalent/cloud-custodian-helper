package cloud

import (
	"c7n-helper/pkg/converter"
	"c7n-helper/pkg/dto"
	"encoding/json"
	"fmt"
	"time"
)

type EKS struct {
	Name      string    `converter:"name"`
	CreatedAt time.Time `converter:"createdAt"`
}

type EC2 struct {
	InstanceId   string    `converter:"InstanceId"`
	LaunchTime   time.Time `converter:"LaunchTime"`
	InstanceType string    `converter:"InstanceType"`
}

func EksFromFile(file string) ([]dto.Resource, error) {
	content, err := converter.JsonToBytes(file)
	if err != nil {
		return nil, err
	}
	var clusters []EKS
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

func Ec2FromFile(file string) ([]dto.Resource, error) {
	content, err := converter.JsonToBytes(file)
	if err != nil {
		return nil, err
	}
	var vms []EC2
	if err := json.Unmarshal(content, &vms); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(vms))
	for _, ec2 := range vms {
		result = append(result, dto.Resource{
			Name:    fmt.Sprintf("%s [%s]", ec2.InstanceId, ec2.InstanceType),
			Created: ec2.LaunchTime,
		})
	}
	return result, nil
}
