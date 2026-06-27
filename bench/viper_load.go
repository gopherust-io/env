package bench

import (
	"github.com/spf13/viper"
)

func loadSmallViper() ViperSmallConfig {
	v := newViper()
	var cfg ViperSmallConfig
	_ = v.Unmarshal(&cfg)
	return cfg
}

func loadMediumViper() ViperMediumConfig {
	v := newViper()
	var cfg ViperMediumConfig
	_ = v.Unmarshal(&cfg)
	return cfg
}

func loadLargeViper() ViperLargeConfig {
	v := newViper()
	var cfg ViperLargeConfig
	_ = v.Unmarshal(&cfg)
	return cfg
}

func newViper() *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()
	return v
}
