package crowemi_trades

import "github.com/crowemi-io/crowemi-go-utils/config"

type Client struct {
	Config config.Alpaca
}

func (t *Client) GetPositions()  {}
func (t *Client) GetActivities() {}
func (t *Client) Rebalance()     {}
