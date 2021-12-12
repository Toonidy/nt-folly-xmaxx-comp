package serve

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
	Use:   "nt-folly-xmaxx-comp-serve",
	Short: "Nitro Type Folly Team Xmaxx Comp Data API Server.",
	Long:  "Nitro Type Folly Team Xmaxx Comp Data API Server.",
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
	rootCmd.PersistentFlags().String("api_addr", ":8080", "api server to host as")
	rootCmd.PersistentFlags().String("cors_allowed_origins", "*", "allowed origins to access CORS")
	rootCmd.PersistentFlags().String("cors_allowed_methods", "GET,POST,PUT,DELETE,OPTIONS", "allowed http methods for CORS")
	rootCmd.PersistentFlags().String("cors_allowed_headers", "Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Requested-With", "allowed http headers for CORS")
	rootCmd.PersistentFlags().Bool("cors_allow_credentials", true, "whether to allow credentials for CORS")
	rootCmd.PersistentFlags().Int("cors_max_age", 1728000, "TTL to cache CORS")

	viper.BindPFlag("db_url", rootCmd.PersistentFlags().Lookup("db_url"))
	viper.BindPFlag("db_host", rootCmd.PersistentFlags().Lookup("db_host"))
	viper.BindPFlag("db_port", rootCmd.PersistentFlags().Lookup("db_port"))
	viper.BindPFlag("db_user", rootCmd.PersistentFlags().Lookup("db_user"))
	viper.BindPFlag("db_pass", rootCmd.PersistentFlags().Lookup("db_pass"))
	viper.BindPFlag("db_name", rootCmd.PersistentFlags().Lookup("db_name"))
	viper.BindPFlag("db_sslmode", rootCmd.PersistentFlags().Lookup("db_sslmode"))
	viper.BindPFlag("api_addr", rootCmd.PersistentFlags().Lookup("api_addr"))
	viper.BindPFlag("cors_allowed_origins", rootCmd.PersistentFlags().Lookup("cors_allowed_origins"))
	viper.BindPFlag("cors_allowed_methods", rootCmd.PersistentFlags().Lookup("cors_allowed_methods"))
	viper.BindPFlag("cors_allowed_headers", rootCmd.PersistentFlags().Lookup("cors_allowed_headers"))
	viper.BindPFlag("cors_allow_credentials", rootCmd.PersistentFlags().Lookup("cors_allow_credentials"))
	viper.BindPFlag("cors_max_age", rootCmd.PersistentFlags().Lookup("cors_max_age"))

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
