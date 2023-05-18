package slack

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"c7n-helper/pkg/dto"
	"c7n-helper/pkg/log"
	"github.com/lensesio/tableprinter"
)

type msgLine struct {
	Index   int    `header:"#"`
	Region  string `header:"Region"`
	Name    string `header:"Name"`
	Created string `header:"Created date"`
	Expiry  string `header:"Expiry date"`
}

func Notify(ctx context.Context, resourceFile, slackToken, slackDefaultChannel, title string) error {
	logger := log.FromContext(ctx)
	logger.Info("reading resource file...")
	var report dto.PolicyReport
	if err := report.ReadFromFile(resourceFile); err != nil {
		return err
	}
	if len(report.Accounts) == 0 {
		logger.Info("nothing to send")
		return nil
	}
	slack := newSlackProvider(slackToken, title, slackDefaultChannel)
	logger.Info("reading slack users...")
	if err := slack.readUsers(ctx); err != nil {
		return err
	}
	logger.Info("preparing slack messages...")
	channelMessages := prepareSlackMessage(groupSlackMessage(report.Accounts, slack))
	logger.Info("sending slack notification...")
	return slack.notify(ctx, channelMessages)
}

// Groups Slack messages: SlackChannelID -> Account|Project|Subscription -> []Resources
func groupSlackMessage(accounts []dto.Account, slack *slackProvider) map[string]map[string][]dto.Resource {
	groups := make(map[string]map[string][]dto.Resource)
	for _, account := range accounts {
		for _, resource := range account.Resources {
			channel := slack.getSlackIDByOwner(resource.Owner)
			accountResources, ok := groups[channel]
			if !ok {
				accountResources = make(map[string][]dto.Resource)
				groups[channel] = accountResources
			}
			if _, ok := accountResources[account.Name]; !ok {
				accountResources[account.Name] = make([]dto.Resource, 0)
			}
			accountResources[account.Name] = append(accountResources[account.Name], resource)
		}
	}
	return groups
}

func prepareSlackMessage(groups map[string]map[string][]dto.Resource) map[string][]string {
	channelMessages := make(map[string][]string)
	for channel, accountResources := range groups {
		channelMessages[channel] = make([]string, 0)
		for account, resources := range accountResources {
			channelMessages[channel] = append(channelMessages[channel], fmt.Sprintf("*%s*\n", account))
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].Created.Before(resources[j].Created)
			})
			buf := bytes.NewBufferString("")
			tableprinter.Print(buf, normalizeDTO(resources))
			for _, message := range splitMessage(buf.String()) {
				channelMessages[channel] = append(channelMessages[channel], fmt.Sprintf("```\n%s```\n", message))
			}
		}
	}
	return channelMessages
}

func normalizeDTO(resources []dto.Resource) []msgLine {
	result := make([]msgLine, 0, len(resources))
	for i, r := range resources {
		result = append(result, msgLine{
			Index:   i + 1,
			Region:  r.Location,
			Name:    r.Name,
			Created: r.Created.Format("2006-01-02"),
			Expiry:  r.Expiry.Format("2006-01-02"),
		})
	}
	return result
}
