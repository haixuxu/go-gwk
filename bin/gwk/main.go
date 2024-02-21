package main

import (
	"fmt"
	gwk "github/xuxihai123/go-gwk/v1/src"
	"github/xuxihai123/go-gwk/v1/src/types"
	utils "github/xuxihai123/go-gwk/v1/src/utils"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// 版本号
	Version = "1.0.0"
	// Git提交哈希
	GitCommitHash = "none"
)

var cfgFile string
var port float64
var subdomain string

var RootCmd = &cobra.Command{
	Use: "gwk",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("start gwk client.\n")

		cliopts := &types.ClientOpts{
			LogLevel:   viper.GetString("logLevel"),
			ServerHost: viper.GetString("serverHost"),
			ServerPort: viper.GetInt("serverPort"),
			Tunnels:    nil,
		}

		if cliopts.ServerHost == "" {
			cliopts.ServerHost = "gank.75cos.com"
			cliopts.ServerPort = 4443
		}

		tunnels1 := viper.Get("tunnels")
		var tunnels map[string]interface{}
		if tunnels1 == nil {
			if port > 65534 || port < 1024 {
				fmt.Println("invalid local port", port)
				os.Exit(1)
			}
			tunnels = make(map[string]interface{})
			tunobj := make(map[string]interface{})
			if subdomain == "" {
				subdomain = utils.GenSubdomain()
			}
			tunobj["subdomain"] = subdomain
			tunobj["localport"] = port
			tunnels["unamed"] = tunobj
		} else {
			tunnels = tunnels1.(map[string]interface{})
		}

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
		fmt.Printf("gwkd release %s-%s\n", Version, GitCommitHash)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "--config client.json")
	RootCmd.Flags().Float64VarP(&port, "port", "p", 8080, "set web tunnel local port")
	RootCmd.Flags().StringVarP(&subdomain, "subdomain", "s", "", "set web tunnel subdomain")
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
