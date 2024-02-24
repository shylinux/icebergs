Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { var list = {"contexts": {}}, machine = []
		msg.Table(function(value) { var space = value.space, _machine = value.machine;
			machine.indexOf(_machine) == -1 && (machine.push(_machine))
			list[space] = list[space]||{}, list[space][_machine] = value })
		can.ui = can.page.Appends(can, can._output, [{view: [wiki.CONTENT, html.TABLE], list: [
			{type: html.THEAD, list: [{type: html.TR, list: can.core.List(machine, function(machine) {
				return {type: html.TH, list: [can.onimport.item(can, list["contexts"][machine], list)]}
			}) }]},
			{type: html.TBODY, list: can.core.Item(list, function(space, value) { if (space == "contexts") { return }
				return {type: html.TR, list: can.core.List(machine, function(machine) { var item = value[machine]
					return {type: html.TD, list: [item? can.onimport.item(can, item, list): can.onimport.void(can, space, machine, list)]}
				})}
			})},
		] }]), can.onmotion.delay(can, function() { can.Status(mdb.COUNT, can.core.Item(list).length+"x"+can.core.Item(machine).length) })
	},
	void: function(can, space, machine, list) {
		return {view: html.ACTION, _init: function(target) { var worker = list[space]["localhost"], server = list["contexts"][machine]
			worker && can.onappend.input(can, {type: html.BUTTON, name: code.INSTALL, onclick: function(event) {
				can.Update(can.request(event, {space: space, machine: machine}, worker), [ctx.ACTION, code.INSTALL])
			}}, "", target)
		}}
	},
	item: function(can, item, list) {
		function cb(action) { return function(event) { can.Update(can.request(event, item), [ctx.ACTION, action]) } }
		return {view: [[html.ITEM, item.type, item.status, can.onimport.style(can, item, list)]], list: [
			{img: can.misc.Resource(can, item.icons||nfs.USR_ICONS_ICEBERGS, item.machine), onclick: cb(web.DESKTOP)}, {view: wiki.TITLE, list: [
				{text: (item.type == web.SERVER? item.machine: item.space)||location.host, onclick: cb(web.OPEN)},
				item.status != cli.STOP && can.onappend.label(can, item, {version: icon.version, time: icon.compile}),
				can.onappend.buttons(can, item),
			]},
		]}
	},
	style: function(can, item, list) { var space = item.space, machine = item.machine, worker = list[space]["localhost"]
		return !worker? html.NOTICE: (worker.status != cli.STOP && item.status != cli.STOP && (item.version != worker.version || item.time < worker.time))? html.DANGER: ""
	},
}, [""])
