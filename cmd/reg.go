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
	"os"

	"github.com/k-awata/schtab/schtab"
	"github.com/spf13/cobra"
)

// regCmd represents the reg command
var regCmd = &cobra.Command{
	Use:   "reg",
	Short: "Register tasks in your schtab at Task Scheduler",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(schtabFile)
		cobra.CheckErr(err)
		defer f.Close()
		cobra.CheckErr(schtab.UnregisterAll())
		cobra.CheckErr(schtab.RegisterAll(f))
		cmd.Println("schtab registered your tasks at Task Scheduler")
	},
}

func init() {
	rootCmd.AddCommand(regCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// regCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// regCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
