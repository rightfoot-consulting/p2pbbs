/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package cmd

import (
	"fmt"

	"github.com/rightfoot-consulting/p2pbbs/chat"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start a command line chat",
	Long: `Use this command to establish a distributed chat on the command line. For example:

			chat --listen 6666 
			Will listen on port 6666 and use ./config.json for configuration

			chat --listen 6655 --config /etc/chat/config.json
			Will listen on port 6655 and use /etc/chat/config.json for configuration
		.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("chat called")

		query, err := cmd.Flags().GetBool("query")
		if err != nil {
			panic(err)
		}
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			panic(err)
		}
		config, err := chat.LoadChatConfig(configFile)
		if err != nil {
			panic(err)
		}
		keyFile, err := cmd.Flags().GetString("keyfile")
		if err != nil {
			panic(err)
		}
		if keyFile != "" {
			config.KeyFile = keyFile
		}
		listen, err := cmd.Flags().GetStringArray("listen")
		if err != nil {
			panic(err)
		}
		if len(listen) > 0 {
			config.ListenIps = listen
		}
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			panic(err)
		}
		if group != "" {
			config.RendezvousString = group
		}
		var bsPeers []string
		bsPeers, err = cmd.Flags().GetStringArray("bootstrap-peers")
		if err != nil {
			panic(err)
		}
		config.BootstrapPeers = append(config.BootstrapPeers, bsPeers...)

		port, err := cmd.Flags().GetInt32("port")
		if err != nil {
			panic(err)
		}
		if port > 0 {
			config.Port = int(port)
		}

		node, err := chat.NewChatNode(config)
		if err != nil {
			panic(err)
		}
		if query {
			node.Query()
		} else {
			node.Run()
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().StringP("config", "c", "~/.p2bbs/chatconfig.json", "Location of the configuration file (default '~/.p2bbs/chatconfig.json')")
	chatCmd.Flags().BoolP("query", "q", false, "When true prints configuration information and default relays")
	chatCmd.Flags().StringArrayP("listen", "l", []string{}, "Specifies network interfaces to listen on (default '0.0.0.0')")
	chatCmd.Flags().StringP("keyfile", "k", "", "Specifies a key file to use for static addressed nodes")
	chatCmd.Flags().StringP("group", "g", "", "Unique string to identify group of nodes. Default provided in config.")
	chatCmd.Flags().StringArrayP("bootstrap-peers", "b", []string{}, "Adds a public peer multiaddreses to the bootstrap list")
	chatCmd.Flags().Int32P("port", "p", -1, "Specifies the listen port")
	/*
		chatCmd.Flags().Int32P("port", "p", 6666, "Specifies the listen port")
		chatCmd.Flags().StringP("protocol-id", "i", "/chat/1.1.0", "Sets a protocol id for stream headers")


	*/
}
