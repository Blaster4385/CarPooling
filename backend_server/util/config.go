package util

import "github.com/spf13/viper"

type Config struct {
	MongoUri      string `mapstructure:"MONGO_URI"`
	DBName        string `mapstructure:"DB_NAME"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
	TokenSecret   string `mapstructure:"TOKEN_SECRET"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return config, err
}