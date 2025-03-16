// fetch.go
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/atotto/clipboard"
	pkg "github.com/julilili42/GitSnatch/pkg"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch [repoOwner] [repoName]",
	Short: "Fetch all file contents from a GitHub repository",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		var repoOwner, repoName, commitSHA string

		if len(args) >= 1 {
			repoOwner = args[0]
		} else {
			repoOwner = pkg.AskQuestion("Who is the owner of the GitHub repo?")
		}

		if len(args) >= 2 {
			repoName = args[1]
		} else {
			repoName = pkg.AskQuestion("What is the name of the GitHub repo?")
		}

		if len(args) >= 3 {
			commitSHA = args[2]
		} else {
			commitSHA = pkg.AskQuestion("Enter the commit SHA from the GitHub repository.")
			if commitSHA == "" {
				pkg.InfoText("No commit SHA entered, press ENTER to select latest commit.")
			}
		}

		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			log.Fatal("GITHUB_TOKEN environment variable not set")
		}

		client := pkg.NewClient(token)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		fileTree, err := pkg.FetchRepositoryData(ctx, client, nil, repoOwner, repoName, commitSHA)
		if err != nil {
			log.Fatal(err)
		}

		paths := pkg.GetSortedFilePaths(fileTree)

		selectedFiles := pkg.MultiSelect("Select the files you want to copy.", paths)

		selectedTree := pkg.FilterFileTree(fileTree, selectedFiles)

		finalContent := pkg.ProcessFileTree(ctx, client, selectedTree)

		if err := clipboard.WriteAll(finalContent); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if finalContent == "" {
			pkg.InfoText("No content copied to clipboard.")
			os.Exit(0)
		}

		pkg.InfoText("Copied to clipboard successfully.")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
