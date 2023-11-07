package mall

import (
	"math"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const LOAN = "loan"

func init() {
	const (
		YEAR      = "year"
		RATE      = "rate"
		MONTH     = "month"
		PAYMENT   = "payment"
		PRESENT   = "present"
		INTEREST  = "interest"
		INTERESTS = "interests"
		PRESENTS  = "presents"
	)
	Index.MergeCommands(ice.Commands{
		LOAN: {Name: "loan auto loan1 loan2", Help: "分期贷款", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
				YEAR, "年数", RATE, "利率", MONTH, "期数", PAYMENT, "月供",
				PRESENT, "本金", INTEREST, "利息", INTERESTS, "累积利息", PRESENTS, "还欠本金",
				AMOUNT, "累积还款",
			)),
		), Actions: ice.MergeActions(ice.Actions{
			"loan1": {Name: "load present=300 year=30 rate=4.2", Help: "等额本息", Hand: func(m *ice.Message, arg ...string) {
				//〔贷款本金×月利率×(1＋月利率)^还款月数〕÷〔(1＋月利率)^还款月数 - 1〕
				number, rate := kit.Float(m.Option(YEAR))*12, kit.Float(m.Option(RATE))/100/12
				present := kit.Float(m.Option(PRESENT)) * 10000
				pow := math.Pow((1 + rate), number)
				p := present * rate * pow / (pow - 1)
				var interests, presents float64
				presents = present
				for i := float64(0); i < number; i++ {
					_p := presents * rate
					interests += _p
					presents -= (present / number)
					m.Push(MONTH, kit.Int(i+1)).Push(PAYMENT, p)
					m.Push(PRESENT, p-_p).Push(INTEREST, _p)
					m.Push(INTERESTS, interests)
					m.Push(PRESENTS, presents)
					m.Push(AMOUNT, p*(i+1))
				}
				m.Status(kit.Dict(
					PAYMENT, kit.Format("%0.2f 元", p),
					PRESENT, kit.Format("%0.2f 万", present/10000),
					INTEREST, kit.Format("%0.2f 万", p*number/10000-present/10000),
					AMOUNT, kit.Format("%0.2f 万", p*number/10000),
					MONTH, kit.Format("%v 期", number),
				))
			}},
			"loan2": {Name: "load present=300 year=30 rate=4.2", Help: "等额本金", Hand: func(m *ice.Message, arg ...string) {
				// 每月还款金额 =（贷款本金 ÷ 还款月数）+（本金 — 已归还本金累计额）×每月利率
				present := kit.Float(m.Option(PRESENT)) * 10000
				number, rate := kit.Float(m.Option(YEAR))*12, kit.Float(m.Option(RATE))/100/12
				var interests, presents, amount, payment float64
				presents = present
				for i := float64(0); i < number; i++ {
					p := present/number + (present-i*(present/number))*rate
					interests += (present - i*(present/number)) * rate
					presents -= present / number
					amount += p
					kit.If(i == 0, func() { payment = p })
					m.Push(MONTH, kit.Int(i+1)).Push(PAYMENT, p)
					m.Push(PRESENT, present/number).Push(INTEREST, (present-i*(present/number))*rate)
					m.Push(INTERESTS, interests)
					m.Push(PRESENTS, presents)
					m.Push(AMOUNT, amount)
				}
				m.Status(kit.Dict(
					PAYMENT, kit.Format("%0.2f 元", payment),
					PRESENT, kit.Format("%0.2f 万", present/10000),
					INTEREST, kit.Format("%0.2f 万", amount/10000-present/10000),
					AMOUNT, kit.Format("%0.2f 万", amount/10000),
					MONTH, kit.Format("%v 期", number),
				))
			}},
		})},
	})
}
