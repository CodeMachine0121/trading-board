package domains_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
)

func TestFeeScheduleDomainEstimatesTheFeeLeftOut(t *testing.T) {
	cases := []struct {
		name     string
		side     vo.OrderSideVo
		amount   string
		expected string
	}{
		{name: "buy 90 at 52.5 pays 4", side: vo.OrderSideBuy, amount: "4725", expected: "4"},
		{name: "a tiny buy pays the minimum 1", side: vo.OrderSideBuy, amount: "10", expected: "1"},
		{name: "sell 50 at 68 pays fee 3 plus tax 10", side: vo.OrderSideSell, amount: "3400", expected: "13"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			feeSchedule := domains.NewFeeScheduleDomain(defaultFeeSchedule)

			fee := feeSchedule.EstimateFee(testCase.side, money(testCase.amount))

			assert.Equal(t, testCase.expected, fee.String())
		})
	}
}

func TestFeeScheduleDomainTreatsNegativeSettingsAsZero(t *testing.T) {
	feeSchedule := domains.NewFeeScheduleDomain(vo.FeeScheduleVo{
		BrokerFeeRate: money("-0.1"), Discount: money("-1"), MinimumFee: money("-5"), TransactionTaxRate: money("-0.3"),
	})

	fee := feeSchedule.EstimateFee(vo.OrderSideSell, money("3400"))

	assert.Equal(t, "0", fee.String())
}
