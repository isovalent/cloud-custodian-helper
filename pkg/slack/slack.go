package slack

import (
	"bytes"
	"c7n-helper/pkg/dto"
	"fmt"
	"github.com/lensesio/tableprinter"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	MaxSlackMessageLength = 3_900
	SplitMessageThreshold = MaxSlackMessageLength - MaxSlackMessageLength/5
)

type Resource struct {
	Index   int    `header:"#"`
	Region  string `header:"Region"`
	Name    string `header:"Name"`
	Created string `header:"Created date"`
}

func Notify(resourceFile, url, title string) error {
	log.Println("Reading resource file...")
	var report dto.PolicyReport
	if err := report.ReadFromFile(resourceFile); err != nil {
		return err
	}
	if len(report.Accounts) == 0 {
		log.Println("Nothing to send...")
		return nil
	}
	log.Println("Preparing slack messages...")
	messages := reportToSlackMessages(title, report)
	log.Println("Sending slack notification...")
	return notifySlack(url, messages)
}

func reportToSlackMessages(title string, report dto.PolicyReport) []string {
	messages := make([]string, 0, len(report.Accounts)+1)
	if title != "" {
		messages = append(messages, fmt.Sprintf("{\"text\":\"%s\"}", title))
	}
	for _, account := range report.Accounts {
		buf := bytes.NewBufferString("")
		tableprinter.Print(buf, resourcesFromDto(account.Resources))
		for _, message := range normalizeMessage(buf.String()) {
			payload := fmt.Sprintf("*%s*\n```\n%s```\n", account.Name, message)
			messages = append(messages, fmt.Sprintf("{\"text\":\"%s\"}", payload))
		}
	}
	return messages
}

func resourcesFromDto(resources []dto.Resource) []Resource {
	result := make([]Resource, 0, len(resources))
	for i, r := range resources {
		result = append(result, Resource{
			Index:   i + 1,
			Region:  r.Location,
			Name:    r.Name,
			Created: r.Created.Format("2006-01-02"),
		})
	}
	return result
}

func normalizeMessage(message string) []string {
	if utf8.RuneCountInString(message) > MaxSlackMessageLength {
		messages := make([]string, 0)
		builder := strings.Builder{}
		lines := strings.Split(message, "\n")
		for _, line := range lines {
			builder.WriteString(line + "\n")
			if utf8.RuneCountInString(builder.String()) > SplitMessageThreshold {
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

func notifySlack(url string, messages []string) error {
	client := &http.Client{}
	for _, payload := range messages {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if err := sendMessage(client, req); err != nil {
			return err
		}
	}
	return nil
}

func sendMessage(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
