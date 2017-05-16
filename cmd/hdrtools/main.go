package main

import (
	"fmt"

	"github.com/mdouchement/hdr/cmd/hdrtools/cmd"
	"github.com/spf13/cobra"
)

func main() {
	c := &cobra.Command{
		Use: "hdrtools",
	}
	c.AddCommand(cmd.QualityCommand)
	c.AddCommand(cmd.ConvertCommand)

	if err := c.Execute(); err != nil {
		fmt.Println(err)
	}
}
