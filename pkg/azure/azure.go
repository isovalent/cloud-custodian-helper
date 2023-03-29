package azure

import (
	"c7n-helper/pkg/dto"
	"encoding/json"
	"time"
)

type tags struct {
	Expiry string `json:"expiry"`
}

func RG(_ string, content []byte) ([]dto.Resource, error) {
	var groups []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
		Tags     tags   `json:"tags"`
	}
	if err := json.Unmarshal(content, &groups); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(groups))
	for _, group := range groups {
		created, err := time.Parse("2006-01-02 15:04:05", group.Tags.Expiry)
		if err != nil {
			created = time.Now()
		}
		result = append(result, dto.Resource{
			Name:     group.Name,
			Location: group.Location,
			Created:  created,
		})
	}
	return result, nil
}
