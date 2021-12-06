package shadow

import (
	"context"
	
	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/sirupsen/logrus"
	"github.com/c9s/bbgo/pkg/types"
)

const ID = "shadow"

var log = logrus.WithField("strategy", ID)

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

type Strategy struct {
	Symbol string `json:"symbol"`

	Interval          types.Interval   `json:"interval"`
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: string(s.Interval)})
}

func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	if s.Interval == "" {
		s.Interval = types.Interval4h
	}


	// buy when price drops -8%
	market, ok := session.Market(s.Symbol)
	if !ok {
		return fmt.Errorf("market %s is not defined", s.Symbol)
	}

	standardIndicatorSet, ok := session.StandardIndicatorSet(s.Symbol)
	if !ok {
		return fmt.Errorf("standardIndicatorSet is nil, symbol %s", s.Symbol)
	}


	session.MarketDataStream.OnKLineClosed(func(kline types.KLine) {
		// skip k-lines from other symbols
		if kline.Symbol != s.Symbol {
			return
		}



		s := (kline.High + kline.Low)/2.0 - kline.Close

		if s > 0.0 {
			_, err := orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
				Symbol:   kline.Symbol,
				Market:   market,
				Side:     types.SideTypeBuy,
				Type:     types.OrderTypeMarket,
				Quantity: 0.001,
			})
			if err != nil {
				log.WithError(err).Error("submit buy order error")
			}
		}
                if s < 0.0 {
                         _, err := orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
                                 Symbol:   kline.Symbol,
                                 Market:   market,
                                 Side:     types.SideTypeSell,
                                 Type:     types.OrderTypeMarket,
                                 Quantity: 0.001,
                         })
                         if err != nil {
                                 log.WithError(err).Error("submit sell order error")
                         }
                 }
	})

	return nil
}

