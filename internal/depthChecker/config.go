package depthChecker

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Keys           Keys
	Tickers        []Ticker
	PlaySoundAlert bool `yaml:"play_sound_alert"`
}

type Keys struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
}

type Ticker struct {
	Symbol     string
	LargeOrder float64 `yaml:"large_order"`
}

func ParseConfig(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
