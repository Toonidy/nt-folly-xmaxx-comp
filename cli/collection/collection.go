package collection

import (
	"fmt"
	"log"
	"nt-folly-xmaxx-comp/cli"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nt-folly-xmaxx-comp-collection",
	Short: "Nitro Type Folly Team Xmaxx Comp Data Collection.",
	Long:  "Nitro Type Folly Team Xmaxx Comp Data Collection.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	var cfgFile string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nt-folly-xmaxx-comp)")

	// Define Default Configuration
	rootCmd.PersistentFlags().Bool("prod", false, "Whether this is used for production.")
	rootCmd.PersistentFlags().String("db_url", "", "database url to connect to (overrides the other db_* values)")
	rootCmd.PersistentFlags().String("db_host", "localhost", "database host")
	rootCmd.PersistentFlags().String("db_port", "5444", "database port")
	rootCmd.PersistentFlags().String("db_user", "nt-folly-xmaxx-comp", "database user")
	rootCmd.PersistentFlags().String("db_pass", "dev", "database password")
	rootCmd.PersistentFlags().String("db_name", "nt-folly-xmaxx-comp", "database name")
	rootCmd.PersistentFlags().String("db_sslmode", "disable", "database connection ssl mode")
	rootCmd.PersistentFlags().Int("db_max_open_connections", 0, "max number of connections to make")
	rootCmd.PersistentFlags().String("browser_user_agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36", "browser user agent used for the data scraper")
	rootCmd.PersistentFlags().String("team_tag", "FOLLY", "team tag to track stats")
	rootCmd.PersistentFlags().Int("team_id", 1411729, "nitro type team id to cross check stats against")

	viper.BindPFlag("db_url", rootCmd.PersistentFlags().Lookup("db_url"))
	viper.BindPFlag("db_host", rootCmd.PersistentFlags().Lookup("db_host"))
	viper.BindPFlag("db_port", rootCmd.PersistentFlags().Lookup("db_port"))
	viper.BindPFlag("db_user", rootCmd.PersistentFlags().Lookup("db_user"))
	viper.BindPFlag("db_pass", rootCmd.PersistentFlags().Lookup("db_pass"))
	viper.BindPFlag("db_name", rootCmd.PersistentFlags().Lookup("db_name"))
	viper.BindPFlag("db_sslmode", rootCmd.PersistentFlags().Lookup("db_sslmode"))
	viper.BindPFlag("db_max_open_connections", rootCmd.PersistentFlags().Lookup("db_max_open_connections"))
	viper.BindPFlag("browser_user_agent", rootCmd.PersistentFlags().Lookup("browser_user_agent"))
	viper.BindPFlag("team_tag", rootCmd.PersistentFlags().Lookup("team_tag"))
	viper.BindPFlag("team_id", rootCmd.PersistentFlags().Lookup("team_id"))

	// Setup logger
	var (
		logConfig zap.Config
		err       error
	)
	if viper.GetBool("prod") {
		logConfig = zap.NewProductionConfig()
	} else {
		logConfig = zap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	logger, err = logConfig.Build()
	if err != nil {
		log.Fatalln("failed to setup logger", err)
	}
	logger = cli.InitConfig(&cfgFile, logger)
}
