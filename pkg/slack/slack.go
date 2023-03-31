package slack

import (
	"bytes"
	"c7n-helper/pkg/dto"
	"c7n-helper/pkg/log"
	"context"
	"fmt"
	"github.com/lensesio/tableprinter"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	maxSlackMessageLength = 3_000
	splitMessageThreshold = maxSlackMessageLength - maxSlackMessageLength/5
)

type msgLine struct {
	Index   int    `header:"#"`
	Region  string `header:"Region"`
	Name    string `header:"Name"`
	Created string `header:"Created date"`
}

func Notify(ctx context.Context, resourceFile, slackToken, slackDefaultChannel, membersFile, title string) error {
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
	slackClient := slack.New(slackToken)
	logger.Info("reading slack members file...")
	slackMembers, err := readSlackMembers(ctx, membersFile, slackClient)
	if err != nil {
		return err
	}
	logger.Info("preparing slack messages...")
	slackGroups := groupSlackMessage(report.Accounts, slackMembers, slackDefaultChannel)
	channelMessages := prepareSlackMessage(slackGroups)
	logger.Info("sending slack notification...")
	return notifySlack(ctx, slackClient, title, channelMessages)
}

func readSlackMembers(ctx context.Context, file string, client *slack.Client) (map[string]string, error) {
	if file == "" {
		return nil, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}
	if len(rawData) == 0 {
		return nil, nil
	}
	members := rawData["members"].(map[string]interface{})
	result := make(map[string]string)
	for name := range members {
		member := members[name].(map[string]interface{})
		if id, ok := member["slackID"]; ok {
			slackID := id.(string)
			// Filter out not existed users
			if _, err := client.GetUserInfo(slackID); err != nil {
				log.FromContext(ctx).Errorf("slack user [%s] not found: %s", slackID, err.Error())
				continue
			}
			result[strings.ToLower(name)] = id.(string)
		}
	}
	return result, nil
}

// Groups Slack messages: SlackChannelID -> Account|Project|Subscription -> []Resources
func groupSlackMessage(accounts []dto.Account, slackMembers map[string]string, defaultChannel string) map[string]map[string][]dto.Resource {
	groups := make(map[string]map[string][]dto.Resource)
	for _, account := range accounts {
		for _, resource := range account.Resources {
			channel, ok := slackMembers[strings.ToLower(resource.Owner)]
			if !ok {
				channel = defaultChannel
			}
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
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].Created.Before(resources[j].Created)
			})
			buf := bytes.NewBufferString("")
			tableprinter.Print(buf, normalizeDTO(resources))
			for _, message := range splitMessage(buf.String()) {
				payload := fmt.Sprintf("*%s*\n```\n%s```\n", account, message)
				channelMessages[channel] = append(channelMessages[channel], payload)
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
		})
	}
	return result
}

func splitMessage(message string) []string {
	if utf8.RuneCountInString(message) > maxSlackMessageLength {
		messages := make([]string, 0)
		builder := strings.Builder{}
		lines := strings.Split(message, "\n")
		for _, line := range lines {
			builder.WriteString(line + "\n")
			if utf8.RuneCountInString(builder.String()) > splitMessageThreshold {
				messages = append(messages, builder.String())
				builder.Reset()
			}
		}
		if builder.Len() > 0 {
			messages = append(messages, builder.String())
		}
		return messages
	}
	return []string{message}
}

func notifySlack(ctx context.Context, client *slack.Client, title string, channelMessages map[string][]string) error {
	var header *slack.HeaderBlock
	if title != "" {
		header = &slack.HeaderBlock{
			Type: slack.MBTHeader,
			Text: slack.NewTextBlockObject(slack.PlainTextType, title, false, false),
		}
	}
	for channel, messages := range channelMessages {
		if header != nil {
			_, _, _, err := client.SendMessageContext(ctx, channel, slack.MsgOptionBlocks(header))
			if err != nil {
				return err
			}
		}
		for _, message := range messages {
			_, _, _, err := client.SendMessageContext(ctx, channel, slack.MsgOptionText(message, false))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
