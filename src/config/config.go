package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"platform_engineer_clone/src/utils/validation"
	"strings"
)

var (
	errValidatingStructParams      = errors.New("error validating struct params")
	errReturnedFromParamValidation = errors.New("errors returned from param validation")
	errTokenDaysValidLessThanOne   = errors.New("error, token days valid is less than one")
)

// DatabaseCredentials holds our database env settings
type DatabaseCredentials struct {
	Host     string `mapstructure:"DB_HOST" validate:"required"`
	User     string `mapstructure:"DB_USERNAME" validate:"required"`
	Pass     string `mapstructure:"DB_PASSWORD" validate:"required"`
	Port     string `mapstructure:"DB_PORT" validate:"required"`
	Database string `mapstructure:"DB_DATABASE" validate:"required"`
}

type App struct {
	TokenDaysValid      int `mapstructure:"APP_TOKEN_DAYS_VALID" validate:"required"`
	RandomCharMinLength int `mapstructure:"APP_RANDOM_CHAR_MIN_LENGTH" validate:"required"`
	RandomCharMaxLength int `mapstructure:"APP_RANDOM_CHAR_MAX_LENGTH" validate:"required"`
}

type API struct {
	Port int `mapstructure:"API_PORT"`
}

type Config struct {
	DatabaseCredentials DatabaseCredentials
	API                 API
	App                 App
}

// NewConfig reads values from the .env file, and writes them to the Config struct
func NewConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error reading config file")
	}
	viper.AutomaticEnv()

	config := &Config{}

	err = viper.Unmarshal(&config.DatabaseCredentials)
	if err != nil {
		return config, errors.Wrap(err, "error trying to unmarshal the database credentials")
	}

	err = viper.Unmarshal(&config.API)
	if err != nil {
		return config, errors.Wrap(err, "error unmarshalling the port")
	}

	err = viper.Unmarshal(&config.App)
	if err != nil {
		return config, errors.Wrap(err, "error unmarshalling the app")
	}
	if config.App.TokenDaysValid < 1 {
		return config, errTokenDaysValidLessThanOne
	}

	configStructs := []interface{}{
		config.DatabaseCredentials,
		config.API,
		config.App,
	}
	for _, configStruct := range configStructs {
		errs, err := validation.ValidateStructParams(&configStruct)
		if err != nil {
			return nil, errors.Wrap(err, errValidatingStructParams.Error())
		}
		if len(errs) > 0 {
			return nil, errors.Wrap(errors.New(strings.Join(errs[:], ",")),
				errReturnedFromParamValidation.Error())
		}
	}

	return config, nil
}
