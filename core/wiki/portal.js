Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.db = {nav: {}}, can.require(["/plugin/local/wiki/word.js"]), can.Conf(html.PADDING, 0)
		can.ui = can.onappend.layout(can, ["header", ["nav", "main", "aside"]], html.FLOW)
		can.ui.header.innerHTML = msg.Append("header"), can.ui.nav.innerHTML = msg.Append("nav")
		can.db.prefix = location.pathname.indexOf("/wiki/portal/") == 0? "/wiki/portal/": "/chat/cmd/web.wiki.portal/"
		can.db.current = can.isCmdMode()? can.base.trimPrefix(location.pathname, can.db.prefix): can.Option(nfs.PATH)
		if (msg.Append("nav") == "") {
			can.db.current == "" && can.page.style(can, can.ui.main, "padding", "0px", "max-width", "2000px")
			can.onmotion.hidden(can, can.ui.nav), can.onmotion.hidden(can, can.ui.aside)
			can.onimport.content(can, "content.shy")
		}
		if (can.isCmdMode()) { can.onappend.style(can, html.OUTPUT), can.ConfHeight(can.page.height()), can.ConfWidth(can.page.width()) }
		can.onappend.style(can, {height: can.ConfHeight()-can.ui.header.offsetHeight, width: can.ConfWidth()-can.ui.nav.offsetWidth-can.ui.aside.offsetWidth}, can.ui.main)
		can.ConfHeight(can.ui.main.offsetHeight), can.ConfWidth(can.ui.main.offsetWidth)
		can.page.Select(can, can._output, wiki.STORY_ITEM, function(target) { var meta = target.dataset||{}
			can.core.CallFunc([can.onimport, can.onimport[meta.name]? meta.name: meta.type||target.tagName.toLowerCase()], [can, meta, target, can.ConfWidth()])
			meta.style && can.page.style(can, target, can.base.Obj(meta.style))
		})
		var file = nfs.SRC_DOCUMENT+can.db.current+(can.isCmdMode()? can.base.trimPrefix(location.hash, "#"): can.Option(nfs.FILE))
		var nav = can.db.nav[file]; nav && nav.click()
	},
	navmenu: function(can, meta, target) {
		can.onimport.list(can, can.base.Obj(meta.data), function(event, item) {
			can.page.Select(can, target, html.DIV_ITEM, function(target) { target != event.target && can.page.ClassList.del(can, target, html.SELECT) })
			if (item.list && item.list.length > 0) { return }
			can.onaction.route(event, can, item.meta.link)
		}, target, can.page.ClassList.has(can, target.parentNode, "header")? function(target, item) {
			if (item.meta.name == "_") { target.innerHTML = "", can.onappend.style(can, html.SPACE, target) }
		}: function(target, item) { can.db.nav[item.meta.link] = target
			location.hash || item.list && item.list.length > 0 || can.onaction.route({}, can, item.meta.link, true)
		})
	},
	button: function(can, meta, target) { var item = can.base.Obj(meta.meta)
		target.onclick = function(event) { can.onaction.route(event, can, item.route) }
	},
	field: function(can, meta, target, width) { var item = can.base.Obj(meta.meta); item.inputs = item.list, item.feature = item.meta
		can.onappend._init(can, item, [chat.PLUGIN_STATE_JS], function(sub) {
			sub.run = function(event, cmds, cb, silent) { can.runActionCommand(event, item.index, cmds, cb, true) }
			sub.onimport.size(sub, parseInt(item.height)||can.base.Min(can.ConfHeight()/2, 300, 600), can.base.Max(parseInt(item.width)||width||can.ConfWidth(), 1000))
		}, can.ui.main, target)
	},
	image: function(can, meta, target) {
		can.page.style(can, target, html.WIDTH, can.ConfWidth())
	},
	content: function(can, file) {
		can.runActionCommand(event, web.WIKI_WORD, [nfs.SRC_DOCUMENT+can.db.current+file], function(msg) { can.ui.main.innerHTML = msg.Result(), can.onmotion.clear(can, can.ui.aside)
			can.page.Select(can, can.ui.main, wiki.STORY_ITEM, function(target) { var meta = target.dataset||{}
				meta.type == wiki.TITLE && can.onappend.style(can, meta.name, can.onimport.item(can, {name: meta.text}, function(event) { target.scrollIntoView() }, function() {}, can.ui.aside))
				can.core.CallFunc([can.onimport, can.onimport[meta.name]? meta.name: meta.type||target.tagName.toLowerCase()], [can, meta, target, can.ui.main.offsetWidth-80])
				var _meta = can.base.Obj(meta.meta); _meta && _meta.style && can.page.style(can, target, can.base.Obj(_meta.style))
				meta.style && can.page.style(can, target, can.base.Obj(meta.style))
			})
		})
	},
}, [""])
Volcanos(chat.ONACTION, {
	route: function(event, can, route, internal) {
		var link = can.base.trimPrefix(route||"", nfs.SRC_DOCUMENT); if (!link || link == can.db.current) { return }
		if (!internal) {
			if (link == nfs.PS) { return can.user.jumps(can.db.prefix) }
			if (can.base.beginWith(link, web.HTTP, nfs.PS)) { return can.user.opens(link) }
			if (link.indexOf(can.db.current) < 0 || link.endsWith(nfs.PS)) { return can.user.jumps(can.db.prefix+link) }
		}
		var file = can.base.trimPrefix(link, can.db.current); can.user.jumps("#"+file)
		if (can.onmotion.cache(can, function() { return file }, can.ui.main, can.ui.aside)) { return }
		can.onimport.content(can, file)
	},
})
