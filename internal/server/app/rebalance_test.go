package app_test

import (
	"context"
	"math"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
	"stocks_calculator/internal/server/handler"
	"stocks_calculator/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RebalanceUnitTestSuite struct {
	suite.Suite
	app      handler.App
	repoMock *mocks.Repository
	moexMock *mocks.MoexApi
}

func TestRebalanceTestSuite(t *testing.T) {
	suite.Run(t, &RebalanceUnitTestSuite{})
}

func (s *RebalanceUnitTestSuite) TestRebalance_Success() {
	roundFunc := func(x float64) float64 {
		p := math.Pow(10, float64(2))
		return math.Round(x*p) / p
	}

	for _, testCase := range input {
		s.updateReturns()
		response, err := s.app.Rebalance(
			context.Background(),
			1,
			testCase.noSell,
			testCase.valueDiff,
		)

		response.CostDiff = roundFunc(response.CostDiff)
		response.TotalPrevCost = roundFunc(response.TotalPrevCost)

		s.Nil(err)
		s.Equal(testCase.expected, response)
	}
}

func (s *RebalanceUnitTestSuite) updateReturns() {
	var (
		groups = []model.Group{
			{Name: "Technology", Weight: 45},
			{Name: "Finance", Weight: 35},
			{Name: "Energy", Weight: 20},
		}

		prices = []model.MoexInstrumentInfoResult{
			{Price: 189.50, LotSize: 1, Err: nil}, // AAPL
			{Price: 415.20, LotSize: 1, Err: nil}, // MSFT
			{Price: 875.40, LotSize: 1, Err: nil}, // NVDA
			{Price: 198.40, LotSize: 1, Err: nil}, // JPM
			{Price: 278.90, LotSize: 1, Err: nil}, // V
			{Price: 480.60, LotSize: 1, Err: nil}, // MA
			{Price: 112.30, LotSize: 1, Err: nil}, // XOM
			{Price: 158.70, LotSize: 1, Err: nil}, // CVX
			{Price: 32.15, LotSize: 1, Err: nil},  // SHEL
		}

		stocks = []model.Stock{
			{Name: "Apple", Code: "AAPL", GroupName: "Technology", LotAmount: 5, LotSize: 1},
			{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", LotAmount: 3, LotSize: 1},
			{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", LotAmount: 10, LotSize: 1},
			{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", LotAmount: 8, LotSize: 1},
			{Name: "Visa", Code: "V", GroupName: "Finance", LotAmount: 6, LotSize: 1},
			{Name: "Mastercard", Code: "MA", GroupName: "Finance", LotAmount: 4, LotSize: 1},
			{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", LotAmount: 9, LotSize: 1},
			{Name: "Chevron", Code: "CVX", GroupName: "Energy", LotAmount: 6, LotSize: 1},
			{Name: "Shell", Code: "SHEL", GroupName: "Energy", LotAmount: 4, LotSize: 1},
		}
	)

	s.repoMock = &mocks.Repository{}
	s.moexMock = &mocks.MoexApi{}
	s.app = app.New(s.repoMock, &mocks.TokenManager{}, s.moexMock)

	s.moexMock.On("GetInstrumentsInfo", mock.Anything, mock.Anything).Return(prices)
	s.repoMock.On("GetStocks", mock.Anything, mock.Anything).Return(stocks, nil)
	s.repoMock.On("GetGroups", mock.Anything, mock.Anything).Return(groups, nil)
}

var (
	input = []struct {
		valueDiff float64
		noSell    bool
		expected  model.RebalanceResponse
	}{
		{
			valueDiff: 5000.00,
			noSell:    false,
			expected: model.RebalanceResponse{
				Stocks: []model.Stock{
					{Name: "Apple", Code: "AAPL", GroupName: "Technology", Price: 189.50, LotAmount: 19, LotSize: 1},
					{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", Price: 415.20, LotAmount: 8, LotSize: 1},
					{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", Price: 875.40, LotAmount: 4, LotSize: 1},
					{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", Price: 198.40, LotAmount: 14, LotSize: 1},
					{Name: "Visa", Code: "V", GroupName: "Finance", Price: 278.90, LotAmount: 9, LotSize: 1},
					{Name: "Mastercard", Code: "MA", GroupName: "Finance", Price: 480.60, LotAmount: 6, LotSize: 1},
					{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", Price: 112.30, LotAmount: 14, LotSize: 1},
					{Name: "Chevron", Code: "CVX", GroupName: "Energy", Price: 158.70, LotAmount: 10, LotSize: 1},
					{Name: "Shell", Code: "SHEL", GroupName: "Energy", Price: 32.15, LotAmount: 46, LotSize: 1},
				},
				TotalPrevCost: 18221.60,
				CostDiff:      5011.50,
			},
		},
		{
			valueDiff: 500.00,
			noSell:    false,
			expected: model.RebalanceResponse{
				Stocks: []model.Stock{
					{Name: "Apple", Code: "AAPL", GroupName: "Technology", Price: 189.50, LotAmount: 15, LotSize: 1},
					{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", Price: 415.20, LotAmount: 7, LotSize: 1},
					{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", Price: 875.40, LotAmount: 3, LotSize: 1},
					{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", Price: 198.40, LotAmount: 11, LotSize: 1},
					{Name: "Visa", Code: "V", GroupName: "Finance", Price: 278.90, LotAmount: 7, LotSize: 1},
					{Name: "Mastercard", Code: "MA", GroupName: "Finance", Price: 480.60, LotAmount: 5, LotSize: 1},
					{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", Price: 112.30, LotAmount: 11, LotSize: 1},
					{Name: "Chevron", Code: "CVX", GroupName: "Energy", Price: 158.70, LotAmount: 8, LotSize: 1},
					{Name: "Shell", Code: "SHEL", GroupName: "Energy", Price: 32.15, LotAmount: 41, LotSize: 1},
				},
				TotalPrevCost: 18221.60,
				CostDiff:      514.25,
			},
		},
		{
			valueDiff: -1000.00,
			noSell:    false,
			expected: model.RebalanceResponse{
				Stocks: []model.Stock{
					{Name: "Apple", Code: "AAPL", GroupName: "Technology", Price: 189.50, LotAmount: 14, LotSize: 1},
					{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", Price: 415.20, LotAmount: 6, LotSize: 1},
					{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", Price: 875.40, LotAmount: 3, LotSize: 1},
					{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", Price: 198.40, LotAmount: 11, LotSize: 1},
					{Name: "Visa", Code: "V", GroupName: "Finance", Price: 278.90, LotAmount: 7, LotSize: 1},
					{Name: "Mastercard", Code: "MA", GroupName: "Finance", Price: 480.60, LotAmount: 4, LotSize: 1},
					{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", Price: 112.30, LotAmount: 10, LotSize: 1},
					{Name: "Chevron", Code: "CVX", GroupName: "Energy", Price: 158.70, LotAmount: 7, LotSize: 1},
					{Name: "Shell", Code: "SHEL", GroupName: "Energy", Price: 32.15, LotAmount: 36, LotSize: 1},
				},
				TotalPrevCost: 18221.60,
				CostDiff:      -1002.80,
			},
		},
		{
			valueDiff: 5000.00,
			noSell:    true,
			expected: model.RebalanceResponse{
				Stocks: []model.Stock{
					{Name: "Apple", Code: "AAPL", GroupName: "Technology", Price: 189.50, LotAmount: 5, LotSize: 1},
					{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", Price: 415.20, LotAmount: 3, LotSize: 1},
					{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", Price: 875.40, LotAmount: 10, LotSize: 1},
					// existing 8 + 5 = 13; existing 6 + 4 = 10; existing 4 + 1 = 5
					{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", Price: 198.40, LotAmount: 13, LotSize: 1},
					{Name: "Visa", Code: "V", GroupName: "Finance", Price: 278.90, LotAmount: 10, LotSize: 1},
					{Name: "Mastercard", Code: "MA", GroupName: "Finance", Price: 480.60, LotAmount: 5, LotSize: 1},
					// existing 9 + 5 = 14; existing 6 + 3 = 9; existing 4 + 42 = 46
					{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", Price: 112.30, LotAmount: 14, LotSize: 1},
					{Name: "Chevron", Code: "CVX", GroupName: "Energy", Price: 158.70, LotAmount: 9, LotSize: 1},
					{Name: "Shell", Code: "SHEL", GroupName: "Energy", Price: 32.15, LotAmount: 47, LotSize: 1},
				},
				TotalPrevCost: 18221.60,
				CostDiff:      5008.25,
			},
		},
		{
			valueDiff: 900.00,
			noSell:    true,
			expected: model.RebalanceResponse{
				Stocks: []model.Stock{
					{Name: "Apple", Code: "AAPL", GroupName: "Technology", Price: 189.50, LotAmount: 5, LotSize: 1},
					{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", Price: 415.20, LotAmount: 3, LotSize: 1},
					{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", Price: 875.40, LotAmount: 10, LotSize: 1},
					{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", Price: 198.40, LotAmount: 8, LotSize: 1},
					{Name: "Visa", Code: "V", GroupName: "Finance", Price: 278.90, LotAmount: 6, LotSize: 1},
					{Name: "Mastercard", Code: "MA", GroupName: "Finance", Price: 480.60, LotAmount: 4, LotSize: 1},
					// existing 9 + 0 = 9; existing 6 + 0 = 6; existing 4 + 27 = 31
					{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", Price: 112.30, LotAmount: 9, LotSize: 1},
					{Name: "Chevron", Code: "CVX", GroupName: "Energy", Price: 158.70, LotAmount: 6, LotSize: 1},
					{Name: "Shell", Code: "SHEL", GroupName: "Energy", Price: 32.15, LotAmount: 32, LotSize: 1},
				},
				TotalPrevCost: 18221.60,
				CostDiff:      900.2,
			},
		},
		{
			valueDiff: 200.00,
			noSell:    true,
			expected: model.RebalanceResponse{
				Stocks: []model.Stock{
					{Name: "Apple", Code: "AAPL", GroupName: "Technology", Price: 189.50, LotAmount: 5, LotSize: 1},
					{Name: "Microsoft", Code: "MSFT", GroupName: "Technology", Price: 415.20, LotAmount: 3, LotSize: 1},
					{Name: "Nvidia", Code: "NVDA", GroupName: "Technology", Price: 875.40, LotAmount: 10, LotSize: 1},
					{Name: "JPMorgan", Code: "JPM", GroupName: "Finance", Price: 198.40, LotAmount: 8, LotSize: 1},
					{Name: "Visa", Code: "V", GroupName: "Finance", Price: 278.90, LotAmount: 6, LotSize: 1},
					{Name: "Mastercard", Code: "MA", GroupName: "Finance", Price: 480.60, LotAmount: 4, LotSize: 1},
					{Name: "ExxonMobil", Code: "XOM", GroupName: "Energy", Price: 112.30, LotAmount: 9, LotSize: 1},
					// existing 9 + 0 = 9; existing 6 + 0 = 6; existing 4 + 6 = 10
					{Name: "Chevron", Code: "CVX", GroupName: "Energy", Price: 158.70, LotAmount: 6, LotSize: 1},
					{Name: "Shell", Code: "SHEL", GroupName: "Energy", Price: 32.15, LotAmount: 10, LotSize: 1},
				},
				TotalPrevCost: 18221.60,
				CostDiff:      192.9,
			},
		},
	}
)
