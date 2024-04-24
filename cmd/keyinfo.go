/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package cmd

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rightfoot-consulting/p2pbbs/bbscrypto"
	"github.com/spf13/cobra"
)

// keyinfoCmd represents the keyinfo command
var keyinfoCmd = &cobra.Command{
	Use:   "keyinfo",
	Short: "Read a private key and print out it's corresponding peer id",
	Long: `Peer ids are a multihash of the public key for a node.  This key is read from the private key file generated with generateKey.
 This command will read the private key file and print out the corresponding peer id.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("keyinfo called")
		fileName, err := cmd.Flags().GetString("file")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Reading file '%s'\n", fileName)

		privateKey, err := bbscrypto.LoadPrivateKey(fileName)
		if err != nil {
			panic(err)
		}
		id, err := peer.IDFromPrivateKey(privateKey)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Id: %s\n\n", id.String())
	},
}

func init() {
	rootCmd.AddCommand(keyinfoCmd)
	keyinfoCmd.Flags().StringP("file", "f", "", "Specifies the private key to read in, this is used as the basis for the peer id.")
}
