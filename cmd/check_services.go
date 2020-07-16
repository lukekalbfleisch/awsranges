package cmd

import (
	"fmt"
	"strings"

	"github.com/lukekalbfleisch/awsranges"
	"github.com/spf13/cobra"
)

var (
	SvcsCmd = &cobra.Command{
		Use:   "check-services",
		Short: "Check which AWS services an IP address or network belongs to",
		Run: func(cmd *cobra.Command, args []string) {
			checkServices(args[0])
		},
	}
)

func checkServices(addr string) error {
	ranges, err := awsranges.New()
	if err != nil {
		return err
	}
	resp, err := ranges.CheckServices(addr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if len(resp.Services) == 0 {
		fmt.Printf("%s does not beling to any AWS service", addr)
	}

	svcs := strings.Join(resp.Services, ", ")
	fmt.Printf("%s belongs to the service %s in the %s region\n", addr, svcs, resp.Region)

	return nil
}
