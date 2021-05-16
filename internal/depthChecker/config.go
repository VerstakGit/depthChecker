package depthChecker

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type Config struct {
	Keys              Keys
	Tickers           []Ticker
	PlaySoundAlert    bool          `yaml:"play_sound_alert"`
	AlertCooldown     time.Duration `yaml:"alert_cooldown"`
	CheckTickersByVol bool          `yaml:"check_tickers_by_vol"`
	VolTickers        VolTickers    `yaml:"vol_tickers"`
	FuturesDepthLimit int           `yaml:"futures_depth_limit"`
	SpotDepthLimit    int           `yaml:"spot_depth_limit"`
}

type VolTickers struct {
	SymbolBanList []string `yaml:"symbol_ban_list"`
	SymbolCnt     int      `yaml:"symbol_cnt"`
	LargeOrder    float64  `yaml:"large_order"`
}

type Keys struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
}

type Ticker struct {
	Symbol     string
	LargeOrder float64 `yaml:"large_order"`
	MarketType string  `yaml:"market_type"` // f - futures, s - spot
}

type ErrUnknownMarketType struct {
	Symbol string
}

func (err ErrUnknownMarketType) Error() string {
	return fmt.Sprintf("uknown market type for %s, market type can be 'f' or 's'", err.Symbol)
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

	for idx := range c.Tickers {
		if c.Tickers[idx].MarketType != futuresMarket && c.Tickers[idx].MarketType != spotMarket {
			return nil, ErrUnknownMarketType{Symbol: c.Tickers[idx].Symbol}
		}
	}

	return c, nil
}
