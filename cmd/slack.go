package cmd

import (
	"c7n-helper/pkg/slack"
	"github.com/spf13/cobra"
	"log"
)

var slackCmd = &cobra.Command{
	Use:     "slack",
	Short:   "Send Slack notification (via webhook) with information from resource JSON file",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(0),
	Run:     notify,
}

var slackFile, slackURL, slackTitle *string

func init() {
	slackFile = slackCmd.Flags().StringP("resource-file", "r", "resources.json", "Resource JSON file")
	_ = slackCmd.MarkFlagFilename("resource-file")
	slackURL = slackCmd.Flags().StringP("url", "u", "", "Slack webhook URL")
	_ = slackCmd.MarkFlagRequired("url")
	slackTitle = slackCmd.Flags().StringP("title", "t", "", "Slack notification title")
	rootCmd.AddCommand(slackCmd)
}

func notify(_ *cobra.Command, _ []string) {
	if err := slack.Notify(*slackFile, *slackURL, *slackTitle); err != nil {
		log.Fatal(err.Error())
	}
}
