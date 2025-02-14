package report

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/config"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/compliance"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/holdings"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/statistics"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/statistics/currencystatistics"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctkline "github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

const testExchange = "binance"

func TestGenerateReport(t *testing.T) {
	t.Parallel()
	e := testExchange
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Problem creating temp dir at %s: %s\n", tempDir, err)
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}(tempDir)
	d := Data{
		Config:       &config.Config{},
		OutputPath:   filepath.Join("..", "results"),
		TemplatePath: filepath.Join("tpl.gohtml"),
		OriginalCandles: []*gctkline.Item{
			{
				Candles: []gctkline.Candle{
					{
						Time:             time.Now(),
						Open:             1337,
						High:             1337,
						Low:              1337,
						Close:            1337,
						Volume:           1337,
						ValidationIssues: "hello world!",
					},
				},
			},
		},
		EnhancedCandles: []DetailedKline{
			{
				Exchange:  e,
				Asset:     a,
				Pair:      p,
				Interval:  gctkline.OneHour,
				Watermark: "Binance - SPOT - BTC-USDT",
				Candles: []DetailedCandle{
					{
						Time:           time.Now().Add(-time.Hour * 5).Unix(),
						Open:           decimal.NewFromInt(1337),
						High:           decimal.NewFromInt(1339),
						Low:            decimal.NewFromInt(1336),
						Close:          decimal.NewFromInt(1338),
						Volume:         decimal.NewFromInt(3),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(47, 194, 27, 0.8)",
					},
					{
						Time:           time.Now().Add(-time.Hour * 4).Unix(),
						Open:           decimal.NewFromInt(1332),
						High:           decimal.NewFromInt(1332),
						Low:            decimal.NewFromInt(1330),
						Close:          decimal.NewFromInt(1331),
						Volume:         decimal.NewFromInt(2),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(252, 3, 3, 0.8)",
					},
					{
						Time:           time.Now().Add(-time.Hour * 3).Unix(),
						Open:           decimal.NewFromInt(1337),
						High:           decimal.NewFromInt(1339),
						Low:            decimal.NewFromInt(1336),
						Close:          decimal.NewFromInt(1338),
						Volume:         decimal.NewFromInt(3),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(47, 194, 27, 0.8)",
					},
					{
						Time:           time.Now().Add(-time.Hour * 2).Unix(),
						Open:           decimal.NewFromInt(1337),
						High:           decimal.NewFromInt(1339),
						Low:            decimal.NewFromInt(1336),
						Close:          decimal.NewFromInt(1338),
						Volume:         decimal.NewFromInt(3),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(252, 3, 3, 0.8)",
					},
					{
						Time:         time.Now().Unix(),
						Open:         decimal.NewFromInt(1337),
						High:         decimal.NewFromInt(1339),
						Low:          decimal.NewFromInt(1336),
						Close:        decimal.NewFromInt(1338),
						Volume:       decimal.NewFromInt(3),
						VolumeColour: "rgba(47, 194, 27, 0.8)",
					},
				},
			},
			{
				Exchange:  "Bittrex",
				Asset:     a,
				Pair:      currency.NewPair(currency.BTC, currency.USD),
				Interval:  gctkline.OneDay,
				Watermark: "BITTREX - SPOT - BTC-USD - 1d",
				Candles: []DetailedCandle{
					{
						Time:           time.Now().Add(-time.Hour * 5).Unix(),
						Open:           decimal.NewFromInt(1337),
						High:           decimal.NewFromInt(1339),
						Low:            decimal.NewFromInt(1336),
						Close:          decimal.NewFromInt(1338),
						Volume:         decimal.NewFromInt(3),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(47, 194, 27, 0.8)",
					},
					{
						Time:           time.Now().Add(-time.Hour * 4).Unix(),
						Open:           decimal.NewFromInt(1332),
						High:           decimal.NewFromInt(1332),
						Low:            decimal.NewFromInt(1330),
						Close:          decimal.NewFromInt(1331),
						Volume:         decimal.NewFromInt(2),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(252, 3, 3, 0.8)",
					},
					{
						Time:           time.Now().Add(-time.Hour * 3).Unix(),
						Open:           decimal.NewFromInt(1337),
						High:           decimal.NewFromInt(1339),
						Low:            decimal.NewFromInt(1336),
						Close:          decimal.NewFromInt(1338),
						Volume:         decimal.NewFromInt(3),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(47, 194, 27, 0.8)",
					},
					{
						Time:           time.Now().Add(-time.Hour * 2).Unix(),
						Open:           decimal.NewFromInt(1337),
						High:           decimal.NewFromInt(1339),
						Low:            decimal.NewFromInt(1336),
						Close:          decimal.NewFromInt(1338),
						Volume:         decimal.NewFromInt(3),
						MadeOrder:      true,
						OrderDirection: gctorder.Buy,
						OrderAmount:    decimal.NewFromInt(1337),
						Shape:          "arrowUp",
						Text:           "hi",
						Position:       "aboveBar",
						Colour:         "green",
						PurchasePrice:  decimal.NewFromInt(50),
						VolumeColour:   "rgba(252, 3, 3, 0.8)",
					},
					{
						Time:         time.Now().Unix(),
						Open:         decimal.NewFromInt(1337),
						High:         decimal.NewFromInt(1339),
						Low:          decimal.NewFromInt(1336),
						Close:        decimal.NewFromInt(1338),
						Volume:       decimal.NewFromInt(3),
						VolumeColour: "rgba(47, 194, 27, 0.8)",
					},
				},
			},
		},
		Statistics: &statistics.Statistic{
			Funding:      &funding.Report{},
			StrategyName: "testStrat",
			ExchangeAssetPairStatistics: map[string]map[asset.Item]map[currency.Pair]*currencystatistics.CurrencyStatistic{
				e: {
					a: {
						p: &currencystatistics.CurrencyStatistic{
							MaxDrawdown:              currencystatistics.Swing{},
							LowestClosePrice:         decimal.NewFromInt(100),
							HighestClosePrice:        decimal.NewFromInt(200),
							MarketMovement:           decimal.NewFromInt(100),
							StrategyMovement:         decimal.NewFromInt(100),
							RiskFreeRate:             decimal.NewFromInt(1),
							CompoundAnnualGrowthRate: decimal.NewFromInt(1),
							BuyOrders:                1,
							SellOrders:               1,
							FinalHoldings:            holdings.Holding{},
							FinalOrders:              compliance.Snapshot{},
						},
					},
				},
			},
			RiskFreeRate:    decimal.NewFromFloat(0.03),
			TotalBuyOrders:  1337,
			TotalSellOrders: 1330,
			TotalOrders:     200,
			BiggestDrawdown: &statistics.FinalResultsHolder{
				Exchange: e,
				Asset:    a,
				Pair:     p,
				MaxDrawdown: currencystatistics.Swing{
					Highest: currencystatistics.Iteration{
						Time:  time.Now(),
						Price: decimal.NewFromInt(1337),
					},
					Lowest: currencystatistics.Iteration{
						Time:  time.Now(),
						Price: decimal.NewFromInt(137),
					},
					DrawdownPercent: decimal.NewFromInt(100),
				},
				MarketMovement:   decimal.NewFromInt(1377),
				StrategyMovement: decimal.NewFromInt(1377),
			},
			BestStrategyResults: &statistics.FinalResultsHolder{
				Exchange: e,
				Asset:    a,
				Pair:     p,
				MaxDrawdown: currencystatistics.Swing{
					Highest: currencystatistics.Iteration{
						Time:  time.Now(),
						Price: decimal.NewFromInt(1337),
					},
					Lowest: currencystatistics.Iteration{
						Time:  time.Now(),
						Price: decimal.NewFromInt(137),
					},
					DrawdownPercent: decimal.NewFromInt(100),
				},
				MarketMovement:   decimal.NewFromInt(1337),
				StrategyMovement: decimal.NewFromInt(1337),
			},
			BestMarketMovement: &statistics.FinalResultsHolder{
				Exchange: e,
				Asset:    a,
				Pair:     p,
				MaxDrawdown: currencystatistics.Swing{
					Highest: currencystatistics.Iteration{
						Time:  time.Now(),
						Price: decimal.NewFromInt(1337),
					},
					Lowest: currencystatistics.Iteration{
						Time:  time.Now(),
						Price: decimal.NewFromInt(137),
					},
					DrawdownPercent: decimal.NewFromInt(100),
				},
				MarketMovement:   decimal.NewFromInt(1337),
				StrategyMovement: decimal.NewFromInt(1337),
			},
		},
	}
	d.OutputPath = tempDir
	err = d.GenerateReport()
	if err != nil {
		t.Error(err)
	}
}

