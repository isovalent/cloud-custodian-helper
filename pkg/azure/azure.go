package azure

import (
	"c7n-helper/pkg/dto"
	"encoding/json"
	"time"
)

type rg struct {
	Name string `json:"name"`
}

func RG(content []byte) ([]dto.Resource, error) {
	var groups []rg
	if err := json.Unmarshal(content, &groups); err != nil {
		return nil, err
	}
	result := make([]dto.Resource, 0, len(groups))
	for _, group := range groups {
		result = append(result, dto.Resource{
			Name:    group.Name,
			Created: time.Now(), //TODO: fix it
		})
	}
	return result, nil
}
