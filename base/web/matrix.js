Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { var list = {}, domain = [""]
		msg.Table(function(value) { var name = value.name, _domain = value.domain
			list[name] = list[name]||{}, list[name][_domain] = value, domain.indexOf(_domain) == -1 && domain.push(_domain)
		})
		can.ui = can.page.Appends(can, can._output, [{view: [wiki.CONTENT, html.TABLE], list: [
			{type: html.THEAD, list: [{type: html.TR, list: can.core.List(domain, function(domain) {
				return {type: html.TH, list: [can.onimport.item(can, list[""][domain], list)]}
			}) }]},
			{type: html.TBODY, list: can.core.Item(list, function(name, value) { if (!name) { return }
				return {type: html.TR, list: can.core.List(domain, function(domain) { var item = value[domain]
					return {type: html.TD, list: [item? can.onimport.item(can, item, list): can.onimport.void(can, name, domain, list)]}
				})}
			})},
		] }]), can.onmotion.delay(can, function() { can.Status(mdb.COUNT, can.core.Item(list).length+"x"+can.core.List(domain).length) })
		can.onimport._online(can)
		can.onmotion.orderShow(can, can.page.SelectOne(can, can._output, "table>tbody"), "tr")
	},
	void: function(can, name, domain, list) { var worker = list[name][""], server = list[""][domain]
		return {view: html.ACTION, _init: function(target) {
			worker && can.onappend.input(can, {type: html.BUTTON, name: code.INSTALL, onclick: function(event) {
				can.Update(can.request(event, {name: name, domain: domain}, worker), [ctx.ACTION, code.INSTALL])
			}}, "", target)
		}}
	},
	item: function(can, item, list) { var name = item.name, domain = item.domain, worker = list[name][""], server = list[""][domain]
		item["server.type"] = server.type
		function cb(action) { return function(event) { can.Update(can.request(event, item), [ctx.ACTION, action]) } }
		return {view: [[html.ITEM, item.type, item.status, can.onimport.style(can, item, list)]], list: [
			{img: can.misc.Resource(can, item.icons, can.core.Keys(item.domain, item.name)), onclick: cb(web.DESKTOP)}, {view: wiki.TITLE, list: [
				{text: item.name||item.domain||location.host, onclick: cb(web.OPEN)},
				item.status != cli.STOP && can.onappend.label(can, item, {version: icon.version, time: icon.compile}),
				can.onappend.buttons(can, item),
			]},
		]}
	},
	style: function(can, item, list) { var name = item.name, domain = item.domain, worker = list[name][""]
		return !worker? html.NOTICE: (worker.status != cli.STOP && item.status != cli.STOP && (item.version != worker.version || item.time < worker.time))? html.DANGER: ""
	},
}, [""])
