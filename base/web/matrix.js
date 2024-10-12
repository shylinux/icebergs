Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { var list = {}, domain = [""], server = []
		msg.Table(function(value) { var name = value.name, _domain = value.domain
			list[name] = list[name]||{}, list[name][_domain] = value, domain.indexOf(_domain) == -1 && domain.push(_domain)
			value.type == web.SERVER && server.push(value.domain)
		}), can.db.list = list, can.db.domain = domain, can.db.server = server
		if (domain.length > can.core.Item(list).length) {
			can.ui = can.page.Appends(can, can._output, [{view: [wiki.CONTENT, html.TABLE], list: [
				{type: html.THEAD, list: [{type: html.TR, list: can.core.Item(list, function(name, value) {
					return {type: html.TH, list: [value[""]? can.onimport.item(can, value[""], list): can.onimport.void(can, name, domain, list)]}
				}) }]},
				{type: html.TBODY, list: can.core.List(domain, function(domain) { if (!domain) { return }
					return {type: html.TR, list: can.core.Item(list, function(name, value) { var item = value[domain]
						return {type: html.TD, list: [item? can.onimport.item(can, item, list): can.onimport.void(can, name, domain, list)]}
					})}
				})},
			] }]), can.onmotion.delay(can, function() { can.Status(mdb.COUNT, can.core.List(domain).length+"x"+can.core.Item(list).length) })
		} else {
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
		}
	},
	void: function(can, name, domain, list) { var worker = list[name][""], server = list[""][domain]
		return {view: html.ACTION, _init: function(target) {
			worker && server.type != web.ORIGIN && can.onappend.input(can, {type: html.BUTTON, name: code.INSTALL, onclick: function(event) {
				can.Update(can.request(event, {name: name, domain: domain}, worker), [ctx.ACTION, code.INSTALL])
			}}, "", target)
		}}
	},
	item: function(can, item, list) { var name = item.name, domain = item.domain, worker = list[name][""], server = list[""][domain]; item["server.type"] = server.type
		function cb(action) { return function(event) { can.Update(can.request(event, item), [ctx.ACTION, action]) } }
		return {view: [[html.ITEM, item.type, item.status, can.onimport.style(can, item, list)]], list: [
			{img: item.icons, onclick: cb(web.DESKTOP)},
			{view: wiki.TITLE, list: [
				{text: item.name||item.domain||location.host, onclick: cb(web.OPEN)},
				item.status != cli.STOP && can.onappend.label(can, item, {version: icon.version, time: icon.compile, access: "bi bi-file-lock"}),
				{text: [item.text, "", mdb.STATUS]}, can.onappend.buttons(can, item),
			]},
		], _init: function(target) { item._target = target }}
	},
	style: function(can, item, list) { var name = item.name, domain = item.domain, worker = list[name][""]
		if (worker && worker.module != item.module) { return
			can.core.Item(list, function(key, value) { value = value[""]
				if (value.module == item.module) { worker = value }
			})
		}
		return !worker? html.NOTICE: (worker.status != cli.STOP && item.status != cli.STOP && (item.version != worker.version ||
			(item["server.type"] == "origin"? item.time > worker.time: item.time < worker.time)
		))? html.DANGER: ""
	},
}, [""])
Volcanos(chat.ONACTION, {
	upgrade: function(event, can) { var msg = can.request(event)
		if (msg.Option(mdb.NAME) || msg.Option(web.DOMAIN)) { return can.Update(event, [ctx.ACTION, code.UPGRADE]) }
		can.page.ClassList.add(can, can._output, ice.PROCESS)
		can.core.Next(can.db.server, function(server, next, index) {
			can.core.Next(can.core.Item(can.db.list, function(key, value) { return value }), function(list, next, i) {
				var item = list[server]; if (!item) { return next() } if (!item.name || item.status != cli.START) { return next() }
				can.page.ClassList.add(can, item._target, ice.PROCESS)
				can.Update(can.request({}, item, {_handle: ice.TRUE}), [ctx.ACTION, code.UPGRADE], function(msg) { next() })
			}, next)
		}, function() {
			can.core.Next(can.db.server, function(server, next, index) {
				var item = can.db.list[""][server]; can.page.ClassList.add(can, item._target, ice.PROCESS)
				can.Update(can.request({}, item, {_handle: ice.TRUE}), [ctx.ACTION, code.UPGRADE], function(msg) { next() })
			}, function() { can.onmotion.delay(can, function() {can.Update() }, 3000) })
		})
	},
})
