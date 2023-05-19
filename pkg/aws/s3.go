package aws

import (
	"encoding/json"
	"strings"
	"time"

	"c7n-helper/pkg/date"
	"c7n-helper/pkg/dto"
)

type location struct {
	LocationConstraint string `json:"LocationConstraint"`
}

func ParseS3(region string, content []byte) ([]dto.Resource, error) {
	var buckets []struct {
		Name      string     `json:"Name"`
		CreatedAt time.Time  `json:"CreationDate"`
		Tags      []keyValue `json:"Tags"`
		Location  location   `json:"Location"`
	}
	if err := json.Unmarshal(content, &buckets); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.Location.LocationConstraint != region {
			// custodian returns all buckets for each region
			// skip buckets from another regions
			continue
		}
		owner, expiry := "", ""
		for _, tag := range bucket.Tags {
			switch strings.ToLower(tag.Key) {
			case "owner":
				owner = tag.Value
			case "expiry":
				expiry = tag.Value
			}
		}
		result = append(result, dto.Resource{
			Name:     bucket.Name,
			Location: region,
			Owner:    owner,
			Created:  bucket.CreatedAt,
			Expiry:   date.ParseOrDefault(expiry, time.Now()),
		})
	}
	return result, nil
}
