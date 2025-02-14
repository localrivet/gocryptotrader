package coinbene

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// GetDefaultConfig returns a default exchange config
func (c *Coinbene) GetDefaultConfig() (*config.Exchange, error) {
	c.SetDefaults()
	exchCfg := new(config.Exchange)
	exchCfg.Name = c.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = c.BaseCurrencies

	err := c.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if c.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = c.UpdateTradablePairs(context.TODO(), true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets the basic defaults for Coinbene
func (c *Coinbene) SetDefaults() {
	c.Name = "Coinbene"
	c.Enabled = true
	c.Verbose = true
	c.API.CredentialsValidator.RequiresKey = true
	c.API.CredentialsValidator.RequiresSecret = true

	err := c.StoreAssetPairFormat(asset.Spot, currency.PairStore{
		RequestFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: currency.ForwardSlashDelimiter,
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: currency.ForwardSlashDelimiter,
		},
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	err = c.StoreAssetPairFormat(asset.PerpetualSwap, currency.PairStore{
		RequestFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: currency.DashDelimiter,
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: currency.ForwardSlashDelimiter,
		},
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	c.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:        true,
				TradeFetching:         true,
				OrderbookFetching:     true,
				AccountBalance:        true,
				AutoPairUpdates:       true,
				GetOrder:              true,
				GetOrders:             true,
				CancelOrder:           true,
				CancelOrders:          true,
				SubmitOrder:           true,
				TradeFee:              true,
				CryptoDeposit:         true,
				CryptoWithdrawal:      true,
				MultiChainDeposits:    true,
				MultiChainWithdrawals: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:         true,
				AccountBalance:         true,
				AccountInfo:            true,
				OrderbookFetching:      true,
				TradeFetching:          true,
				KlineFetching:          true,
				Subscribe:              true,
				Unsubscribe:            true,
				AuthenticatedEndpoints: true,
				GetOrders:              true,
				GetOrder:               true,
			},
			WithdrawPermissions: exchange.NoFiatWithdrawals |
				exchange.WithdrawCryptoWithAPIPermission,
			Kline: kline.ExchangeCapabilitiesSupported{
				DateRanges: true,
				Intervals:  true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: map[string]bool{
					kline.OneMin.Word():     true,
					kline.ThreeMin.Word():   true,
					kline.FiveMin.Word():    true,
					kline.FifteenMin.Word(): true,
					kline.ThirtyMin.Word():  true,
					kline.OneHour.Word():    true,
					kline.TwoHour.Word():    true,
					kline.FourHour.Word():   true,
					kline.SixHour.Word():    true,
					kline.TwelveHour.Word(): true,
					kline.OneDay.Word():     true,
					kline.ThreeDay.Word():   true,
					kline.OneWeek.Word():    true,
				},
			},
		},
	}
	c.Requester = request.New(c.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	c.API.Endpoints = c.NewEndpoints()
	err = c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      coinbeneAPIURL,
		exchange.RestSwap:      coinbeneSwapAPIURL,
		exchange.WebsocketSpot: wsContractURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	c.Websocket = stream.New()
	c.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	c.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	c.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (c *Coinbene) Setup(exch *config.Exchange) error {
	if !exch.Enabled {
		c.SetEnabled(false)
		return nil
	}

	err := c.SetupDefaults(exch)
	if err != nil {
		return err
	}

	wsRunningURL, err := c.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	err = c.Websocket.Setup(&stream.WebsocketSetup{
		ExchangeConfig:        exch,
		DefaultURL:            wsContractURL,
		RunningURL:            wsRunningURL,
		Connector:             c.WsConnect,
		Subscriber:            c.Subscribe,
		Unsubscriber:          c.Unsubscribe,
		GenerateSubscriptions: c.GenerateDefaultSubscriptions,
		Features:              &c.Features.Supports.WebsocketCapabilities,
		SortBuffer:            true,
	})
	if err != nil {
		return err
	}

	return c.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
}

// Start starts the Coinbene go routine
func (c *Coinbene) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		c.Run()
		wg.Done()
	}()
}

// Run implements the Coinbene wrapper
func (c *Coinbene) Run() {
	if c.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s. (url: %s).\n",
			c.Name,
			common.IsEnabled(c.Websocket.IsEnabled()),
			c.Websocket.GetWebsocketURL(),
		)
		c.PrintEnabledPairs()
	}

	if !c.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := c.UpdateTradablePairs(context.TODO(), false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s Failed to update tradable pairs. Error: %s",
			c.Name,
			err)
	}
}

