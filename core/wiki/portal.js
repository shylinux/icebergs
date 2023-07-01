Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.require(["/plugin/local/wiki/word.js"]), can.db = {nav: {}}, can.Conf(html.PADDING, 0)
		can.onmotion.clear(can), can.sup.onexport.link = function() { return "/wiki/portal/" }
		can.ui = can.onappend.layout(can, ["header", ["nav", "main", "aside"]], html.FLOW)
		can.ui.header.innerHTML = msg.Append("header"), can.ui.nav.innerHTML = msg.Append("nav")
		can.db.prefix = location.pathname.indexOf("/wiki/portal/") == 0? "/wiki/portal/": "/chat/cmd/web.wiki.portal/"
		can.db.current = can.isCmdMode()? can.base.trimPrefix(location.pathname, can.db.prefix): can.Option(nfs.PATH)
		if (can.isCmdMode()) { can.onappend.style(can, html.OUTPUT), can.ConfHeight(can.page.height()), can.ConfWidth(can.page.width()) }
		can.page.ClassList.del(can, can._fields, "home")
		if (msg.Append("nav") == "") {
			can.onmotion.hidden(can, can.ui.nav), can.onmotion.hidden(can, can.ui.aside)
			can.db.current == "" && can.onappend.style(can, "home"), can.onimport.content(can, "content.shy")
		}
		can.ui.layout(can.ConfHeight(), can.ConfWidth()), can.ConfHeight(can.ui.main.offsetHeight), can.ConfWidth(can.ui.main.offsetWidth)
		can.page.Select(can, can._output, wiki.STORY_ITEM, function(target) { var meta = target.dataset||{}
			can.core.CallFunc([can.onimport, can.onimport[meta.name]? meta.name: meta.type||target.tagName.toLowerCase()], [can, meta, target])
			meta.style && can.page.style(can, target, can.base.Obj(meta.style))
		})
		var file = nfs.SRC_DOCUMENT+can.db.current+(can.isCmdMode()? can.base.trimPrefix(location.hash, "#"): can.Option(nfs.FILE))
		var nav = can.db.nav[file]; nav && nav.click()
	},
	navmenu: function(can, meta, target) {
		can.onimport.list(can, can.base.Obj(meta.data), function(event, item) {
			can.page.Select(can, target, html.DIV_ITEM, function(target) { target != event.target && can.page.ClassList.del(can, target, html.SELECT) })
			item.list && item.list.length > 0 || can.onaction.route(event, can, item.meta.link)
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
			sub.onimport.size(sub, parseInt(item.height)||can.base.Min(can.ConfHeight()/2, 300, 600), parseInt(item.width)||can.base.Max(width||can.ConfWidth(), 1000))
		}, can.ui.main, target)
	},
	content: function(can, file) {
		can.runActionCommand(event, web.WIKI_WORD, [nfs.SRC_DOCUMENT+can.db.current+file], function(msg) { can.ui.main.innerHTML = msg.Result(), can.onmotion.clear(can, can.ui.aside)
			can.page.Select(can, can.ui.main, wiki.STORY_ITEM, function(target) { var meta = target.dataset||{}
				meta.type == wiki.TITLE && can.onappend.style(can, meta.name, can.onimport.item(can, {name: meta.text}, function(event) { target.scrollIntoView() }, function() {}, can.ui.aside))
				can.core.CallFunc([can.onimport, can.onimport[meta.name]? meta.name: meta.type||target.tagName.toLowerCase()], [can, meta, target, can.ui.main.offsetWidth-80])
				var _meta = can.base.Obj(meta.meta); _meta && _meta.style && can.page.style(can, target, can.base.Obj(_meta.style))
				meta.style && can.page.style(can, target, can.base.Obj(meta.style))
			})
			can.page.Select(can, can.ui.main, "a", function(target) {
				target.innerText = target.innerText || target.href || "http://localhost:9020"
				target.href = target.href || target.innerText
			})
		})
	},
}, [""])
Volcanos(chat.ONACTION, {
	route: function(event, can, route, internal) {
		var link = can.base.trimPrefix(route||"", nfs.SRC_DOCUMENT); if (!link || link == can.db.current) { return }
		if (!internal) {
			if (link == nfs.PS) { return can.isCmdMode()? can.user.jumps(can.db.prefix): (can.Option(nfs.PATH, ""), can.Update()) }
			if (can.base.beginWith(link, web.HTTP, nfs.PS)) { return can.user.opens(link) }
			if (link.indexOf(can.db.current) < 0 || link.endsWith(nfs.PS)) { return can.isCmdMode()? can.user.jumps(can.db.prefix+link): (can.Option(nfs.PATH, link), can.Update()) }
		}
		var file = can.base.trimPrefix(link, can.db.current); can.isCmdMode() && can.user.jumps("#"+file)
		if (can.onmotion.cache(can, function() { return file }, can.ui.main, can.ui.aside)) { return }
		can.onimport.content(can, file)
	},
})
