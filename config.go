package trestCommon

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var BaseDirectory string = ""

func LoadConfig() {
	dir, err := os.Getwd()
	if err != nil {
		panic(errors.Wrapf(err, "failed to retrieve working directory"))
	}
	path := dir + "/"
	isConfigLoaded := false
	for i := 0; i < 20; i++ {
		viper.SetConfigName("reference")
		viper.AddConfigPath(path + "/conf/")
		err = viper.MergeInConfig()
		viper.SetConfigName("application")
		viper.AddConfigPath(path + "/conf-override/")
		err2 := viper.MergeInConfig()
		if err == nil {
			if _, ok := err2.(viper.ConfigFileNotFoundError); err2 != nil && !ok {
				panic(errors.Wrapf(err2, "failed to load override configuration from wd: %s/conf-override/application.yml", dir))
			}
			_ = os.Chdir(path)
			isConfigLoaded = true
			BaseDirectory = path
			break
		}
		lastIndex := strings.LastIndex(path, "/")
		if lastIndex < 0 {
			break
		} else {
			path = path[:lastIndex]
		}
	}
	if !isConfigLoaded {
		panic(errors.Errorf("failed to load configuration from wd: %s/conf/reference.yml", dir))
	}
	viper.AutomaticEnv()
	viper.SetEnvPrefix(viper.GetString("app.env_prefix"))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}
