/*
Copyright © 2022 K.Awata

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
	"bufio"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var schtabFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     `schtab {file | -}`,
	Short:   "schtab sets tasks to Windows Task Scheduler from a text in crontab format",
	Version: "0.1.0",
	Args:    cobra.ExactArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// Read input
		var src *os.File
		if args[0] == "-" {
			src = os.Stdin
		} else {
			var err error
			src, err = os.Open(args[0])
			cobra.CheckErr(err)
			defer src.Close()
		}
		r := bufio.NewReader(src)
		s, err := io.ReadAll(r)
		cobra.CheckErr(err)
		if len(s) == 0 {
			return
		}

		// Write to schtab
		dst, err := os.Create(schtabFile)
		cobra.CheckErr(err)
		defer dst.Close()
		_, err = dst.Write(s)
		cobra.CheckErr(err)

		// Register
		regCmd.Run(cmd, nil)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.schtab.yaml)")
	rootCmd.PersistentFlags().StringVar(&schtabFile, "schtab", "", `schtab file (default is %APPDATA%\schtab)`)
	if schtabFile == "" {
		schtabFile = os.Getenv("APPDATA") + string(os.PathSeparator) + "schtab"
	}

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
