package cmd

import (
	"c7n-helper/pkg/slack"
	"context"
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

var slackResourceFile, slackToken, slackChannel, slackMembersFile, slackTitle *string

func init() {
	slackResourceFile = slackCmd.Flags().StringP("resource-file", "r", "resources.json", "Resource JSON file")
	_ = slackCmd.MarkFlagFilename("resource-file")
	slackToken = slackCmd.Flags().StringP("auth-token", "a", "", "Slack token")
	_ = slackCmd.MarkFlagRequired("auth-token")
	slackChannel = slackCmd.Flags().StringP("channel", "c", "", "Slack default channel ID")
	_ = slackCmd.MarkFlagRequired("channel")
	slackMembersFile = slackCmd.Flags().StringP("members", "m", "", "Slack members YAML file")
	_ = slackCmd.MarkFlagFilename("members")
	slackTitle = slackCmd.Flags().StringP("title", "t", "", "Slack notification title")
	rootCmd.AddCommand(slackCmd)
}

func notify(_ *cobra.Command, _ []string) {
	ctx := context.Background()
	if err := slack.Notify(ctx, *slackResourceFile, *slackToken, *slackChannel, *slackMembersFile, *slackTitle); err != nil {
		log.Fatal(err.Error())
	}
}
