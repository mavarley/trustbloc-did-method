/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/trustbloc/trustbloc-did-method/cmd/did-method-rest/startcmd"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "rp",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	rootCmd.AddCommand(startcmd.GetStartCmd(&startcmd.HTTPServer{}))

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to run rp: %s", err.Error())
	}
}
