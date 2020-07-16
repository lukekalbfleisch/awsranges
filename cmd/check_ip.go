package cmd

import (
	"strings"

	"github.com/lukekalbfleisch/awsranges"
	"github.com/spf13/cobra"
)

var (
	IPCmd = &cobra.Command{
		Use:   "check-ip",
		Short: "Check if an IP address or network belongs to AWS",
		Run: func(cmd *cobra.Command, args []string) {
			addr := args[0]
			if strings.Contains(addr, "/") {
				checkCIDR(addr)
			} else {
				checkIP(addr)
			}
		},
	}
)

func checkIP(addr string) (bool, error) {
	ranges, err := awsranges.New()
	if err != nil {
		return false, err
	}
	return ranges.CheckAddress(addr)
}

func checkCIDR(addr string) (bool, error) {
	ranges, err := awsranges.New()
	if err != nil {
		return false, err
	}
	return ranges.CheckCIDR(addr)
}
