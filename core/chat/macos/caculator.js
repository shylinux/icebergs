Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		var list = [
			["AC", "+/-", "%", "/"],
			["7", "8", "9", "*"],
			["4", "5", "6", "-"],
			["1", "2", "3", "+"],
			["0", "00", ".", "="],
		]
		can.ui.display = can.page.Append(can, can._output, [{view: "display", inner: "0"}])._target
		var table = can.page.Append(can, can._output, [{type: html.TABLE}])._target
		can.page.Append(can, table, can.core.List(list, function(list) {
			return {type: html.TR, list: can.core.List(list, function(item) {
				return {type: html.TD, inner: item, onclick: function(event) {
					var cb = can.ondetail[item]; cb? cb(event, can, item): (
						can.ui.display.innerHTML = can.base.trimPrefix(can.ui.display.innerHTML + item, "0")
					)
				}}
			}) }
		}))
	},
	_show: function(can) {
	},
}, [""])
Volcanos(chat.ONACTION, {
	onkeydown: function(event, can) {
		switch (event.key) {
			case "=":
				can.ondetail[event.key](evnt, can, event.key)
			case "Shift":
			case "Backspace":
				break
			default:
				can.ui.display.innerHTML += event.key
		}
	},
})
Volcanos(chat.ONDETAIL, {
	"AC": function(event, can, button) {
		can.ui.display.innerHTML = "0"
	},
	"=": function(event, can, button) {
		var list = []
		can.core.List(can.core.Split(can.ui.display.innerHTML, "", "+-*/%"), function(item) {
			switch (item) {
				case "+":
				case "-":
				case "*":
				case "/":
				case "%":
				default: list.push(item)
			}
		})
	},
	"+/-": function(event, can, button) {
	},
})
