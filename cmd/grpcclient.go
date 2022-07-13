/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	myclient "github.com/jackyzhangfudan/sidecar/pkg/grpc/client"
	"github.com/spf13/cobra"
)

// grpcclientCmd represents the grpcclient command
var grpcclientCmd = &cobra.Command{
	Use:   "grpcclient",
	Short: "start a gRPC client",
	Long:  `Start a gRPC client to execute gRPC call`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("grpcclient called")
	},
}

var certId *string

func init() {
	rootCmd.AddCommand(grpcclientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// grpcclientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	certId = grpcclientCmd.Flags().String("certid", "", "id of the client certificate")
	grpcclientCmd.MarkFlagRequired("certid")
}

func Run() {
	myclient.Run(*certId)
}
