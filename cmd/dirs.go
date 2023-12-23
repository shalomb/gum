/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	ps "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// dirsCmd represents the dirs command
var dirsCmd = &cobra.Command{
	Use:   "dirs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		doUpdateDirs()
	},
}

func init() {
	rootCmd.AddCommand(dirsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dirsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dirsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func doUpdateDirs() {
	pslist, err := ps.Processes()
	if err != nil {
		log.Printf("error listing processes")
	}

	dirs := make(map[string]int64)

	for _, pid := range pslist {
		// log.Printf("%d\t%s\t%+v\n", pid.Pid(), pid.Executable(), pid)
		cwd := fmt.Sprintf("/proc/%d/cwd", pid.Pid())

		dir, _ := os.Readlink(cwd)
		// if err != nil {
			// log.Printf("error reading cwd for pid: %d", pid)
		// }

		if len(dir) > 0 {
			now := time.Now()
			var newval int64 = 1
			if val, ok := dirs[dir]; ok {
				newval = int64(val)
			}
			log.Printf("%v \t -> %v | %v\t-> %v\n", dirs[dir], now.UnixNano(), newval, dir)
			dirs[dir] = (now.UnixNano() - int64(newval))
			log.Printf("%v \t -> %v | %v\t-> %v\n", dirs[dir], now.UnixNano(), newval, dir)
		}
	}

	keys := make([]string, 0, len(dirs))
	for key := range dirs {
		keys = append(keys, key)
	}
	sort.Slice(keys,
		func(i, j int) bool {
			return dirs[keys[i]] > dirs[keys[j]]
		})

	for _, key := range keys {
		fmt.Printf("%v\t%v\n", dirs[key], key)
	}
}
