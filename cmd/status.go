package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sslcon/rpc"
)

var status = &cobra.Command{
	Use:   "status",
	Short: "Get VPN connection information",
	Run: func(cmd *cobra.Command, args []string) {
		var result json.RawMessage
		err := rpcCall("status", nil, &result, rpc.STATUS)
		if err != nil {
			after, _ := strings.CutPrefix(err.Error(), "jsonrpc2: code 1 message: ")
			fmt.Println(after)
		} else {
			prettyPrint(result)
		}
	},
}

func init() {
	rootCmd.AddCommand(status)
}
