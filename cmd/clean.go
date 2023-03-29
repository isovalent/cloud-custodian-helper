package cmd

import (
	"c7n-helper/pkg/cleaner"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Clean all resources from resource file",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(0),
	Run:     clean,
}

var (
	cleanFile  *string
	cleanTries *int
	cleanRetry *time.Duration
)

func init() {
	cleanFile = cleanCmd.Flags().StringP("resource-file", "r", "", "Resource JSON file")
	_ = cleanCmd.MarkFlagRequired("resource-file")
	_ = cleanCmd.MarkFlagFilename("resource-file")
	cleanTries = cleanCmd.Flags().IntP("tries-count", "t", 3, "Clean tries count")
	cleanRetry = cleanCmd.Flags().DurationP("retry-duration", "d", time.Minute, "Clean retry pause")
	rootCmd.AddCommand(cleanCmd)
}

func clean(_ *cobra.Command, _ []string) {
	if err := cleaner.Clean(*cleanFile, *cleanTries, *cleanRetry); err != nil {
		log.Fatal(err.Error())
	}
}