func TestEnhanceCandles(t *testing.T) {
	t.Parallel()
	tt := time.Now()
	var d Data
	err := d.enhanceCandles()
	if !errors.Is(err, errNoCandles) {
		t.Errorf("received: %v, expected: %v", err, errNoCandles)
	}
	d.AddKlineItem(&gctkline.Item{})
	err = d.enhanceCandles()
	if !errors.Is(err, errStatisticsUnset) {
		t.Errorf("received: %v, expected: %v", err, errStatisticsUnset)
	}
	d.Statistics = &statistics.Statistic{}
	err = d.enhanceCandles()
	if err != nil {
		t.Error(err)
	}

	d.Statistics.ExchangeAssetPairStatistics = make(map[string]map[asset.Item]map[currency.Pair]*currencystatistics.CurrencyStatistic)
	d.Statistics.ExchangeAssetPairStatistics[testExchange] = make(map[asset.Item]map[currency.Pair]*currencystatistics.CurrencyStatistic)
	d.Statistics.ExchangeAssetPairStatistics[testExchange][asset.Spot] = make(map[currency.Pair]*currencystatistics.CurrencyStatistic)
	d.Statistics.ExchangeAssetPairStatistics[testExchange][asset.Spot][currency.NewPair(currency.BTC, currency.USDT)] = &currencystatistics.CurrencyStatistic{}

	d.AddKlineItem(&gctkline.Item{
		Exchange: testExchange,
		Pair:     currency.NewPair(currency.BTC, currency.USDT),
		Asset:    asset.Spot,
		Interval: gctkline.OneDay,
		Candles: []gctkline.Candle{
			{
				Time:   tt,
				Open:   1336,
				High:   1338,
				Low:    1336,
				Close:  1337,
				Volume: 1337,
			},
		},
	})
	err = d.enhanceCandles()
	if err != nil {
		t.Error(err)
	}

	d.AddKlineItem(&gctkline.Item{
		Exchange: testExchange,
		Pair:     currency.NewPair(currency.BTC, currency.USDT),
		Asset:    asset.Spot,
		Interval: gctkline.OneDay,
		Candles: []gctkline.Candle{
			{
				Time:   tt,
				Open:   1336,
				High:   1338,
				Low:    1336,
				Close:  1336,
				Volume: 1337,
			},
			{
				Time:   tt,
				Open:   1336,
				High:   1338,
				Low:    1336,
				Close:  1335,
				Volume: 1337,
			},
		},
	})

	err = d.enhanceCandles()
	if err != nil {
		t.Error(err)
	}

	d.Statistics.ExchangeAssetPairStatistics[testExchange][asset.Spot][currency.NewPair(currency.BTC, currency.USDT)].FinalOrders = compliance.Snapshot{
		Orders: []compliance.SnapshotOrder{
			{
				ClosePrice:          decimal.NewFromInt(1335),
				VolumeAdjustedPrice: decimal.NewFromInt(1337),
				SlippageRate:        decimal.NewFromInt(1),
				CostBasis:           decimal.NewFromInt(1337),
				Detail:              nil,
			},
		},
		Timestamp: tt,
	}
	err = d.enhanceCandles()
	if err != nil {
		t.Error(err)
	}

	d.Statistics.ExchangeAssetPairStatistics[testExchange][asset.Spot][currency.NewPair(currency.BTC, currency.USDT)].FinalOrders = compliance.Snapshot{
		Orders: []compliance.SnapshotOrder{
			{
				ClosePrice:          decimal.NewFromInt(1335),
				VolumeAdjustedPrice: decimal.NewFromInt(1337),
				SlippageRate:        decimal.NewFromInt(1),
				CostBasis:           decimal.NewFromInt(1337),
				Detail: &gctorder.Detail{
					Date: tt,
					Side: gctorder.Buy,
				},
			},
		},
		Timestamp: tt,
	}
	err = d.enhanceCandles()
	if err != nil {
		t.Error(err)
	}

	d.Statistics.ExchangeAssetPairStatistics[testExchange][asset.Spot][currency.NewPair(currency.BTC, currency.USDT)].FinalOrders = compliance.Snapshot{
		Orders: []compliance.SnapshotOrder{
			{
				ClosePrice:          decimal.NewFromInt(1335),
				VolumeAdjustedPrice: decimal.NewFromInt(1337),
				SlippageRate:        decimal.NewFromInt(1),
				CostBasis:           decimal.NewFromInt(1337),
				Detail: &gctorder.Detail{
					Date: tt,
					Side: gctorder.Sell,
				},
			},
		},
		Timestamp: tt,
	}
	err = d.enhanceCandles()
	if err != nil {
		t.Error(err)
	}

	if len(d.EnhancedCandles) == 0 {
		t.Error("expected enhanced candles")
	}
}
