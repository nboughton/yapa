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
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	store      *pod.Store
	showNotify bool

	dateFmt = "2006-01-02 15:04"

	tw = tabwriter.NewWriter(os.Stdout, 1, 2, 1, ' ', 0)
)

const (
	hideCursor = "civis"
	showCursor = "cvvis"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yapa",
	Short: "A basic podcast aggregator and player for listening to podcasts in episode order",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		store, err = pod.ReadStore(viper.GetString("store"))
		if err != nil && err.Error() == pod.ErrorInvalidPath {
			log.Fatal(err)
		} else if err != nil && err.Error() == pod.ErrorStoreDoesNotExist {
			fmt.Println("No store found, creating blank db...")
		}

		showNotify = viper.GetBool("notify")
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/yapa/config.json)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".yapa" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".config/yapa/config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("")
	}
}

// Check feed id
func validFeed(id int) error {
	if id >= len(store.Feeds) {
		return fmt.Errorf("no feed with id %d", id)
	}

	return nil
}

// clear terminal screen
func clear() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// run tput (used for show/hide cursor during playback)
func tput(arg string) error {
	cmd := exec.Command("tput", arg)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}

// Send desktop notifications via dbus
func sendNotify(title, message string) error {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = conn.Auth(nil); err != nil {
		return err
	}

	if err = conn.Hello(); err != nil {
		return err
	}

	// Send notification
	_, err = notify.SendNotification(conn, notify.Notification{
		AppName:       "yapa",
		ReplacesID:    uint32(0),
		Summary:       title,
		Body:          message,
		Hints:         map[string]dbus.Variant{},
		ExpireTimeout: 10000,
	})

	return err
}