// FetchTradablePairs returns a list of exchange tradable pairs
func (c *Coinbene) FetchTradablePairs(ctx context.Context, a asset.Item) ([]string, error) {
	if !c.SupportsAsset(a) {
		return nil, fmt.Errorf("%s does not support asset type %s", c.Name, a)
	}

	var currencies []string
	switch a {
	case asset.Spot:
		pairs, err := c.GetAllPairs(ctx)
		if err != nil {
			return nil, err
		}

		for x := range pairs {
			currencies = append(currencies, pairs[x].Symbol)
		}
	case asset.PerpetualSwap:
		instruments, err := c.GetSwapInstruments(ctx)
		if err != nil {
			return nil, err
		}
		pFmt, err := c.GetPairFormat(asset.PerpetualSwap, false)
		if err != nil {
			return nil, err
		}
		for x := range instruments {
			currencies = append(currencies,
				instruments[x].InstrumentID.Format(pFmt.Delimiter, pFmt.Uppercase).String())
		}
	}
	return currencies, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them
func (c *Coinbene) UpdateTradablePairs(ctx context.Context, forceUpdate bool) error {
	assets := c.GetAssetTypes(false)
	for x := range assets {
		pairs, err := c.FetchTradablePairs(ctx, assets[x])
		if err != nil {
			return err
		}

		p, err := currency.NewPairsFromStrings(pairs)
		if err != nil {
			return err
		}

		err = c.UpdatePairs(p, assets[x], false, forceUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTickers updates the ticker for all currency pairs of a given asset type
func (c *Coinbene) UpdateTickers(ctx context.Context, a asset.Item) error {
	if !c.SupportsAsset(a) {
		return fmt.Errorf("%s does not support asset type %s", c.Name, a)
	}

	allPairs, err := c.GetEnabledPairs(a)
	if err != nil {
		return err
	}

	switch a {
	case asset.Spot:
		tickers, err := c.GetTickers(ctx)
		if err != nil {
			return err
		}

		for i := range tickers {
			var newP currency.Pair
			newP, err = currency.NewPairFromString(tickers[i].Symbol)
			if err != nil {
				return err
			}

			if !allPairs.Contains(newP, true) {
				continue
			}

			err = ticker.ProcessTicker(&ticker.Price{
				Pair:         newP,
				Last:         tickers[i].LatestPrice,
				High:         tickers[i].DailyHigh,
				Low:          tickers[i].DailyLow,
				Bid:          tickers[i].BestBid,
				Ask:          tickers[i].BestAsk,
				Volume:       tickers[i].DailyVolume,
				ExchangeName: c.Name,
				AssetType:    a})
			if err != nil {
				return err
			}
		}
	case asset.PerpetualSwap:
		tickers, err := c.GetSwapTickers(ctx)
		if err != nil {
			return err
		}

		for x := range allPairs {
			fpair, err := c.FormatExchangeCurrency(allPairs[x], a)
			if err != nil {
				return err
			}

			tick, ok := tickers[fpair.String()]
			if !ok {
				log.Warnf(log.ExchangeSys,
					"%s SWAP ticker item was not found",
					c.Name)
				continue
			}

			err = ticker.ProcessTicker(&ticker.Price{
				Pair:         allPairs[x],
				Last:         tick.LastPrice,
				High:         tick.High24Hour,
				Low:          tick.Low24Hour,
				Bid:          tick.BestBidPrice,
				Ask:          tick.BestAskPrice,
				Volume:       tick.Volume24Hour,
				LastUpdated:  tick.Timestamp,
				ExchangeName: c.Name,
				AssetType:    a})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (c *Coinbene) UpdateTicker(ctx context.Context, p currency.Pair, a asset.Item) (*ticker.Price, error) {
	if err := c.UpdateTickers(ctx, a); err != nil {
		return nil, err
	}
	return ticker.GetTicker(c.Name, p, a)
}

// FetchTicker returns the ticker for a currency pair
func (c *Coinbene) FetchTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	if !c.SupportsAsset(assetType) {
		return nil,
			fmt.Errorf("%s does not support asset type %s", c.Name, assetType)
	}

	tickerNew, err := ticker.GetTicker(c.Name, p, assetType)
	if err != nil {
		return c.UpdateTicker(ctx, p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (c *Coinbene) FetchOrderbook(ctx context.Context, p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	if !c.SupportsAsset(assetType) {
		return nil,
			fmt.Errorf("%s does not support asset type %s", c.Name, assetType)
	}

	ob, err := orderbook.Get(c.Name, p, assetType)
	if err != nil {
		return c.UpdateOrderbook(ctx, p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (c *Coinbene) UpdateOrderbook(ctx context.Context, p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:        c.Name,
		Pair:            p,
		Asset:           assetType,
		VerifyOrderbook: c.CanVerifyOrderbook,
	}
	if !c.SupportsAsset(assetType) {
		return book,
			fmt.Errorf("%s does not support asset type %s", c.Name, assetType)
	}

	fpair, err := c.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return book, err
	}

	var tempResp Orderbook
	switch assetType {
	case asset.Spot:
		tempResp, err = c.GetOrderbook(ctx,
			fpair.String(),
			100, // TO-DO: Update this once we support configurable orderbook depth
		)
	case asset.PerpetualSwap:
		tempResp, err = c.GetSwapOrderbook(ctx,
			fpair.String(),
			100, // TO-DO: Update this once we support configurable orderbook depth
		)
	}
	if err != nil {
		return book, err
	}
	for x := range tempResp.Asks {
		item := orderbook.Item{
			Price:  tempResp.Asks[x].Price,
			Amount: tempResp.Asks[x].Amount,
		}
		if assetType == asset.PerpetualSwap {
			item.OrderCount = tempResp.Asks[x].Count
		}
		book.Asks = append(book.Asks, item)
	}
	for x := range tempResp.Bids {
		item := orderbook.Item{
			Price:  tempResp.Bids[x].Price,
			Amount: tempResp.Bids[x].Amount,
		}
		if assetType == asset.PerpetualSwap {
			item.OrderCount = tempResp.Bids[x].Count
		}
		book.Bids = append(book.Bids, item)
	}
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(c.Name, p, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies for the
// Coinbene exchange
func (c *Coinbene) UpdateAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	var info account.Holdings
	balance, err := c.GetAccountBalances(ctx)
	if err != nil {
		return info, err
	}
	var acc account.SubAccount
	for key := range balance {
		c := currency.NewCode(balance[key].Asset)
		hold := balance[key].Reserved
		available := balance[key].Available
		acc.Currencies = append(acc.Currencies,
			account.Balance{
				CurrencyName: c,
				TotalValue:   hold + available,
				Hold:         hold,
			})
	}
	info.Accounts = append(info.Accounts, acc)
	info.Exchange = c.Name

	err = account.Process(&info)
	if err != nil {
		return account.Holdings{}, err
	}

	return info, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (c *Coinbene) FetchAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	acc, err := account.GetHoldings(c.Name, assetType)
	if err != nil {
		return c.UpdateAccountInfo(ctx, assetType)
	}

	return acc, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (c *Coinbene) GetFundingHistory(ctx context.Context) ([]exchange.FundHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetWithdrawalsHistory returns previous withdrawals data
func (c *Coinbene) GetWithdrawalsHistory(_ context.Context, _ currency.Code) (resp []exchange.WithdrawalHistory, err error) {
	return nil, common.ErrNotYetImplemented
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (c *Coinbene) GetRecentTrades(ctx context.Context, p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	var err error
	p, err = c.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	var tradeData Trades
	tradeData, err = c.GetTrades(ctx, p.String(), 100)
	if err != nil {
		return nil, err
	}
	var resp []trade.Data
	for i := range tradeData {
		side := order.Buy
		if tradeData[i].Direction == "sell" {
			side = order.Sell
		}
		resp = append(resp, trade.Data{
			Exchange:     c.Name,
			CurrencyPair: p,
			AssetType:    assetType,
			Side:         side,
			Price:        tradeData[i].Price,
			Amount:       tradeData[i].Volume,
			Timestamp:    tradeData[i].TradeTime,
		})
	}

	err = c.AddTradesToBuffer(resp...)
	if err != nil {
		return nil, err
	}

	sort.Sort(trade.ByDate(resp))
	return resp, nil
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (c *Coinbene) GetHistoricTrades(_ context.Context, _ currency.Pair, _ asset.Item, _, _ time.Time) ([]trade.Data, error) {
	return nil, common.ErrFunctionNotSupported
}

// SubmitOrder submits a new order
func (c *Coinbene) SubmitOrder(ctx context.Context, s *order.Submit) (order.SubmitResponse, error) {
	var resp order.SubmitResponse
	if err := s.Validate(); err != nil {
		return resp, err
	}

	if s.Side != order.Buy && s.Side != order.Sell {
		return resp,
			fmt.Errorf("%s orderside is not supported by this exchange",
				s.Side)
	}

	fpair, err := c.FormatExchangeCurrency(s.Pair, asset.Spot)
	if err != nil {
		return resp, err
	}

	tempResp, err := c.PlaceSpotOrder(ctx,
		s.Price,
		s.Amount,
		fpair.String(),
		s.Side.String(),
		s.Type.String(),
		s.ClientID,
		0)
	if err != nil {
		return resp, err
	}
	resp.IsOrderPlaced = true
	resp.OrderID = tempResp.OrderID
	return resp, nil
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (c *Coinbene) ModifyOrder(ctx context.Context, action *order.Modify) (order.Modify, error) {
	return order.Modify{}, common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (c *Coinbene) CancelOrder(ctx context.Context, o *order.Cancel) error {
	if err := o.Validate(o.StandardCancel()); err != nil {
		return err
	}
	_, err := c.CancelSpotOrder(ctx, o.ID)
	return err
}

// CancelBatchOrders cancels an orders by their corresponding ID numbers
func (c *Coinbene) CancelBatchOrders(ctx context.Context, o []order.Cancel) (order.CancelBatchResponse, error) {
	return order.CancelBatchResponse{}, common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
func (c *Coinbene) CancelAllOrders(ctx context.Context, orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	if err := orderCancellation.Validate(); err != nil {
		return order.CancelAllResponse{}, err
	}

	var resp order.CancelAllResponse
	fpair, err := c.FormatExchangeCurrency(orderCancellation.Pair,
		orderCancellation.AssetType)
	if err != nil {
		return resp, err
	}

	orders, err := c.FetchOpenSpotOrders(ctx, fpair.String())
	if err != nil {
		return resp, err
	}

	tempMap := make(map[string]string)
	for x := range orders {
		_, err := c.CancelSpotOrder(ctx, orders[x].OrderID)
		if err != nil {
			tempMap[orders[x].OrderID] = "Failed"
		} else {
			tempMap[orders[x].OrderID] = "Success"
		}
	}
	resp.Status = tempMap
	return resp, nil
}

// GetOrderInfo returns order information based on order ID
func (c *Coinbene) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	var resp order.Detail
	tempResp, err := c.FetchSpotOrderInfo(ctx, orderID)
	if err != nil {
		return resp, err
	}
	resp.Exchange = c.Name
	resp.ID = orderID
	resp.Pair = currency.NewPairWithDelimiter(tempResp.BaseAsset,
		"/",
		tempResp.QuoteAsset)
	resp.Price = tempResp.OrderPrice
	resp.Date = tempResp.OrderTime
	resp.ExecutedAmount = tempResp.FilledAmount
	resp.Fee = tempResp.TotalFee
	return resp, nil
}

// GetDepositAddress returns a deposit address for a specified currency
func (c *Coinbene) GetDepositAddress(ctx context.Context, curr currency.Code, _, chain string) (*deposit.Address, error) {
	d, err := c.ListDepositAddress(ctx, curr)
	if err != nil {
		return nil, err
	}

	if len(d) == 0 {
		return nil, errors.New("no address(es) returned, please create one via the Coinbene website")
	}

	if chain != "" {
		for x := range d {
			if strings.EqualFold(d[x].Chain, chain) {
				return &deposit.Address{Address: d[x].Address, Tag: d[x].AddressTag}, nil
			}
		}
		return nil, fmt.Errorf("no chain %s found", chain)
	}
	return &deposit.Address{Address: d[0].Address, Tag: d[0].AddressTag}, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (c *Coinbene) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.Withdraw(ctx,
		withdrawRequest.Currency,
		withdrawRequest.Crypto.Address,
		withdrawRequest.Crypto.AddressTag,
		withdrawRequest.Crypto.Chain,
		withdrawRequest.Amount)
	if err != nil {
		return nil, err
	}

	return &withdraw.ExchangeResponse{
		ID: resp.ID,
	}, err
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (c *Coinbene) WithdrawFiatFunds(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (c *Coinbene) WithdrawFiatFundsToInternationalBank(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetActiveOrders retrieves any orders that are active/open
func (c *Coinbene) GetActiveOrders(ctx context.Context, getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := getOrdersRequest.Validate(); err != nil {
		return nil, err
	}

	if len(getOrdersRequest.Pairs) == 0 {
		allPairs, err := c.GetAllPairs(ctx)
		if err != nil {
			return nil, err
		}
		for a := range allPairs {
			p, err := currency.NewPairFromString(allPairs[a].Symbol)
			if err != nil {
				return nil, err
			}
			getOrdersRequest.Pairs = append(getOrdersRequest.Pairs, p)
		}
	}

	var resp []order.Detail
	for x := range getOrdersRequest.Pairs {
		fpair, err := c.FormatExchangeCurrency(getOrdersRequest.Pairs[x],
			asset.Spot)
		if err != nil {
			return nil, err
		}

		var tempData OrdersInfo
		tempData, err = c.FetchOpenSpotOrders(ctx, fpair.String())
		if err != nil {
			return nil, err
		}

		for y := range tempData {
			var tempResp order.Detail
			tempResp.Exchange = c.Name
			tempResp.Pair = getOrdersRequest.Pairs[x]
			tempResp.Side = order.Buy
			if strings.EqualFold(tempData[y].OrderType, order.Sell.String()) {
				tempResp.Side = order.Sell
			}
			tempResp.Date = tempData[y].OrderTime
			if tempResp.Status, err = order.StringToOrderStatus(tempData[y].OrderStatus); err != nil {
				log.Errorf(log.ExchangeSys, "%s %v", c.Name, err)
			}
			tempResp.Price = tempData[y].OrderPrice
			tempResp.Amount = tempData[y].Amount
			tempResp.ExecutedAmount = tempData[y].FilledAmount
			tempResp.RemainingAmount = tempData[y].Amount - tempData[y].FilledAmount
			tempResp.Fee = tempData[y].TotalFee
			resp = append(resp, tempResp)
		}
	}
	return resp, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (c *Coinbene) GetOrderHistory(ctx context.Context, getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	if err := getOrdersRequest.Validate(); err != nil {
		return nil, err
	}

	if len(getOrdersRequest.Pairs) == 0 {
		allPairs, err := c.GetAllPairs(ctx)
		if err != nil {
			return nil, err
		}

		for a := range allPairs {
			p, err := currency.NewPairFromString(allPairs[a].Symbol)
			if err != nil {
				return nil, err
			}
			getOrdersRequest.Pairs = append(getOrdersRequest.Pairs, p)
		}
	}

	var resp []order.Detail
	var tempData OrdersInfo
	for x := range getOrdersRequest.Pairs {
		fpair, err := c.FormatExchangeCurrency(getOrdersRequest.Pairs[x],
			asset.Spot)
		if err != nil {
			return nil, err
		}

		tempData, err = c.FetchClosedOrders(ctx, fpair.String(), "")
		if err != nil {
			return nil, err
		}

		for y := range tempData {
			var tempResp order.Detail
			tempResp.Exchange = c.Name
			tempResp.Pair = getOrdersRequest.Pairs[x]
			tempResp.Side = order.Buy
			if strings.EqualFold(tempData[y].OrderType, order.Sell.String()) {
				tempResp.Side = order.Sell
			}
			tempResp.Date = tempData[y].OrderTime
			if tempResp.Status, err = order.StringToOrderStatus(tempData[y].OrderStatus); err != nil {
				log.Errorf(log.ExchangeSys, "%s %v", c.Name, err)
			}
			tempResp.Price = tempData[y].OrderPrice
			tempResp.AverageExecutedPrice = tempData[y].AvgPrice
			tempResp.Amount = tempData[y].Amount
			tempResp.ExecutedAmount = tempData[y].FilledAmount
			tempResp.RemainingAmount = tempData[y].Amount - tempData[y].FilledAmount
			tempResp.Cost = tempData[y].Quantity
			tempResp.CostAsset = tempResp.Pair.Quote
			tempResp.Fee = tempData[y].TotalFee
			tempResp.InferCostsAndTimes()
			resp = append(resp, tempResp)
		}
	}
	return resp, nil
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (c *Coinbene) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	fpair, err := c.FormatExchangeCurrency(feeBuilder.Pair, asset.Spot)
	if err != nil {
		return 0, err
	}

	tempData, err := c.GetPairInfo(ctx, fpair.String())
	if err != nil {
		return 0, err
	}

	if feeBuilder.IsMaker {
		return feeBuilder.PurchasePrice * feeBuilder.Amount * tempData.MakerFeeRate, nil
	}
	return feeBuilder.PurchasePrice * feeBuilder.Amount * tempData.TakerFeeRate, nil
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (c *Coinbene) AuthenticateWebsocket(_ context.Context) error {
	return c.Login()
}

// ValidateCredentials validates current credentials used for wrapper
// functionality
func (c *Coinbene) ValidateCredentials(ctx context.Context, assetType asset.Item) error {
	_, err := c.UpdateAccountInfo(ctx, assetType)
	return c.CheckTransientError(err)
}

// FormatExchangeKlineInterval returns Interval to string
func (c *Coinbene) FormatExchangeKlineInterval(in kline.Interval) string {
	switch in {
	case kline.OneMin, kline.ThreeMin, kline.FiveMin, kline.FifteenMin,
		kline.ThirtyMin, kline.OneHour, kline.TwoHour, kline.FourHour, kline.SixHour, kline.TwelveHour:
		return strconv.FormatFloat(in.Duration().Minutes(), 'f', 0, 64)
	case kline.OneDay:
		return "D"
	case kline.OneWeek:
		return "W"
	case kline.OneMonth:
		return "M"
	}
	return ""
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (c *Coinbene) GetHistoricCandles(ctx context.Context, pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	if err := c.ValidateKline(pair, a, interval); err != nil {
		return kline.Item{}, err
	}

	formattedPair, err := c.FormatExchangeCurrency(pair, asset.PerpetualSwap)
	if err != nil {
		return kline.Item{}, err
	}

	var candles CandleResponse
	if a == asset.PerpetualSwap {
		candles, err = c.GetSwapKlines(ctx,
			formattedPair.String(),
			start, end,
			c.FormatExchangeKlineInterval(interval))
	} else {
		candles, err = c.GetKlines(ctx,
			formattedPair.String(),
			start, end,
			c.FormatExchangeKlineInterval(interval))
	}
	if err != nil {
		return kline.Item{}, err
	}

	ret := kline.Item{
		Exchange: c.Name,
		Pair:     pair,
		Interval: interval,
		Asset:    a,
	}

	for x := range candles.Data {
		if len(candles.Data[x]) < 6 {
			return kline.Item{}, errors.New("unexpected candle data length")
		}
		var tempCandle kline.Candle
		tempTime, ok := candles.Data[x][0].(string)
		if !ok {
			return kline.Item{}, errors.New("timestamp conversion failed")
		}
		timestamp, err := time.Parse(time.RFC3339, tempTime)
		if err != nil {
			return kline.Item{}, err
		}
		tempCandle.Time = timestamp
		open, ok := candles.Data[x][1].(string)
		if !ok {
			return kline.Item{}, errors.New("open conversion failed")
		}
		tempCandle.Open, err = strconv.ParseFloat(open, 64)
		if err != nil {
			return kline.Item{}, err
		}
		high, ok := candles.Data[x][2].(string)
		if !ok {
			return kline.Item{}, errors.New("high conversion failed")
		}
		tempCandle.High, err = strconv.ParseFloat(high, 64)
		if err != nil {
			return kline.Item{}, err
		}

		low, ok := candles.Data[x][3].(string)
		if !ok {
			return kline.Item{}, errors.New("low conversion failed")
		}
		tempCandle.Low, err = strconv.ParseFloat(low, 64)
		if err != nil {
			return kline.Item{}, err
		}

		closeTemp, ok := candles.Data[x][4].(string)
		if !ok {
			return kline.Item{}, errors.New("close conversion failed")
		}
		tempCandle.Close, err = strconv.ParseFloat(closeTemp, 64)
		if err != nil {
			return kline.Item{}, err
		}

		vol, ok := candles.Data[x][5].(string)
		if !ok {
			return kline.Item{}, errors.New("vol conversion failed")
		}
		tempCandle.Volume, err = strconv.ParseFloat(vol, 64)
		if err != nil {
			return kline.Item{}, err
		}

		ret.Candles = append(ret.Candles, tempCandle)
	}

	ret.SortCandlesByTimestamp(false)
	return ret, nil
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (c *Coinbene) GetHistoricCandlesExtended(ctx context.Context, pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	return c.GetHistoricCandles(ctx, pair, a, start, end, interval)
}

// GetAvailableTransferChains returns the available transfer blockchains for the specific
// cryptocurrency
func (c *Coinbene) GetAvailableTransferChains(ctx context.Context, cryptocurrency currency.Code) ([]string, error) {
	r, err := c.ListDepositAddress(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}

	var availableChains []string
	for x := range r {
		if r[x].Chain != "" {
			availableChains = append(availableChains, r[x].Chain)
		}
	}
	return availableChains, nil
}
