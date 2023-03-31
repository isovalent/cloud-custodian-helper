package cmd

import (
	"context"

	"c7n-helper/pkg/log"
	"c7n-helper/pkg/parser"
	"github.com/spf13/cobra"
)

var parserCmd = &cobra.Command{
	Use:     "parse",
	Short:   "Parse C7N report directory and save result in resource JSON file",
	Aliases: []string{"p"},
	Args:    cobra.ExactArgs(0),
	Run:     parse,
}

var parseType, parseDir, parsePolicy, parseResult *string

func init() {
	parseType = parserCmd.Flags().StringP("type", "t", "", "Cloud resource type (eks, ec2, gke, gce, arg)")
	_ = parserCmd.MarkFlagRequired("type")
	parseDir = parserCmd.Flags().StringP("report-dir", "d", "", "C7N report directory")
	_ = parserCmd.MarkFlagRequired("report-dir")
	_ = parserCmd.MarkFlagDirname("report-dir")
	parsePolicy = parserCmd.Flags().StringP("policy", "p", "", "C7N policy name")
	_ = parserCmd.MarkFlagRequired("policy")
	parseResult = parserCmd.Flags().StringP("resource-file", "r", "resources.json", "Resource JSON file")
	_ = parserCmd.MarkFlagFilename("resource-file")
	rootCmd.AddCommand(parserCmd)
}

func parse(_ *cobra.Command, _ []string) {
	ctx := context.Background()
	if err := parser.Parse(ctx, *parseType, *parseDir, *parsePolicy, *parseResult); err != nil {
		log.FromContext(ctx).Fatal(err)
	}
}
