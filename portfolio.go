package trader

import (
	"context"

	"github.com/crowemi-io/crowemi-go-utils/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Portfolio struct {
	ID               primitive.ObjectID `bson:"_id, omitempty"`
	Type             string             `bson:"type" json:"type" omitempty:"true"`
	Name             string             `bson:"name" json:"name" omitempty:"true"`
	AllocationGroup  []AllocationGroup  `bson:"allocation_group" omitempty:"true"`
	CurrentCostBasis float64            `bson:"-"`
}

type AllocationGroup struct {
	Type             string       `bson:"type" json:"type" omitempty:"true"`
	Percent          float64      `bson:"percent" json:"percent" omitempty:"true"`
	Allocation       []Allocation `bson:"allocation" json:"allocation" omitempty:"true"`
	CurrentCostBasis float64      `bson:"-"`
}

type Allocation struct {
	Symbol  string            `bson:"symbol" json:"symbol" omitempty:"true"`
	Percent float64           `bson:"percent" json:"percent" omitempty:"true"`
	Group   string            `bson:"group" json:"group" omitempty:"true"`
	Current CurrentAllocation `bson:"-"`
}

type CurrentAllocation struct {
	CostBasis   float64
	Percent     float64
	PercentDiff float64
	HasIncrease bool
}

func GetPortfolio(mongoClient *db.MongoClient, alpacaClient *Alpaca, filters []db.MongoFilter, includeCurrentAllocation bool) ([]Portfolio, error) {
	res, err := db.GetMany[Portfolio](context.TODO(), mongoClient, "portfolios", filters, nil)
	if err != nil {
		return nil, err
	}

	if includeCurrentAllocation {
		for portfolio_index, portfolio := range *res {
			// TODO: calculate total percent for portfolio, should be 100%, WARN!
			var groupTotalCostBasis float64
			for group_index, allocationGroup := range portfolio.AllocationGroup {
				for allocation := range allocationGroup.Allocation {
					err := GetCurrentAllocation(alpacaClient, &allocationGroup.Allocation[allocation])
					if err != nil {
						// we dont want to return here
						// return nil, err
						continue
					}
					portfolio.AllocationGroup[group_index].CurrentCostBasis += allocationGroup.Allocation[allocation].Current.CostBasis
					(*res)[portfolio_index].CurrentCostBasis += allocationGroup.Allocation[allocation].Current.CostBasis
				}
			}
			for _, allocationGroup := range portfolio.AllocationGroup {
				// TODO: calculate total percent for group, should be 100%, WARN!
				for allocation := range allocationGroup.Allocation {
					if groupTotalCostBasis > 0 {
						allocationGroup.Allocation[allocation].Current.Percent = (allocationGroup.Allocation[allocation].Current.CostBasis / (*res)[portfolio_index].CurrentCostBasis)
					} else {
						allocationGroup.Allocation[allocation].Current.Percent = 0
					}
				}
			}
		}

	}
	return *res, nil
}

func GetCurrentAllocation(alpacaClient *Alpaca, allocation *Allocation) error {
	p, err := alpacaClient.GetPosition(allocation.Symbol)
	if err != nil {
		return err
	}
	allocation.Current.CostBasis, _ = p.CostBasis.Float64()
	return err
}

func CreatePortfolio(mongoClient *db.MongoClient, portfolio Portfolio) (Portfolio, error) {
	if portfolio.ID == primitive.NilObjectID {
		portfolio.ID = primitive.NewObjectID()
	}
	_, err := db.InsertOne(context.TODO(), mongoClient, "portfolios", portfolio)
	if err != nil {
		return Portfolio{}, err
	}
	return portfolio, nil
}

func GetDefaultPortfolio() Portfolio {
	return Portfolio{
		ID:   primitive.NewObjectID(),
		Type: "DIV",
		Name: "dividend",
		AllocationGroup: []AllocationGroup{
			{
				Type:    "EFT",
				Percent: .6,
				Allocation: []Allocation{
					{
						Symbol:  "QQQI",
						Percent: 0.25,
						Group:   "ETF",
					},
					{
						Symbol:  "SPYI",
						Percent: 0.25,
						Group:   "ETF",
					},
					{
						Symbol:  "JEPI",
						Percent: 0.10,
						Group:   "ETF",
					},
					{
						Symbol:  "JEPQ",
						Percent: 0.10,
						Group:   "ETF",
					},
				},
			},
			{
				Type:    "REIT",
				Percent: .35,
				Allocation: []Allocation{
					{
						Symbol:  "STAG",
						Percent: 0.0625,
						Group:   "REIT",
					},
					{
						Symbol:  "O",
						Percent: 0.0625,
						Group:   "REIT",
					},
					{
						Symbol:  "APLE",
						Percent: 0.0625,
						Group:   "REIT",
					},
					{
						Symbol:  "MAIN",
						Percent: 0.0625,
						Group:   "REIT",
					},
				},
			},
			{
				Type:    "YieldMax",
				Percent: 0.05,
				Allocation: []Allocation{
					{
						Symbol:  "YMAX",
						Percent: 0.0250,
						Group:   "YieldMax",
					},
					{
						Symbol:  "MSTY",
						Percent: 0.0250,
						Group:   "YieldMax",
					},
				},
			},
		},
	}
}
