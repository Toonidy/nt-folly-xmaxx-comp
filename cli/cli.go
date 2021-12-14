package cli

import (
	"log"
	"nt-folly-xmaxx-comp/internal/pkg/build"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitConfig(cfgFile *string, logger *zap.Logger) *zap.Logger {
	cobra.OnInitialize(func() {
		if cfgFile != nil && *cfgFile != "" {
			// Use config file from the flag.
			viper.SetConfigFile(*cfgFile)
		} else {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				log.Fatal(err)
			}

			// Search config in home directory with name ".nt-folly-xmaxx-comp" (without extension).
			viper.AddConfigPath(home)
			viper.SetConfigName(".nt-folly-xmaxx-comp")
		}

		viper.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			log.Println("Using config file:", viper.ConfigFileUsed())
		}
	})
	return logger.With(zap.String("version", build.Version), zap.String("build", build.BuildHash))
}

func CreateLogger() (*zap.Logger, error) {
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
	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalln("failed to setup logger", err)
	}
	return logger, nil
}
