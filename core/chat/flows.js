(function() { const ACTION_STORE = "web.flows:action:"
Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) { can.ui = can.onappend.layout(can), cb && cb(msg)
		can.core.Item(can.Action(), function(key) { can.onaction[key] = can.onaction[key]||can.onaction.refresh, can.Action(key, can.misc.localStorage(can, ACTION_STORE+key)) })
		if (can.Option(mdb.ZONE)) { can.onmotion.hidden(can, can.ui.project), can.onimport._content(can, msg, can.Option(mdb.ZONE)) } else { can.onimport._project(can, msg) }
	},
	_project: function(can, msg) { var target; msg.Table(function(value) {
		var item = can.onimport.item(can, value, function(event) { can.Option(mdb.ZONE, value.zone)
			if (can.onmotion.cache(can, function(save, load) {
				save({db: can.db, toggle: can.ui.toggle, _display_class: can.ui.display.className, _profile_class: can.ui.profile.className})
				return load(value.zone, function(bak) { can.db = bak.db, can.ui.toggle = bak.toggle
					can.ui.display.className = bak._display_class||can.ui.display.className
					can.ui.profile.className = bak._profile_class||can.ui.profile.className
				})
			}, can.ui.content, can.ui.display, can._status)) { return can.page.isDisplay(can.ui.profile) && can.onimport._profile(can, can.db.current), can.onimport.layout(can) }
			can.run(event, [value.zone], function(msg) { can.onimport._content(can, msg, can.Option(mdb.ZONE)) })
		}, null, can.ui.project); target = value.zone == can.Option(mdb.ZONE)? item: target||item
	}), target && target.click() },
	_content: function(can, msg, zone) { if (msg.Length() == 0) { return can.Update(can.request({target: can._legend}, {title: mdb.INSERT, zone: zone}), [ctx.ACTION, mdb.INSERT]) }
		can.db.list = {}; msg.Table(function(value) { can.db.list[value.hash] = value })
		var root; can.core.Item(can.db.list, function(key, item) { if (!item.prev && !item.from) { return root = item }
			try { if (item.prev) { can.db.list[item.prev].next = item } if (item.from) { can.db.list[item.from].to = item } } catch(e) { console.log(e) }
		}), can.db.root = root, can.db.current = root
		var _list = can.onexport.travel(can, can.db.root, true), _msg = can.request(); can.core.List(_list, function(item) { _msg.Push(item, msg.append) })
		var table = can.onappend.table(can, _msg, null, can.ui.display); can.page.Select(can, table, "tbody>tr", function(target, index) { _list[index]._tr = target })
		can.onappend._status(can, can.base.Obj(msg.Option(ice.MSG_STATUS)))
		can.core.Item(can.db.list, function(key, item) { if (item.prev) { item.prev = can.db.list[item.prev] } if (item.from) { item.from = can.db.list[item.from] } })
		can.onappend.plugin(can, {index: web.WIKI_DRAW, display: "/plugin/local/wiki/draw.js", style: html.OUTPUT}, function(sub) { can.ui.toggle = can.onappend.toggle(can, can.ui.content)
			sub.onexport.output = function(_sub, _msg) { sub.Action(svg.GO, "manual"), sub.Action(ice.MODE, html.RESIZE), can.onmotion.hidden(can, _sub._action), can.db._content_plugin = _sub
				can.onimport.layout(can), can.onimport._flows(can, _sub)
			}, sub.run = function(event, cmds, cb) { cb(can.request(event)) }
		}, can.ui.content)
	},
	_profile: function(can, item) { if (!item.index) { return can.onmotion.hidden(can, can.ui.profile), can.onimport.layout(can) }
		if (can.onmotion.cache(can, function() { return can.core.Keys(can.Option(mdb.ZONE), item.hash) }, can.ui.profile)) { return can.onmotion.toggle(can, can.ui.profile, true), can.onimport.layout(can) }
		can.onappend.plugin(can, {space: item.space, index: item.index, args: item.args, width: (can.ConfWidth()-can.ui.project.offsetWidth)/2-1}, function(sub) { can.db._profile_plugin = sub
			sub.run = function(event, cmds, cb) { can.runActionCommand(can.request(event, {pod: item.space}), item.index, cmds, function(msg) { can.onmotion.toggle(can, can.ui.profile, true), cb(msg) }) }
			sub.onexport.output = function() { can.onmotion.toggle(can, can.ui.profile, true), can.onimport.layout(can) }
			sub.onaction._close = function() { can.onmotion.hidden(can, can.ui.profile), can.onimport.layout(can)  }
		}, can.ui.profile)
	},
	_flows: async function(can, _sub) { var margin = can.onexport.margin(can), height = can.onexport.height(can), width = can.onexport.width(can)
		var matrix = {}, horizon = can.Action("direct") == "horizon"; can.onmotion.clear(can, _sub.ui.svg), can.db._height = 0, can.db._width = 0
		async function sleep() { return new Promise(resolve => { setTimeout(resolve, can.Action("delay")) }) }
		async function show(item, main) { var prev = "from", from = "prev"; if (horizon) { var prev = "prev", from = "from" }
			while (matrix[can.core.Keys(item.x, item.y)]) {
				if (horizon && main || !horizon && !main) { item.y++
					for(var _item = item[prev]; _item; _item = _item[prev]) { _item.y++
						if (!horizon && _item[prev]) {
							_item._line.Val("y1", _item._line.Val("y1")+height)
						} _item._line.Val("y2", _item._line.Val("y2")+height)
						_item._rect.Val("y", _item._rect.Val("y")+height)
						_item._text.Val("y", _item._text.Val("y")+height)
					}
				} else { item.x++
					for(var _item = item[from]; _item; _item = _item[from]) { _item.x++
						if (horizon && _item[from]) {
							_item._line.Val("x1", _item._line.Val("x1")+width)
						} _item._line.Val("x2", _item._line.Val("x2")+width)
						_item._rect.Val("x", _item._rect.Val("x")+width)
						_item._text.Val("x", _item._text.Val("x")+width)
					}
				}
			} matrix[can.core.Keys(item.x, item.y)] = item
			if (item.from || item.prev) { item._line = _sub.onimport.draw(_sub, {shape: svg.LINE, points:
				horizon && item.from || !horizon && !item.from? [{x: item.x*width+width/2, y: item.y*height-margin}, {x: item.x*width+width/2, y: item.y*height+margin}]:
				[{x: item.x*width-margin, y: item.y*height+height/2}, {x: item.x*width+margin, y: item.y*height+height/2}]
			}) } can.onimport._block(can, _sub, item, item.x*width, item.y*height), await sleep()
			var next = 0, to = 1; if (horizon) { var next = 1, to = 0 }
			if (main) {
				var _item = item.to; if (_item) { _item.x = item.x+to, _item.y = item.y+next, await show(_item) }
				var _item = item.next; if (_item) { _item.x = item.x+next, _item.y = item.y+to, await show(_item, true) }
			} else {
				var _item = item.next; if (_item) { _item.x = item.x+next, _item.y = item.y+to, await show(_item, true) }
				var _item = item.to; if (_item) { _item.x = item.x+to, _item.y = item.y+next, await show(_item) }
			}
		} can.db.root.x = 0, can.db.root.y = 0, await show(can.db.root, true)
	},
	_block: function(can, _sub, item, x, y) { var margin = can.onexport.margin(can), height = can.onexport.height(can), width = can.onexport.width(can)
		var rect = _sub.onimport.draw(_sub, {shape: svg.RECT, points: [{x: x+margin, y: y+margin}, {x: x+width-margin, y: y+height-margin}]})
		var text = _sub.onimport.draw(_sub, {shape: svg.TEXT, points: [{x: x+width/2, y: y+height/2}], style: {inner: item.name||item.index.split(nfs.PT).pop()}})
		item._rect = rect, item._text = text, can.core.ItemCB(can.ondetail, function(key, cb) { text[key] = rect[key] = function(event) { cb(event, can, _sub, item) } })
		if (item.status) { item._line && item._line.Value(html.CLASS, item.status), rect.Value(html.CLASS, item.status), text.Value(html.CLASS, item.status) }
		if (can.db._height < y+height) { can.db._height = y+height, can.onimport.layout(can), can.isCmdMode() && rect.scrollIntoView() }
		if (can.db._width < x+width) { can.db._width = x+width, can.onimport.layout(can), can.isCmdMode() && rect.scrollIntoView() }
	},
	layout: function(can) {
		if (can.page.isDisplay(can.ui.profile)) { var profile = can.db._profile_plugin; can.page.styleWidth(can, can.ui.profile, (can.ConfWidth()-can.ui.project.offsetWidth)/2) }
		can.page.isDisplay(can.ui.display) && can.page.SelectChild(can, can.ui.display, html.TABLE, function(target) { can.page.styleHeight(can, can.ui.display, can.base.Max(target.offsetHeight, can.ConfHeight()/2)+1) })
		can.ui.layout(can.ConfHeight(), can.ConfWidth(), 0, function(height, width) { can.ui.toggle && can.ui.toggle.layout()
			var _sub = can.db._content_plugin; if (_sub) { _sub.sup.onimport.size(_sub.sup, height, width), _sub.ui.svg.Val(html.HEIGHT, can.db._height), _sub.ui.svg.Val(html.WIDTH, can.db._width) }
			profile && profile.onimport.size(profile, can.ui.profile.offsetHeight, can.ui.profile.offsetWidth-1, true)
		})
	},
}, [""])
Volcanos(chat.ONACTION, {
	list: ["create", "play", "prev", "next",
		["travel", "deep", "wide"],
		["direct", "vertical", "horizon"],
		[html.HEIGHT, 80, 100, 120, 140, 200],
		[html.WIDTH, 200, 240, 280, 400],
		[html.MARGIN, 10, 20, 40, 60],
		["delay", 100, 200, 500, 1000],
	], _trans: {play: "播放", prev: "上一步", next: "下一步"},
	refresh: function(event, can, button) { can.misc.localStorage(can, ACTION_STORE+button, can.Action(button)), can.onimport._flows(can, can.db._content_plugin) },
	travel: function() {}, delay: function() {},
	play: function(event, can) { var list = can.onexport.travel(can, can.db.root, true)
		can.core.List(list, function(item) { item._line && item._line.Value(html.CLASS, ""), item._rect.Value(html.CLASS, ""), item._text.Value(html.CLASS, "") })
		can.core.Next(list, function(item, next, index, list) {
			item._line && item._line.Value(html.CLASS, "done"), item._rect.Value(html.CLASS, "done"), item._text.Value(html.CLASS, "done")
			can.user.toast(can, list[index].index), can.ondetail._select(event, can, item), can.onmotion.delay(can, next, 1000)
		}, function() { can.user.toastSuccess(can) })
	},
	prev: function(event, can) { var list = can.onexport.travel(can, can.db.root, true), prev
		if (!can.db.current) { can.db.current = list.pop() } else {
			can.core.List(list, function(item, index) { if (item == can.db.current) { prev = list[index-1] } }), can.db.current = prev
		} can.ondetail._select(event, can, can.db.current)
	},
	next: function(event, can) {
		if (!can.db.current) { can.db.current = can.db.root } else { var next, list = can.onexport.travel(can, can.db.root, true)
			can.core.List(list, function(item, index) { if (item == can.db.current) { next = list[index+1] } }), can.db.current = next
		} can.ondetail._select(event, can, can.db.current)
	},
	show: function(event, can) { can.onmotion.toggle(can, can.ui.profile), can.onimport.layout(can) },
	exec: function(event, can) { can.onmotion.toggle(can, can.ui.display), can.onimport.layout(can) },
	clear: function(event, can) { if (can.onmotion.clearFloat(can)) { return }
		if (can.page.isDisplay(can.ui.profile)) { return can.onmotion.hidden(can, can.ui.profile), can.onimport.layout(can) }
		if (can.page.isDisplay(can.ui.display)) { return can.onmotion.hidden(can, can.ui.display), can.onimport.layout(can) }
		can.onmotion.toggle(can, can.ui.project), can.onimport.layout(can)
	},
	plugin: function(event, can, msg) { can.ondetail._select(event, can, can.db.list[msg.Option(mdb.HASH)]) },
})
Volcanos(chat.ONDETAIL, {
	_select: function(event, can, item) { if (!item) { return can.onmotion.hidden(can, can.ui.profile) }
		can.isCmdMode() && item._rect.scrollIntoView(), can.db.current = item, can.onimport._profile(can, item)
		can.page.Select(can, item._rect.parentNode, "", function(target) { var _class = (target.Value(html.CLASS)||"").split(lex.SP)
			if (can.base.isIn(target, item._line, item._rect, item._text)) {
				if (_class.indexOf(html.SELECT) == -1) { target.Value(html.CLASS, _class.concat([html.SELECT]).join(lex.SP).trim()) }
			} else {
				if (_class.indexOf(html.SELECT) > -1) { target.Value(html.CLASS, _class.filter(function(c) { return c != html.SELECT }).join(lex.SP).trim()) }
			}
		}), can.page.Select(can, item._tr.parentNode, "", function(target) { can.page.ClassList.set(can, target, html.SELECT, target == item._tr) })
	},
	onclick: function(event, can, _sub, item) { switch (_sub.svg.style.cursor) {
		case "e-resize":
			can.Update(can.request(event, can.Action("direct") == "horizon"? {prev: item.hash}: {from: item.hash}), [ctx.ACTION, mdb.INSERT]); break
		case "s-resize":
			can.Update(can.request(event, can.Action("direct") == "horizon"? {from: item.hash}: {prev: item.hash}), [ctx.ACTION, mdb.INSERT]); break
		default: can.ondetail._select(event, can, item)
	} can.onkeymap.prevent(event) },
	oncontextmenu: function(event, can, _sub, item) {
		can.user.carteItem(event, can, can.base.CopyStr({action: item.action, zone: can.Option(mdb.ZONE)}, item))
	},
})
Volcanos(chat.ONEXPORT, {
	margin: function(can) { var margin = can.Action(html.MARGIN); return parseFloat(margin) },
	height: function(can) { var height = can.Action(html.HEIGHT); return parseFloat(height) },
	width: function(can) { var width = can.Action(html.WIDTH); return parseFloat(width) },
	travel: function(can, root, main) { if (!root) { return [] }
		if (can.Action("travel") == "deep") { var list = [root]
			if (main) {
				if (root.to) { list = list.concat(can.onexport.travel(can, root.to, false)) }
				if (root.next) { list = list.concat(can.onexport.travel(can, root.next, true)) }
			} else {
				if (root.next) { list = list.concat(can.onexport.travel(can, root.next, true)) }
				if (root.to) { list = list.concat(can.onexport.travel(can, root.to, false)) }
			}
		} else { var list = [root], i = 0
			while (i < list.length)  { var count = list.length
				for (i; i < count; i++) { for (var item = list[i].next; item; item = item.next) { list.push(item) } }
				if (count == 1) { i = 0 } var count = list.length
				for (i; i < count; i++) { for (var item = list[i].to; item; item = item.to) { list.push(item) } }
			}
		}
		return list
	},
})
Volcanos(chat.ONKEYMAP, {
	_mode: {plugin: {
		Escape: shy("清屏", function(event, can) { can.onaction.clear(event, can) }),
		g: shy("播放", function(event, can) { can.onaction.play(event, can) }),
		v: shy("预览", function(event, can) { can.onaction.show(event, can) }),
		r: shy("展示", function(event, can) { can.onaction.exec(event, can) }),
		" ": shy("展示", function(event, can) { can.onaction.exec(event, can) }),
		Enter: shy("预览", function(event, can) { can.onaction.show(event, can) }),
		k: shy("上一步", function(event, can) { can.db.current && can.db.current.from? can.ondetail._select(event, can, can.db.current.from): can.onaction.prev(event, can) }),
		h: shy("前一步", function(event, can) { can.db.current && can.db.current.prev? can.ondetail._select(event, can, can.db.current.prev): can.onaction.prev(event, can) }),
		l: shy("后一步", function(event, can) { can.db.current && can.db.current.next? can.ondetail._select(event, can, can.db.current.next): can.onaction.next(event, can) }),
		j: shy("下一步", function(event, can) { can.db.current && can.db.current.to? can.ondetail._select(event, can, can.db.current.to): can.onaction.next(event, can) }),
		ArrowUp: shy("上一步", function(event, can) { can.db.current && can.db.current.from? can.ondetail._select(event, can, can.db.current.from): can.onaction.prev(event, can) }),
		ArrowLeft: shy("前一步", function(event, can) { can.db.current && can.db.current.prev? can.ondetail._select(event, can, can.db.current.prev): can.onaction.prev(event, can) }),
		ArrowRight: shy("后一步", function(event, can) { can.db.current && can.db.current.next? can.ondetail._select(event, can, can.db.current.next): can.onaction.next(event, can) }),
		ArrowDown: shy("下一步", function(event, can) { can.db.current && can.db.current.to? can.ondetail._select(event, can, can.db.current.to): can.onaction.next(event, can) }),
	}}, _engine: {},
})
})()
