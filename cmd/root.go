/*
Copyright © 2022 Gerrard-YNWA gyc.ssdut@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/Gerrard-YNWA/gitlab-analyzer/gitlab"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitlab-analyzer",
	Short: "analyze gitlab project commit stats",
	Long:  `analyze gitlab project commit stats, support analyze specified projects(if not specified scans all projects), support fetch specified author's commit info with line stats and filterd with time range`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		g := gitlab.New(viper.GetString("host"),
			viper.GetString("token"),
			"/api/v4").WithSpecifiedProjects(viper.GetStringSlice("projects"))

		repos, err := g.FetchRepos()
		if err != nil {
			panic(err)
		}

		author := viper.GetString("author")
		from := viper.GetString("from")
		to := viper.GetString("to")
		var commits int
		for _, repo := range repos {
			if author != "" {
				repo.WithSpecifiedAuthor(author)
			}
			if from != "" || to != "" {
				repo.WithDuration(from, to)
			}

			if err := repo.FetchCommits(); err != nil {
				panic(err)
			}

			var authorInfos []*gitlab.Author
			for _, v := range repo.AuthorInfos {
				authorInfos = append(authorInfos, v)
			}
			sort.SliceStable(authorInfos, func(i, j int) bool {
				return authorInfos[i].Count > authorInfos[j].Count
			})

			if bs, err := json.MarshalIndent(authorInfos, "", "\t"); err != nil {
				panic(err)
			} else {
				fmt.Printf("Repo: %s, Commits:%d\nDetail:\n%s\n", repo.Name, len(repo.FilteredCommits), string(bs))
			}
			commits += len(repo.FilteredCommits)
		}

		fmt.Printf("Gitlab: %d Commits on %d Repos.", commits, len(repos))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gitlab-analyzer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gitlab-analyzer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
