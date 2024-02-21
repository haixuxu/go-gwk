package main

import (
	"fmt"
	gwk "github/xuxihai123/go-gwk/v1/src"
	"github/xuxihai123/go-gwk/v1/src/types"
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

var RootCmd = &cobra.Command{
	Use: "gwkd",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("start gwkd server.\n")

		servopts := &types.ServerOpts{
			ServerHost: viper.GetString("serverHost"),
			ServerPort: viper.GetInt("serverPort"),
			LogLevel:   viper.GetString("logLevel"),
			HttpAddr:   viper.GetInt("httpAddr"),
			HttpsAddr:  viper.GetInt("httpsAddr"),
			TlsCA:      viper.GetString("tlsCA"),
			TlsCrt:     viper.GetString("tlsCrt"),
			TlsKey:     viper.GetString("tlsKey"),
		}
		svr := gwk.NewServer(servopts)
		svr.Bootstrap()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gwkd",
	Long:  `All software has versions. This is xuxihai's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gwkd release %s-%s\n", Version, GitCommitHash)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "--config server.json")
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
