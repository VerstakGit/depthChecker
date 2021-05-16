## Binance order book depth checker
This service can be used to find large buy\sell orders in futures\spot binance markets.

### How to use
1. Build service
2. Edit config file
```
keys: # Current version can perfectly work without them, you can left them as is
  api_key: apiKey
  secret_key: secretKey

# Service can work in 2 different modes:
# 1. if "check_tickers_by_vol" = true then service will automatically find the best futures (right now futures market only) tickers for tracking (selection is working by volume)
# In this mode you can select symbols that you don't want to track in "vol_tickers" section
# Section "tickers" doesn't work in this mode.
# 2. if "check_tickers_by_vol" = false then service will track tickers from "tickers" section
# Section "vol_tickers" is ignored in this mode.
check_tickers_by_vol: true

vol_tickers: # Settings for first mode
  symbol_ban_list: # Symbols from this list won't be included in the tracking list
    - BTCUSDT
    - LTCUSDT
    - ETHUSDT
  symbol_cnt: 10 # Number of tickers you want to track
  large_order: 500000 # Large order size (in USD)

tickers: # Settings for second mode
  - symbol: LINKUSDT
    large_order: 100000
    market_type: f

play_sound_alert: true # Make a sound alert on large orders
alert_cooldown: 5m # Set an alert cooldown on exact ticker and price for X interval
futures_depth_limit: 500 # Limit for binance futures order book API (acceptable values: 5,10,20,50,100,500,1000)
spot_depth_limit: 1000 # Limit for binance spot order book API (acceptable values: 5,10,20,50,100,500,1000,5000)
```
3. Exec service `depthChecker --config=path_to_config.yaml`