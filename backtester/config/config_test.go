package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/strategies/base"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/strategies/top2bottom2"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/drivers"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
)

const (
	testExchange = "binance"
	dca          = "dollarcostaverage"
	// change this if you modify a config and want it to save to the example folder
	saveConfig = false
)

var (
	startDate    = time.Date(time.Now().Year()-1, 8, 1, 0, 0, 0, 0, time.Local)
	endDate      = time.Date(time.Now().Year()-1, 12, 1, 0, 0, 0, 0, time.Local)
	tradeEndDate = startDate.Add(time.Hour * 72)
	makerFee     = decimal.NewFromFloat(0.001)
	takerFee     = decimal.NewFromFloat(0.002)
	minMax       = MinMax{
		MinimumSize:  decimal.NewFromFloat(0.005),
		MaximumSize:  decimal.NewFromInt(2),
		MaximumTotal: decimal.NewFromInt(40000),
	}
	initialQuoteFunds1 *decimal.Decimal
	initialQuoteFunds2 *decimal.Decimal
	initialBaseFunds   *decimal.Decimal
)

func TestMain(m *testing.M) {
	iF1 := decimal.NewFromInt(1000000)
	iF2 := decimal.NewFromInt(100000)
	iBF := decimal.NewFromInt(10)
	initialQuoteFunds1 = &iF1
	initialQuoteFunds2 = &iF2
	initialBaseFunds = &iBF
	os.Exit(m.Run())
}

func TestLoadConfig(t *testing.T) {
	_, err := LoadConfig([]byte(`{}`))
	if err != nil {
		t.Error(err)
	}
}

func TestReadConfigFromFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Problem creating temp dir at %s: %s\n", tempDir, err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()
	var passFile *os.File
	passFile, err = ioutil.TempFile(tempDir, "*.start")
	if err != nil {
		t.Fatalf("Problem creating temp file at %v: %s\n", passFile, err)
	}
	_, err = passFile.WriteString("{}")
	if err != nil {
		t.Error(err)
	}
	err = passFile.Close()
	if err != nil {
		t.Error(err)
	}
	_, err = ReadConfigFromFile(passFile.Name())
	if err != nil {
		t.Error(err)
	}
}

