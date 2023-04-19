package aws

import (
	"encoding/json"
	"time"

	"c7n-helper/pkg/dto"
)

func ParseS3(region string, content []byte) ([]dto.Resource, error) {
	var buckets []struct {
		Name      string     `json:"Name"`
		CreatedAt time.Time  `json:"CreationDate"`
		Tags      []keyValue `json:"Tags"`
	}
	if err := json.Unmarshal(content, &buckets); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(buckets))
	for _, bucket := range buckets {
		owner := ""
		for _, tag := range bucket.Tags {
			if tag.Key == "owner" {
				owner = tag.Value
				break
			}
		}
		result = append(result, dto.Resource{
			Name:     bucket.Name,
			Location: region,
			Owner:    owner,
			Created:  bucket.CreatedAt,
		})
	}
	return result, nil
}
