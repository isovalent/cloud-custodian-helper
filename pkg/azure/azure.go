package azure

import (
	"encoding/json"
	"time"

	"c7n-helper/pkg/dto"
)

type tags struct {
	Owner string `json:"owner"`
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
		result = append(result, dto.Resource{
			Name:     group.Name,
			Location: group.Location,
			Owner:    group.Tags.Owner,
			Created:  time.Now(),
		})
	}
	return result, nil
}
