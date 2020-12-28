package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	zlog "github.com/jblunck/bp/internal/zerolog"
)

var (
	// Version information from `build/setlocalversion`
	Version string
	// GitCommitSha information from `build/setlocalversion --git-commit-sha`
	GitCommitSha string
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/config/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.ReadInConfig()
	viper.SetDefault("port", "8080")
	viper.SetDefault("base_url", fmt.Sprintf("http://localhost:%d", viper.GetInt("port")))

	if _, err := url.Parse(viper.GetString("base_url")); err != nil {
		log.Fatal().Err(err).Msgf("invalid base_url: %v", err)
	}

	viper.SetDefault("k8s_app_name", path.Base(os.Args[0]))
	viper.SetDefault("k8s_namespace", "default")

	if viper.IsSet("k8s_app_name") || viper.IsSet("k8s_namespace") {
		viper.SetDefault("base_url", fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			viper.GetString("k8s_app_name"), viper.GetString("k8s_namespace"), viper.GetInt("port")))
	}
}

func jsonStringSettings() string {
	c := viper.AllSettings()
	bs, err := json.Marshal(c)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to marshal config to JSON: %v", err)
	}
	return string(bs)
}

func main() {
	// output application configuration on startup
	log.Info().
		Str("config", jsonStringSettings()).
		Str("GitCommitSha", GitCommitSha).
		Msgf("%s version %s", viper.GetString("k8s_app_name"), Version)

	router := gin.New()
	router.Use(zlog.Middleware())

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	router.Run(fmt.Sprintf(":%d", viper.GetInt("port")))
}
