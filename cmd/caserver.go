/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"

	grpcserver "github.com/jackyzhangfudan/sidecar/pkg/grpc/server"
	"github.com/jackyzhangfudan/sidecar/pkg/httpserver"
	"github.com/jackyzhangfudan/sidecar/pkg/util"
)

// caserverCmd represents the caserver command
var caserverCmd = &cobra.Command{
	Use:   "caserver",
	Short: "start CA web server",
	Long:  `A CA web server can response to certificate signing request`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

var useGRPC *bool
var useMTLS *bool

func init() {
	rootCmd.AddCommand(caserverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// caserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// caserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	useGRPC = caserverCmd.Flags().Bool("grpc", true, "enable the gRPC instead of http1.1")
	useMTLS = caserverCmd.Flags().Bool("mtls", true, "enable the mtls for gRPC, no effect when don't use gRPC")
}

/*
start the http server
*/
func startServer() {
	if *useGRPC {
		grpcserver.Run(*useMTLS, util.Shutdown())
	} else {
		httpserver.Run()
	}
}
