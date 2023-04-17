package cmd

import (
	"context"

	"c7n-helper/pkg/log"
	"c7n-helper/pkg/slack"
	"github.com/spf13/cobra"
)

var slackCmd = &cobra.Command{
	Use:     "slack",
	Short:   "Send Slack notification (via webhook) with information from resource JSON file",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(0),
	Run:     notify,
}

var slackResourceFile, slackToken, slackChannel, slackTitle *string

func init() {
	slackResourceFile = slackCmd.Flags().StringP("resource-file", "r", "resources.json", "Resource JSON file")
	_ = slackCmd.MarkFlagFilename("resource-file")
	slackToken = slackCmd.Flags().StringP("auth-token", "a", "", "Slack token")
	_ = slackCmd.MarkFlagRequired("auth-token")
	slackChannel = slackCmd.Flags().StringP("channel", "c", "", "Slack default channel ID")
	_ = slackCmd.MarkFlagRequired("channel")
	slackTitle = slackCmd.Flags().StringP("title", "t", "", "Slack notification title")
	rootCmd.AddCommand(slackCmd)
}

func notify(_ *cobra.Command, _ []string) {
	ctx := context.Background()
	if err := slack.Notify(ctx, *slackResourceFile, *slackToken, *slackChannel, *slackTitle); err != nil {
		log.FromContext(ctx).Fatal(err)
	}
}
