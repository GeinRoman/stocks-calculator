package app

import (
	"maps"
	"math"
	"stocks_calculator/internal/model"
)

type (
	rebalanceHelper struct {
		groupsData []*groupData

		totalPriorCost float64
		noSell         bool
	}

	groupData struct {
		targetCost   float64
		priorCost    float64
		weight       int
		shouldChange bool
		stocks       []stock
	}
)

func newRebalanceHelper(
	stocks []model.Stock,
	groups []model.Group,
	valueDiff float64,
	noSell bool,
) *rebalanceHelper {
	helper := rebalanceHelper{noSell: noSell}
	helper.populateGroupTargetCosts(stocks, groups, valueDiff)
	if noSell {
		helper.applyNoSellNormalization(valueDiff)
	}
	return &helper
}

func (helper *rebalanceHelper) populateGroupTargetCosts(stocks []model.Stock, groups []model.Group, valueDiff float64) {
	groupMap := make(map[string]*groupData, len(groups))
	for _, g := range groups {
		groupMap[g.Name] = &groupData{
			targetCost:   0.0,
			priorCost:    0.0,
			weight:       g.Weight,
			shouldChange: true,
			stocks:       []stock{},
		}
	}

	for i := range stocks {
		cost := helper.cost(stocks[i])
		helper.totalPriorCost += cost

		gd := groupMap[stocks[i].GroupName]
		gd.priorCost += cost
		gd.stocks = append(gd.stocks, helper.fromModel(stocks[i]))
	}

	totalTargetCost := helper.totalPriorCost + valueDiff
	for _, g := range groupMap {
		g.targetCost = totalTargetCost * float64(g.weight) / 100
		g.shouldChange = len(g.stocks) > 0
		helper.groupsData = append(helper.groupsData, g)
	}
}

func (helper *rebalanceHelper) applyNoSellNormalization(valueDiff float64) {
	for iter := 0; ; iter++ {
		localPercent := 0
		localCost := valueDiff
		for _, gd := range helper.groupsData {
			if !gd.shouldChange {
				continue
			}
			if gd.priorCost > gd.targetCost {
				gd.shouldChange = false
				continue
			}
			localCost += gd.priorCost
			localPercent += gd.weight
		}

		interrupt := true
		for _, gd := range helper.groupsData {
			if gd.shouldChange {
				gd.targetCost = localCost * float64(gd.weight) / float64(localPercent)
				if gd.priorCost > gd.targetCost {
					interrupt = false
				}
			}
		}
		if interrupt {
			break
		}

		if iter > 1000 {
			panic("applyNoSellNormalization did not converge")
		}
	}

	for _, gd := range helper.groupsData {
		if gd.shouldChange {
			gd.targetCost -= gd.priorCost
		}
	}
}

func (helper *rebalanceHelper) computeAmounts() map[string]int {
	amounts := make(map[string]int)
	for _, v := range helper.groupsData {
		if !v.shouldChange {
			for i := range v.stocks {
				amounts[v.stocks[i].code] = v.stocks[i].originalLotAmount
			}
			continue
		}
		var amount map[string]int
		amount = helper.groupProcessing(v)
		maps.Copy(amounts, amount)
	}
	return amounts
}

func (helper *rebalanceHelper) applyAmounts(stocks []model.Stock, amounts map[string]int) ([]model.Stock, float64) {
	totalEndCost := 0.0
	for i := range stocks {
		stocks[i].LotAmount = amounts[stocks[i].Code]
		totalEndCost += helper.cost(stocks[i])
	}
	return stocks, totalEndCost
}

func (helper *rebalanceHelper) cost(stock model.Stock) float64 {
	return stock.Price * float64(stock.LotAmount) * float64(stock.LotSize)
}

func (helper *rebalanceHelper) computeShareCost(gd *groupData) (float64, map[int]struct{}) {
	shareCost := gd.targetCost / float64(len(gd.stocks))
	skip := map[int]struct{}{}
	if helper.noSell {
		shareCost, skip = helper.getTargetNoSellShareCost(gd)
	}
	return shareCost, skip
}

func (helper *rebalanceHelper) approximateAmounts(gd *groupData, skip map[int]struct{}, shareCost float64) (map[string]int, float64) {
	amounts := make(map[string]int, len(gd.stocks))
	totalFinalCost := 0.0

	for i := range gd.stocks {
		if _, ok := skip[i]; ok {
			amounts[gd.stocks[i].code] = 0
			continue
		}
		if helper.noSell {
			amounts[gd.stocks[i].code] = int(math.Round((shareCost - gd.stocks[i].cost()) / gd.stocks[i].lotPrice))
		} else {
			amounts[gd.stocks[i].code] = int(math.Round(shareCost / gd.stocks[i].lotPrice))
		}
		totalFinalCost += float64(amounts[gd.stocks[i].code]) * gd.stocks[i].lotPrice
	}

	return amounts, totalFinalCost
}

