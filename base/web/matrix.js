Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { var list = {}, domain = {contexts: true}
		msg.Table(function(value) { var name = value.name||ice.CONTEXTS; domain[value.domain] = true, list[name] = list[name]||{}, list[name][value.domain] = value })
		var ui = can.page.Append(can, can._output, [{view: [wiki.CONTENT, html.TABLE], list: [{type: html.THEAD, list: [{type: html.TR, list: can.core.Item(domain, function(domain) {
			var item = list[ice.CONTEXTS][domain]
			return {type: html.TH, list: [{view: [html.ITEM], list: [{img: can.misc.Resource(can, item.icons||nfs.USR_ICONS_ICEBERGS)},
				{view: wiki.TITLE, list: [{text: item.domain}, can.onappend.label(can, item, {version: icon.version, compile: icon.compile}), can.onappend.buttons(can, item)]},
			]}]}
		}) }]}, {type: html.TBODY}] }])
		can.core.Item(list, function(name, value) { var i = 0; if (name == ice.CONTEXTS) { return }
			can.page.Append(can, ui.tbody, [{type: html.TR, list: can.core.Item(domain, function(domain) { i++
				var item = value[domain]||{}
				return {type: html.TD, list: [{view: [[html.ITEM, can.core.Value(list, can.core.Keys(name, domain, nfs.VERSION)) != can.core.Value(list, can.core.Keys(name, ice.CONTEXTS, nfs.VERSION))? "danger": ""]], list: item.name? [
					{img: can.misc.Resource(can, item.icons||nfs.USR_ICONS_VOLCANOS)},
					{view: wiki.TITLE, list: [{text: item.name}, can.onappend.label(can, item, {version: icon.version, time: icon.compile, time: icon.compile}), can.onappend.buttons(can, item)]},
				]: [
					{view: html.ACTION, _init: function(target) { var worker = list[name][ice.CONTEXTS], server = list[ice.CONTEXTS][domain]
						can.onappend.input(can, {type: html.BUTTON, name: code.INSTALL, onclick: function(event) {
							can.Update(can.request(event, {name: name, domain: domain, port: server.port}, worker), [ctx.ACTION, code.INSTALL])
						}}, "", target)
					}},
				]}]}
			}) }])
		}), can.onmotion.delay(can, function() { can.Status(mdb.COUNT, can.core.Item(list).length+"x"+can.core.Item(domain).length) })
	},
}, [""])
