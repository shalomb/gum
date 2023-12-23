// Package cmd implements our commands
package cmd

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the database",
	Long:  `Update the database`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if err := doUpdate(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	updateCmd.Flags().BoolP("all", "a", true, "Update all targets (default)")
	updateCmd.Flags().BoolP("projects", "p", false, "Update projects")
	updateCmd.Flags().BoolP("dirs", "d", false, "Update dirs")

	viper.SetDefault("CacheDir", xdg.CacheHome)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(xdg.ConfigHome, "gum"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("fatal: Missing config files: %T", err))
		}
		panic(fmt.Errorf("fatal: Error reading in config: %T", err))
	}
}

func doUpdate() error {
	fmt.Printf("doUpdate called")
	for key, value := range viper.GetViper().AllSettings() {
		log.WithFields(log.Fields{
			key: value,
		}).Info("\nCommand Flag")
	}

	curUser, _ := user.Current()
	homeDir := curUser.HomeDir
	var result []string

	projectDirs := viper.GetStringSlice("projects")
	for _, dir := range projectDirs {
		if strings.HasPrefix(dir, "~/") {
			target := filepath.Join(homeDir, dir[2:])
			log.Printf("\nScanning directory: %v (%v)", target, dir)
			cmd := exec.Command("find", "-L", target, "-iname", ".git", "-type", "d", "-prune")
			stdout, err := cmd.Output()
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if exiterr.ExitCode() == 1 {
						log.Printf("Exit status == 1: Ignoring")
					} else {
						log.Fatalf("error finding projects in %v: %v", target, err)
					}
				}
			}

			for _, dir := range strings.Split(string(stdout), "\n") {
				if len(dir) > 0 {
					p := dir[:len(dir)-4]  // remove ".git" at the end
					// log.Printf("%v\n", p)
					result = append(result, strings.Replace(p, homeDir, "~", 1))
				}
			}
		}
	}

	fmt.Printf("%v", result)
	// return fmt.Errorf("Heh")
	return nil
}