func (helper *rebalanceHelper) groupProcessing(gd *groupData) map[string]int {
	shareCost, skip := helper.computeShareCost(gd)
	amounts, approximatedCost := helper.approximateAmounts(gd, skip, shareCost)
	amounts = helper.minimizeGap(amounts, gd.stocks, approximatedCost-gd.targetCost)
	if helper.noSell {
		for i := range gd.stocks {
			amounts[gd.stocks[i].code] += gd.stocks[i].originalLotAmount
		}
	}
	return amounts
}

func (helper *rebalanceHelper) getTargetNoSellShareCost(gd *groupData) (float64, map[int]struct{}) {
	targetShareCost := (gd.targetCost + gd.priorCost) / float64(len(gd.stocks))
	toChange := len(gd.stocks)
	skip := make(map[int]struct{}, len(gd.stocks)-1)

	for iter := 0; ; iter++ {
		count := 0
		subtract := 0.0
		for i := range gd.stocks {
			if cost := gd.stocks[i].cost(); cost < targetShareCost {
				count++
			} else {
				subtract += cost
				skip[i] = struct{}{}
			}
		}

		if count == toChange {
			return targetShareCost, skip
		}
		toChange = count
		targetShareCost = (gd.targetCost + gd.priorCost - subtract) / float64(count)

		if iter > 1000 {
			panic("getTargetNoSellShareCost did not converge")
		}
	}
}

// minimizeGap adjusts allocations by +/-1 lot to reduce total cost error.
// Greedy: always picks stock with closest unit price to remaining diff.
func (helper *rebalanceHelper) minimizeGap(amounts map[string]int, stocks []stock, diff float64) map[string]int {
	for iter := 0; ; iter++ {
		c := helper.closest(amounts, stocks, diff)

		if c == -1 {
			break
		}

		diffPrior := diff
		if diffPrior > 0 {
			diff = diff - stocks[c].lotPrice
		} else {
			diff = diff + stocks[c].lotPrice
		}

		if math.Abs(diff) >= math.Abs(diffPrior) {
			break
		}

		if diffPrior > 0 {
			amounts[stocks[c].code]--
		} else {
			amounts[stocks[c].code]++
		}

		if iter > 1000 {
			panic("minimizeGap did not converge")
		}
	}

	return amounts
}

func (helper *rebalanceHelper) closest(amounts map[string]int, stocks []stock, target float64) int {
	closest := -1
	negative := target < 0
	diff := math.MaxFloat64
	for i := range stocks {
		d := math.Abs(math.Abs(target) - stocks[i].lotPrice)
		if d < diff && (amounts[stocks[i].code] > 0 || negative) {
			diff = d
			closest = i
		}
	}
	return closest
}

// originalStocks should contain original amounts
func (helper *rebalanceHelper) minimizeGapAcrossAllGroups(
	newAmounts map[string]int,
	originalStocks []model.Stock,
	desiredDiff float64,
	actualDiff float64,
) map[string]int {
	var stocksData []stock
	for i := range originalStocks {
		stocksData = append(stocksData, helper.fromModel(originalStocks[i]))
	}
	if !helper.noSell {
		return helper.minimizeGap(newAmounts, stocksData, actualDiff-desiredDiff)
	}

	appendedAmounts := make(map[string]int, len(newAmounts))
	for _, s := range originalStocks {
		appendedAmounts[s.Code] = newAmounts[s.Code] - s.LotAmount
	}
	appendedAmounts = helper.minimizeGap(appendedAmounts, stocksData, actualDiff-desiredDiff)
	for _, s := range originalStocks {
		newAmounts[s.Code] = appendedAmounts[s.Code] + s.LotAmount
	}
	return newAmounts
}

func (helper *rebalanceHelper) getStocksCost(stocks []model.Stock, amounts map[string]int) float64 {
	totalCost := 0.0
	for i := range stocks {
		stock := stocks[i]
		stock.LotAmount = amounts[stock.Code]
		totalCost += helper.cost(stock)
	}
	return totalCost
}

type stock struct {
	code              string
	lotPrice          float64
	originalLotAmount int
}

func (*rebalanceHelper) fromModel(s model.Stock) stock {
	return stock{
		code:              s.Code,
		lotPrice:          s.Price * float64(s.LotSize),
		originalLotAmount: s.LotAmount,
	}
}

func (s stock) cost() float64 {
	return s.lotPrice * float64(s.originalLotAmount)
}