func TestPrintSettings(t *testing.T) {
	cfg := Config{
		Nickname: "super fun run",
		Goal:     "To demonstrate rendering of settings",
		StrategySettings: StrategySettings{
			Name: dca,
			CustomSettings: map[string]interface{}{
				"dca-dummy1": 30.0,
				"dca-dummy2": 30.0,
				"dca-dummy3": 30.0,
			},
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds1,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneMin.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          endDate,
				InclusiveEndDate: true,
			},
			CSVData: &CSVData{
				FullPath: "fake",
			},
			LiveData: &LiveData{
				APIKeyOverride:        "",
				APISecretOverride:     "",
				APIClientIDOverride:   "",
				API2FAOverride:        "",
				APISubAccountOverride: "",
				RealOrders:            false,
			},
			DatabaseData: &DatabaseData{
				StartDate:        startDate,
				EndDate:          endDate,
				ConfigOverride:   nil,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	cfg.PrintSetting()
}

func TestGenerateConfigForDCAAPICandles(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCAAPICandles",
		Goal:     "To demonstrate DCA strategy using API candles",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          endDate,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-api-candles.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCAAPICandlesExchangeLevelFunding(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCAAPICandlesExchangeLevelFunding",
		Goal:     "To demonstrate DCA strategy using API candles using a shared pool of funds",
		StrategySettings: StrategySettings{
			Name:                         dca,
			SimultaneousSignalProcessing: true,
			UseExchangeLevelFunding:      true,
			ExchangeLevelFunding: []ExchangeLevelFunding{
				{
					ExchangeName: testExchange,
					Asset:        asset.Spot.String(),
					Currency:     currency.USDT.String(),
					InitialFunds: decimal.NewFromInt(100000),
				},
			},
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.BTC.String(),
				Quote:        currency.USDT.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.ETH.String(),
				Quote:        currency.USDT.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          endDate,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-api-candles-exchange-level-funding.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCAAPITrades(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCAAPITrades",
		Goal:     "To demonstrate running the DCA strategy using API trade data",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      "ftx",
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee:                makerFee,
				TakerFee:                takerFee,
				SkipCandleVolumeFitting: true,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneHour.Duration(),
			DataType: common.TradeStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          tradeEndDate,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide: MinMax{
				MinimumSize:  decimal.NewFromFloat(0.1),
				MaximumSize:  decimal.NewFromInt(1),
				MaximumTotal: decimal.NewFromInt(10000),
			},
			SellSide: MinMax{
				MinimumSize:  decimal.NewFromFloat(0.1),
				MaximumSize:  decimal.NewFromInt(1),
				MaximumTotal: decimal.NewFromInt(10000),
			},
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-api-trades.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCAAPICandlesMultipleCurrencies(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCAAPICandlesMultipleCurrencies",
		Goal:     "To demonstrate running the DCA strategy using the API against multiple currencies candle data",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.ETH.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          endDate,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-api-candles-multiple-currencies.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCAAPICandlesSimultaneousProcessing(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCAAPICandlesSimultaneousProcessing",
		Goal:     "To demonstrate how simultaneous processing can work",
		StrategySettings: StrategySettings{
			Name:                         dca,
			SimultaneousSignalProcessing: true,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds1,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.ETH.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          endDate,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-api-candles-simultaneous-processing.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCALiveCandles(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCALiveCandles",
		Goal:     "To demonstrate live trading proof of concept against candle data",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneMin.Duration(),
			DataType: common.CandleStr,
			LiveData: &LiveData{
				APIKeyOverride:        "",
				APISecretOverride:     "",
				APIClientIDOverride:   "",
				API2FAOverride:        "",
				APISubAccountOverride: "",
				RealOrders:            false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-candles-live.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForRSIAPICustomSettings(t *testing.T) {
	cfg := Config{
		Nickname: "TestGenerateRSICandleAPICustomSettingsStrat",
		Goal:     "To demonstrate the RSI strategy using API candle data and custom settings",
		StrategySettings: StrategySettings{
			Name: "rsi",
			CustomSettings: map[string]interface{}{
				"rsi-low":    30.0,
				"rsi-high":   70.0,
				"rsi-period": 14,
			},
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.ETH.String(),
				Quote:             currency.USDT.String(),
				InitialBaseFunds:  initialBaseFunds,
				InitialQuoteFunds: initialQuoteFunds1,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate:        startDate,
				EndDate:          endDate,
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "rsi-api-candles.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCACSVCandles(t *testing.T) {
	fp := filepath.Join("..", "testdata", "binance_BTCUSDT_24h_2019_01_01_2020_01_01.csv")
	cfg := Config{
		Nickname: "ExampleStrategyDCACSVCandles",
		Goal:     "To demonstrate the DCA strategy using CSV candle data",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			CSVData: &CSVData{
				FullPath: fp,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-csv-candles.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCACSVTrades(t *testing.T) {
	fp := filepath.Join("..", "testdata", "binance_BTCUSDT_24h-trades_2020_11_16.csv")
	cfg := Config{
		Nickname: "ExampleStrategyDCACSVTrades",
		Goal:     "To demonstrate the DCA strategy using CSV trade data",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneMin.Duration(),
			DataType: common.TradeStr,
			CSVData: &CSVData{
				FullPath: fp,
			},
		},
		PortfolioSettings: PortfolioSettings{
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-csv-trades.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForDCADatabaseCandles(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyDCADatabaseCandles",
		Goal:     "To demonstrate the DCA strategy using database candle data",
		StrategySettings: StrategySettings{
			Name: dca,
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide:           minMax,
				SellSide:          minMax,
				Leverage: Leverage{
					CanUseLeverage: false,
				},
				MakerFee: makerFee,
				TakerFee: takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			DatabaseData: &DatabaseData{
				StartDate: startDate,
				EndDate:   endDate,
				ConfigOverride: &database.Config{
					Enabled: true,
					Verbose: false,
					Driver:  "sqlite",
					ConnectionDetails: drivers.ConnectionDetails{
						Host:     "localhost",
						Database: "testsqlite.db",
					},
				},
				InclusiveEndDate: false,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{
				CanUseLeverage: false,
			},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "dca-database-candles.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateConfigForTop2Bottom2(t *testing.T) {
	cfg := Config{
		Nickname: "ExampleStrategyTop2Bottom2",
		Goal:     "To demonstrate a complex strategy using exchange level funding and simultaneous processing of data signals",
		StrategySettings: StrategySettings{
			Name:                         top2bottom2.Name,
			UseExchangeLevelFunding:      true,
			SimultaneousSignalProcessing: true,
			ExchangeLevelFunding: []ExchangeLevelFunding{
				{
					ExchangeName: testExchange,
					Asset:        asset.Spot.String(),
					Currency:     currency.BTC.String(),
					InitialFunds: decimal.NewFromFloat(3),
				},
				{
					ExchangeName: testExchange,
					Asset:        asset.Spot.String(),
					Currency:     currency.USDT.String(),
					InitialFunds: decimal.NewFromInt(10000),
				},
			},
			CustomSettings: map[string]interface{}{
				"mfi-low":    32,
				"mfi-high":   68,
				"mfi-period": 14,
			},
		},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.BTC.String(),
				Quote:        currency.USDT.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.DOGE.String(),
				Quote:        currency.USDT.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.ETH.String(),
				Quote:        currency.BTC.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.LTC.String(),
				Quote:        currency.BTC.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.XRP.String(),
				Quote:        currency.USDT.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
			{
				ExchangeName: testExchange,
				Asset:        asset.Spot.String(),
				Base:         currency.BNB.String(),
				Quote:        currency.BTC.String(),
				BuySide:      minMax,
				SellSide:     minMax,
				Leverage:     Leverage{},
				MakerFee:     makerFee,
				TakerFee:     takerFee,
			},
		},
		DataSettings: DataSettings{
			Interval: kline.OneDay.Duration(),
			DataType: common.CandleStr,
			APIData: &APIData{
				StartDate: startDate,
				EndDate:   endDate,
			},
		},
		PortfolioSettings: PortfolioSettings{
			BuySide:  minMax,
			SellSide: minMax,
			Leverage: Leverage{},
		},
		StatisticSettings: StatisticSettings{
			RiskFreeRate: decimal.NewFromFloat(0.03),
		},
	}
	if saveConfig {
		result, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		p, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(p, "examples", "t2b2-api-candles-exchange-funding.strat"), result, 0770)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestValidateDate(t *testing.T) {
	c := Config{}
	err := c.validateDate()
	if err != nil {
		t.Error(err)
	}
	c.DataSettings = DataSettings{
		DatabaseData: &DatabaseData{},
	}
	err = c.validateDate()
	if !errors.Is(err, errStartEndUnset) {
		t.Errorf("received: %v, expected: %v", err, errStartEndUnset)
	}
	c.DataSettings.DatabaseData.StartDate = time.Now()
	c.DataSettings.DatabaseData.EndDate = c.DataSettings.DatabaseData.StartDate
	err = c.validateDate()
	if !errors.Is(err, errBadDate) {
		t.Errorf("received: %v, expected: %v", err, errBadDate)
	}
	c.DataSettings.DatabaseData.EndDate = c.DataSettings.DatabaseData.StartDate.Add(time.Minute)
	err = c.validateDate()
	if err != nil {
		t.Error(err)
	}
	c.DataSettings.APIData = &APIData{}
	err = c.validateDate()
	if !errors.Is(err, errStartEndUnset) {
		t.Errorf("received: %v, expected: %v", err, errStartEndUnset)
	}
	c.DataSettings.APIData.StartDate = time.Now()
	c.DataSettings.APIData.EndDate = c.DataSettings.APIData.StartDate
	err = c.validateDate()
	if !errors.Is(err, errBadDate) {
		t.Errorf("received: %v, expected: %v", err, errBadDate)
	}
	c.DataSettings.APIData.EndDate = c.DataSettings.APIData.StartDate.Add(time.Minute)
	err = c.validateDate()
	if err != nil {
		t.Error(err)
	}
}

func TestValidateCurrencySettings(t *testing.T) {
	c := Config{}
	err := c.validateCurrencySettings()
	if !errors.Is(err, errNoCurrencySettings) {
		t.Errorf("received: %v, expected: %v", err, errNoCurrencySettings)
	}
	c.CurrencySettings = append(c.CurrencySettings, CurrencySettings{})
	err = c.validateCurrencySettings()
	if !errors.Is(err, errBadInitialFunds) {
		t.Errorf("received: %v, expected: %v", err, errBadInitialFunds)
	}
	leet := decimal.NewFromInt(1337)
	c.CurrencySettings[0].InitialQuoteFunds = &leet
	err = c.validateCurrencySettings()
	if !errors.Is(err, errUnsetCurrency) {
		t.Errorf("received: %v, expected: %v", err, errUnsetCurrency)
	}
	c.CurrencySettings[0].Base = "lol"
	err = c.validateCurrencySettings()
	if !errors.Is(err, errUnsetAsset) {
		t.Errorf("received: %v, expected: %v", err, errUnsetAsset)
	}
	c.CurrencySettings[0].Asset = "lol"
	err = c.validateCurrencySettings()
	if !errors.Is(err, errUnsetExchange) {
		t.Errorf("received: %v, expected: %v", err, errUnsetExchange)
	}
	c.CurrencySettings[0].ExchangeName = "lol"
	err = c.validateCurrencySettings()
	if err != nil {
		t.Error(err)
	}
	c.CurrencySettings[0].MinimumSlippagePercent = decimal.NewFromInt(-1)
	err = c.validateCurrencySettings()
	if !errors.Is(err, errBadSlippageRates) {
		t.Errorf("received: %v, expected: %v", err, errBadSlippageRates)
	}
	c.CurrencySettings[0].MinimumSlippagePercent = decimal.NewFromInt(2)
	c.CurrencySettings[0].MaximumSlippagePercent = decimal.NewFromInt(-1)
	err = c.validateCurrencySettings()
	if !errors.Is(err, errBadSlippageRates) {
		t.Errorf("received: %v, expected: %v", err, errBadSlippageRates)
	}
	c.CurrencySettings[0].MinimumSlippagePercent = decimal.NewFromInt(2)
	c.CurrencySettings[0].MaximumSlippagePercent = decimal.NewFromInt(1)
	err = c.validateCurrencySettings()
	if !errors.Is(err, errBadSlippageRates) {
		t.Errorf("received: %v, expected: %v", err, errBadSlippageRates)
	}
}

func TestValidateMinMaxes(t *testing.T) {
	t.Parallel()
	c := &Config{}
	err := c.validateMinMaxes()
	if err != nil {
		t.Error(err)
	}

	c.CurrencySettings = []CurrencySettings{
		{
			SellSide: MinMax{
				MinimumSize: decimal.NewFromInt(-1),
			},
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errSizeLessThanZero) {
		t.Errorf("received %v expected %v", err, errSizeLessThanZero)
	}
	c.CurrencySettings = []CurrencySettings{
		{
			SellSide: MinMax{
				MaximumTotal: decimal.NewFromInt(-1),
			},
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errSizeLessThanZero) {
		t.Errorf("received %v expected %v", err, errSizeLessThanZero)
	}
	c.CurrencySettings = []CurrencySettings{
		{
			SellSide: MinMax{
				MaximumSize: decimal.NewFromInt(-1),
			},
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errSizeLessThanZero) {
		t.Errorf("received %v expected %v", err, errSizeLessThanZero)
	}

	c.CurrencySettings = []CurrencySettings{
		{
			BuySide: MinMax{
				MinimumSize:  decimal.NewFromInt(2),
				MaximumTotal: decimal.NewFromInt(10),
				MaximumSize:  decimal.NewFromInt(1),
			},
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errMaxSizeMinSizeMismatch) {
		t.Errorf("received %v expected %v", err, errMaxSizeMinSizeMismatch)
	}

	c.CurrencySettings = []CurrencySettings{
		{
			BuySide: MinMax{
				MinimumSize: decimal.NewFromInt(2),
				MaximumSize: decimal.NewFromInt(2),
			},
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errMinMaxEqual) {
		t.Errorf("received %v expected %v", err, errMinMaxEqual)
	}

	c.CurrencySettings = []CurrencySettings{
		{
			BuySide: MinMax{
				MinimumSize:  decimal.NewFromInt(1),
				MaximumTotal: decimal.NewFromInt(10),
				MaximumSize:  decimal.NewFromInt(2),
			},
		},
	}
	c.PortfolioSettings = PortfolioSettings{
		BuySide: MinMax{
			MinimumSize: decimal.NewFromInt(-1),
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errSizeLessThanZero) {
		t.Errorf("received %v expected %v", err, errSizeLessThanZero)
	}
	c.PortfolioSettings = PortfolioSettings{
		SellSide: MinMax{
			MinimumSize: decimal.NewFromInt(-1),
		},
	}
	err = c.validateMinMaxes()
	if !errors.Is(err, errSizeLessThanZero) {
		t.Errorf("received %v expected %v", err, errSizeLessThanZero)
	}
}

func TestValidateStrategySettings(t *testing.T) {
	t.Parallel()
	c := &Config{}
	err := c.validateStrategySettings()
	if !errors.Is(err, base.ErrStrategyNotFound) {
		t.Errorf("received %v expected %v", err, base.ErrStrategyNotFound)
	}
	c.StrategySettings = StrategySettings{Name: dca}
	err = c.validateStrategySettings()
	if !errors.Is(err, nil) {
		t.Errorf("received %v expected %v", err, nil)
	}
	c.StrategySettings.UseExchangeLevelFunding = true
	err = c.validateStrategySettings()
	if !errors.Is(err, errSimultaneousProcessingRequired) {
		t.Errorf("received %v expected %v", err, errSimultaneousProcessingRequired)
	}
	c.StrategySettings.SimultaneousSignalProcessing = true
	err = c.validateStrategySettings()
	if !errors.Is(err, errExchangeLevelFundingDataRequired) {
		t.Errorf("received %v expected %v", err, errExchangeLevelFundingDataRequired)
	}
	c.StrategySettings.ExchangeLevelFunding = []ExchangeLevelFunding{
		{
			InitialFunds: decimal.NewFromInt(-1),
		},
	}
	err = c.validateStrategySettings()
	if !errors.Is(err, errBadInitialFunds) {
		t.Errorf("received %v expected %v", err, errBadInitialFunds)
	}
	c.StrategySettings.UseExchangeLevelFunding = false
	err = c.validateStrategySettings()
	if !errors.Is(err, errExchangeLevelFundingRequired) {
		t.Errorf("received %v expected %v", err, errExchangeLevelFundingRequired)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	c := &Config{
		StrategySettings: StrategySettings{Name: dca},
		CurrencySettings: []CurrencySettings{
			{
				ExchangeName:      testExchange,
				Asset:             asset.Spot.String(),
				Base:              currency.BTC.String(),
				Quote:             currency.USDT.String(),
				InitialBaseFunds:  initialBaseFunds,
				InitialQuoteFunds: initialQuoteFunds2,
				BuySide: MinMax{
					MinimumSize:  decimal.NewFromInt(1),
					MaximumSize:  decimal.NewFromInt(10),
					MaximumTotal: decimal.NewFromInt(10),
				},
			},
		},
	}
	if err := c.Validate(); !errors.Is(err, nil) {
		t.Errorf("received %v expected %v", err, nil)
	}
}
