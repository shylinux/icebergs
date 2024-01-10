Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.isCmdMode() && (can.user.toast = function() {})
		var p = "/cmd/"+web.PORTAL; can.db.prefix = location.pathname.indexOf(p) > -1? location.pathname.split(p)[0]+p: nfs.WIKI_PORTAL
		var p = "/c/"+web.PORTAL; can.db.prefix = location.pathname.indexOf(p) > -1? location.pathname.split(p)[0]+p: nfs.WIKI_PORTAL
		can.db.current = can.isCmdMode()? can.base.trimPrefix(location.pathname, can.db.prefix+nfs.PS, can.db.prefix): can.Option(nfs.PATH)
		can.sup.onexport.link = function() { return can.misc.MergeURL(can, {pod: can.ConfSpace(), cmd: can.ConfIndex()}) }
		can.require(["/plugin/local/wiki/word.js"])
		can.Conf(html.PADDING, can.page.styleValueInt(can, "--portal-main-padding", can._output))
		can.onmotion.clear(can), can.isCmdMode() && can.onappend.style(can, html.OUTPUT)
		can.ui = can.onappend.layout(can, [html.HEADER, [html.NAV, html.MAIN, html.ASIDE]], html.FLOW), can.onimport._scroll(can)
		can.ui.header.innerHTML = msg.Append(html.HEADER), can.ui.nav.innerHTML = msg.Append(html.NAV)
		if (msg.Append(html.NAV) == "") {
			can.onmotion.hidden(can, can.ui.nav), can.onmotion.hidden(can, can.ui.aside)
			can.base.isIn(can.db.current, "", nfs.PS) && can.onappend.style(can, ice.HOME), can.onimport.content(can, "content.shy")
		} else {
			can.page.styleWidth(can, can.ui.nav, 230), can.page.styleWidth(can, can.ui.aside, 200)
			if (can.ConfWidth() < 1000) { can.onmotion.hidden(can, can.ui.aside) }
			can.page.ClassList.del(can, can._fields, ice.HOME)
		}
		can.db.nav = {}
		can.page.Select(can, can._output, wiki.STORY_ITEM, function(target) { var meta = target.dataset||{}
			can.core.CallFunc([can.onimport, can.onimport[meta.name]? meta.name: meta.type||target.tagName.toLowerCase()], [can, meta, target])
			meta.style && can.page.style(can, target, can.base.Obj(meta.style))
		})
		var file = can.isCmdMode()? can.base.trimPrefix(location.hash, "#"): can.Option(nfs.FILE)
		can.base.beginWith(file, nfs.SRC, nfs.USR) || (file = can.db.current+file)
		var nav = can.db.nav[file]; nav && nav.click()
	},
	_scroll: function(can) { can.ui.main.onscroll = function(event) { var top = can.ui.main.scrollTop, select
		can.page.SelectChild(can, can.ui.main, "h1,h2,h3", function(target) { if (!select && target.offsetTop > top) {
			select = target, can.onmotion.select(can, can.ui.aside, html.DIV_ITEM, target._menu)
		} })
	} },
	navmenu: function(can, meta, target) { var link
		can.onimport.list(can, can.base.Obj(meta.data), function(event, item) {
			can.page.Select(can, target, html.DIV_ITEM, function(target) { target != event.target && can.page.ClassList.del(can, target, html.SELECT) })
			item.list && item.list.length > 0 || can.onaction.route(event, can, item.meta.link)
		}, target, can.page.ClassList.has(can, target.parentNode, html.HEADER)? function(target, item) {
			item.meta.link == nfs.SRC_DOCUMENT+can.db.current && can.onappend.style(can, html.SELECT, target)
		}: function(target, item) { can.db.nav[can.base.trimPrefix(item.meta.link, nfs.SRC_DOCUMENT)] = target
			location.hash || item.list && item.list.length > 0 || link || (link = can.onaction.route({}, can, item.meta.link, true))
		})
	},
	button: function(can, meta, target) { var item = can.base.Obj(meta.meta)
		target.onclick = function(event) { can.onaction.route(event, can, item.route) }
	},
	content: function(can, file) {
		can.runActionCommand(event, web.WIKI_WORD, [(can.base.beginWith(file, nfs.USR, nfs.SRC)? "": nfs.SRC_DOCUMENT+can.db.current)+file], function(msg) { can.ui.main.innerHTML = msg.Result(), can.onmotion.clear(can, can.ui.aside)
			can.onimport._content(can, can.ui.main, function(target, meta) {
				meta.type == wiki.TITLE && can.onappend.style(can, meta.name, target._menu = can.onimport.item(can, {name: meta.text}, function(event) { target.scrollIntoView() }, function() {}, can.ui.aside))
			}), can.onmotion.select(can, can.ui.aside, html.DIV_ITEM, 0)
			can.sup.onimport.size(can.sup, can.sup.ConfHeight(), can.sup.ConfWidth())
		})
	},
	layout: function(can, height, width) { can.onmotion.delay(can, function() { padding = can.Conf(html.PADDING)
		can.ui.layout(height, width), can.ConfHeight(can.ui.main.offsetHeight), can.ConfWidth(can.ui.main.offsetWidth)
		can.Conf(html.PADDING, can.page.styleValueInt(can, "--portal-main-padding", can._output))
		if (can.user.isMobile && can.isCmdMode()) {
			can.page.style(can, can.ui.nav, html.HEIGHT, "", html.WIDTH, can.page.width())
			can.page.style(can, can.ui.main, html.HEIGHT, "", html.WIDTH, can.page.width())
		}
		can.core.List(can._plugins, function(sub) { sub.onimport.size(sub, can.base.Min(can.ConfHeight()/2, 300, 600), sub.Conf("_width")||(can.ConfWidth()-2*padding), true) })
	}, 10) },
}, [""])
Volcanos(chat.ONACTION, {
	route: function(event, can, route, internal) {
		var link = can.base.trimPrefix(route||"", nfs.SRC_DOCUMENT); if (!link || link == can.db.current) { return }
		if (!internal) { var params = ""; (can.misc.Search(can, log.DEBUG) == ice.TRUE && (params = "?debug=true"))
			if (link == nfs.PS) { return can.isCmdMode()? can.user.jumps(can.db.prefix+params): (can.Option(nfs.PATH, ""), can.Update()) }
			if (can.base.beginWith(link, web.HTTP, nfs.PS)) { return can.user.opens(link) }
			if (!can.base.beginWith(link, nfs.SRC, nfs.USR)) {
				if (link.indexOf(can.db.current) < 0 || link.endsWith(nfs.PS)) {
					return can.isCmdMode()? can.user.jumps(can.base.Path(can.db.prefix, link)+params): (can.Option(nfs.PATH, link), can.Update())
				}
			}
		}
		var file = can.base.trimPrefix(link, can.db.current); can.isCmdMode() && can.user.jumps("#"+file)
		if (can.onmotion.cache(can, function(save, load) { save({plugins: can._plugins})
			return load(file, function(bak) { can._plugins = bak.file })
		}, can.ui.main, can.ui.aside)) { return file }
		can.onimport.content(can, file)
		can.user.toast(can, "加载成功")
		return link
	},
})
