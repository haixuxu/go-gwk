package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gwk "github/xuxihai123/go-gwk/v1/src"
	"github/xuxihai123/go-gwk/v1/src/types"
	"os"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use: "gwk",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("start gwk client.\n")

		cliopts := &types.ClientOpts{
			LogLevel:   viper.GetString("logLevel"),
			TunnelHost: viper.GetString("tunnelHost"),
			TunnelAddr: viper.GetInt("tunnelAddr"),
			Tunnels:    nil,
		}

		tunnels := viper.Get("tunnels").(map[string]interface{})

		tunnelDict := make(map[string]*types.TunnelOpts)
		for key, value := range tunnels {

			tun := value.(map[string]interface{})
			tunobj := types.TunnelOpts{Status: "init"}
			tunobj.Name = key
			for key2, value2 := range tun {
				if key2 == "type" {
					tunobj.Type = value2.(string)
				}
				if key2 == "localport" {
					tunobj.LocalPort = int(value2.(float64))
				}
				if key2 == "subdomain" {
					tunobj.Subdomain = value2.(string)
				}

				if key2 == "remoteport" {
					tunobj.RemotePort = int(value2.(float64))
				}
			}
			tunnelDict[key] = &tunobj
		}
		cliopts.Tunnels = tunnelDict

		cli := gwk.NewClient(cliopts)
		cli.Bootstrap()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gwk",
	Long:  `All software has versions. This is xuxihai's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gwk release v0.0.1 -- HEAD")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "--config client.json")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile == "" {
		return
	}
	// Use config file from the flag.
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
