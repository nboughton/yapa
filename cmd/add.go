// Copyright © 2021 Nick Boughton <nicholasboughton@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"

	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Load a new RSS feed to the store",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Loading ", args[0])
		f, err := pod.FromRSS(args[0])
		if err != nil {
			log.Fatal(err)
		}

		if !store.Exists(f.Title) {
			store.Feeds = append(store.Feeds, &f)
			if err := pod.WriteStore(store); err != nil {
				log.Fatal(err)
			}
			return
		}
		fmt.Printf("Existing feed with name [%s] found.\n", f.Title)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
