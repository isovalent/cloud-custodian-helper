package azure

import (
	"c7n-helper/pkg/dto"
	"encoding/json"
	"time"
)

type rg struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Tags     tags   `json:"tags"`
}

type tags struct {
	Expiry string `json:"expiry"`
}

func RG(content []byte) ([]dto.Resource, error) {
	var groups []rg
	if err := json.Unmarshal(content, &groups); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(groups))
	for _, group := range groups {
		created, err := time.Parse(time.DateTime, group.Tags.Expiry)
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
