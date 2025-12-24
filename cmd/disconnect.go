package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sslcon/rpc"
)

var disconnect = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from the VPN server",
	Run: func(cmd *cobra.Command, args []string) {
		var result json.RawMessage
		err := rpcCall("disconnect", nil, &result, rpc.DISCONNECT)
		if err != nil {
			after, _ := strings.CutPrefix(err.Error(), "jsonrpc2: code 1 message: ")
			fmt.Println(after)
		} else {
			prettyPrint(result)
		}
	},
}

func init() {
	rootCmd.AddCommand(disconnect)
}
