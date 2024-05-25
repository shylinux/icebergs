Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { msg.Push(mdb.ZONE, ctx.COMMAND)
		can.ui = can.onappend.layout(can), can.onimport._project(can, msg), can.ui.toggle = can.onappend.toggle(can)
	},
	_project: function(can, msg) { msg.Table(function(value) {
		can.onimport.item(can, value, function(event, item, show, target) {
			can.onimport.tabsCache(can, item, target, function() {
				if (value.zone == ctx.COMMAND) {
					can.onimport._command(can, value)
				} else {
					can.run(event, [value.zone], function(msg) { can.onimport._content(can, msg, value) })
				}
			}), value._current && can.onimport._profile(can, value._current, value)
		})
	}) },
	_command: function(can, value) {
		can.run(event, [ctx.RUN, ctx.COMMAND], function(msg) { var res = can.request(), cmds = {ice: []}
			res.Push(mdb.HASH, "ice").Push(mdb.NAME, "ice").Push(ctx.INDEX, "").Push("prev", "").Push("from", "")
			msg.Table(function(value) { can.core.List(value.index.split("."), function(cmd, index, list) { if (list[0] == "ice") { return }
				var _mod = list.slice(0, index).join(".")||"ice", _cmd = list.slice(0, index+1).join(".")
				var last = (cmds[_mod][cmds[_mod].length-1])||_mod; _cmd != last && cmds[_mod].push(_cmd)
				var prev = "", from = ""; if (index % 2 == 0) { prev = last } else { from = last }
				if (!cmds[_cmd]) { if (index < list.length-1) { cmds[_cmd] = [] }
					res.Push(mdb.HASH, _cmd).Push(mdb.NAME, cmd).Push(ctx.INDEX, index < list.length-1? "": _cmd).Push("prev", prev).Push("from", from)
				}
			}) }), can.onimport._content(can, res, value)
		})
	},
	_content: function(can, msg, value) {
		can.onappend.plugin(can, {display: "/plugin/local/wiki/draw.js"}, function(sub) {
			sub.onexport.output = function() { value._content_plugin = sub, can.onimport._toolkit(can, msg, value) }
		}), can.onappend._status(can, msg)
	},
	_toolkit: function(can, msg, value) {
		can.onappend.plugin(can, {index: "can._action"}, function(sub) { sub.ConfSpace(can.ConfSpace()), sub.ConfIndex([can.ConfIndex(), value.zone].join(":"))
			sub.run = function(event, cmds, cb) { cmds[0] == ctx.ACTION? can.core.CallFunc(can.onaction[cmds[1]], [event, can, value]): cb && cb(can.request(event)) }
			sub.onexport.output = function() { value._toolkit_plugin = sub, sub.onappend._action(sub, can.onaction._toolkit), sub.onaction._select(), can.onimport.layout(can) }
			sub.onaction._select = function() { can.onimport._display(can, msg, value), can.onimport._flows(can, value) }
		})
	},
	_display: function(can, msg, value) {
		var list = {}; msg.Table(function(value) { list[value.hash] = value })
		can.core.Item(list, function(key, item) { if (!item.prev && !item.from) { return value._root = item }
			if (item.prev) { list[item.prev].next = item, item.prev = list[item.prev] }
			if (item.from) { list[item.from].to = item, item.from = list[item.from] }
		}), value._list = list
		var _list = can.onexport.travel(can, value, value._root), _msg = can.request(); can.core.List(_list, function(item) {
			_msg.Push(mdb.TIME, item.time), _msg.Push(mdb.HASH, item.hash), _msg.Push(mdb.NAME, item.name)
			_msg.Push(web.SPACE, item.space), _msg.Push(ctx.INDEX, item.index||""), _msg.Push(ctx.ARGS, item.args||""), _msg.Push(ctx.ACTION, item.action||"")
		}); var table = can.onappend.table(can, _msg, null, can.ui.display); can.page.Select(can, table, "tbody>tr", function(target, index) { _list[index]._tr = target })
	},
	_profile: function(can, item, value) { value._profile_plugin = item._profile_plugin, value._current = item
		can.onmotion.toggle(can, can.ui.profile, true), can.onimport.layout(can)
		can.onexport.hash(can, value.zone, item.hash), can.onexport.title(can, value.zone, item.name||item.index.split(".").pop())
		if (can.onmotion.cache(can, function() { return can.core.Keys(value.zone, item.hash) }, can.ui.profile)) { return }
		item.index && can.onappend.plugin(can, {pod: item.space, index: item.index, args: item.args}, function(sub) { value._profile_plugin = item._profile_plugin = sub
			sub.onaction._close = function() { can.onmotion.hidden(can, can.ui.profile), can.onimport.layout(can)  }
			sub.onexport.output = function() { can.onimport.layout(can) }
		}, can.ui.profile)
	},
	_flows: function(can, value) { var sub = value._toolkit_plugin.sub, _sub = value._content_plugin.sub
		var margin = can.onexport.margin(sub), height = can.onexport.height(sub), width = can.onexport.width(sub)
		var matrix = {}, lines = [], rects = [], horizon = sub.Action("direct") == "horizon"
		can.onmotion.clear(can, _sub.ui.svg), _sub.ui.svg.Value("font-size", sub.Action("font-size")+"px")
		function show(item, main, deep) { var prev = "from", from = "prev"; if (horizon) { var prev = "prev", from = "from" }
			while (matrix[can.core.Keys(item.x, item.y)]) {
				if (horizon && main || !horizon && !main) {
					for(var _head = item[prev]; _head; _head = _head[prev]) { if (!_head[prev]) { var list = can.onexport.travel(can, value, _head, main)
						can.core.List(list, function(_item) { if (!_item._rect) { return _item.y++ } delete(matrix[can.core.Keys(_item.x, _item.y)]), _item.y++
							if ( _item._line) { _item != _head && _item._line.Val("y1", _item._line.Val("y1")+height), _item._line.Val("y2", _item._line.Val("y2")+height) }
							_item._rect.Val("y", _item._rect.Val("y")+height), _item._text.Val("y", _item._text.Val("y")+height)
						}), can.core.List(list, function(_item) { if (!_item._rect) { return } matrix[can.core.Keys(_item.x, _item.y)] = _item })
					} }
				} else {
					for(var _head = item[from]; _head; _head = _head[from]) { if (!_head[from]) { var list = can.onexport.travel(can, value, _head, main)
						can.core.List(list, function(_item) { if (!_item._rect) { return _item.x++ } delete(matrix[can.core.Keys(_item.x, _item.y)]), _item.x++
							if ( _item._line) { _item != _head && _item._line.Val("x1", _item._line.Val("x1")+width), _item._line.Val("x2", _item._line.Val("x2")+width) }
							_item._rect.Val("x", _item._rect.Val("x")+width), _item._text.Val("x", _item._text.Val("x")+width)
						}), can.core.List(list, function(_item) { if (!_item._rect) { return } matrix[can.core.Keys(_item.x, _item.y)] = _item })
					} }
				}
			} matrix[can.core.Keys(item.x, item.y)] = item
			if (item.prev || item.from) { lines.length <= deep && lines.push(_sub.onimport.group(_sub, "line"+deep))
				item._line = _sub.onimport.draw(_sub, {shape: svg.LINE, points:
					horizon && item.from || !horizon && !item.from? [{x: item.x*width+width/2, y: item.y*height-margin}, {x: item.x*width+width/2, y: item.y*height+margin}]:
					[{x: item.x*width-margin, y: item.y*height+height/2}, {x: item.x*width+margin, y: item.y*height+height/2}]
				}, lines[deep])
			} rects.length <= deep && rects.push(_sub.onimport.group(_sub, "rect"+deep)), can.onimport._block(can, value, item, item.x*width, item.y*height, rects[deep])
		} value._root.x = 0, value._root.y = 0; var list = can.onexport.travel(can, value, value._root, true, 0)
		can.core.Next(list, function(item, next, index) { can.user.toastProcess(can, index+" / "+list.length, "", index/list.length*100)
			show(item, item._main, item._deep), can.onmotion.delay(can, function() { next() }, 30)
			can.isCmdMode() && can.user.isChrome && item._rect.scrollIntoViewIfNeeded()
		}, function() { can.user.toastSuccess(can) })
		var max_x = 0, max_y = 0; can.core.List(list, function(item) { item.x > max_x && (max_x = item.x), item.y > max_y && (max_y = item.y) })
		_sub.ui.svg.Value(html.HEIGHT, max_y*height), _sub.ui.svg.Value(html.WIDTH, max_x*width)
	},
	_block: function(can, value, item, x, y, group) { var sub = value._toolkit_plugin.sub, _sub = value._content_plugin.sub
		var margin = can.onexport.margin(sub), height = can.onexport.height(sub), width = can.onexport.width(sub)
		var rect = _sub.onimport.draw(_sub, {shape: svg.RECT, points: [{x: x+margin, y: y+margin}, {x: x+width-margin, y: y+height-margin}]}, group)
		var text = _sub.onimport.draw(_sub, {shape: svg.TEXT, points: [{x: x+width/2, y: y+height/2}], style: {inner: item.name||item.index.split(nfs.PT).pop()}}, group)
		var line = item._line||{}; item._rect = rect, item._text = text
		can.core.ItemCB(can.ondetail, function(key, cb) { line[key] = rect[key] = text[key] = function(event) { can.request(event, item, value), cb(event, can, item, value) } })
		if (item.status) { item._line && line.Value(html.CLASS, item.status), rect.Value(html.CLASS, item.status), text.Value(html.CLASS, item.status) }
		if (value.zone == can.db.hash[0] && item.hash == can.db.hash[1] && can.onexport.session(can, "profile.show") != ice.FALSE) { can.onmotion.delay(can, function() { rect.onclick({target: rect}) }) }
	},
	layout: function(can) { can.ui.layout(can.ConfHeight(), can.ConfWidth(), 0, function(height, width) {
		var sub = can.db.value && can.db.value._toolkit_plugin; if (sub) {
			can.page.style(can, sub._target, html.LEFT, 0), sub.onimport.size(sub, html.ACTION_HEIGHT, width, true)
			can.page.style(can, sub._target, html.LEFT, (can.ui.content.offsetWidth-sub._target.offsetWidth)/2)
		}
	}) },
})
Volcanos(chat.ONACTION, {
	_trans: {
		style: {
			addnext: "notice",
			addto: "notice",
		},
		icons: {
			addnext: "bi bi-arrow-down-square",
			addto: "bi bi-arrow-right-square",
		},
		addnext: "添加下一步",
		addto: "添加下一项",
	},
	_toolkit: [
		"play", "prev", "next",
		["travel", "deep", "wide"],
		["delay", 1000, 3000, 5000],
		"",
		["direct", "vertical", "horizon"],
		["font-size", 18, 14, 16, 18, 20, 22],
		[html.MARGIN, 10, 5, 10, 20, 40, 60],
		[html.HEIGHT, 60, 40, 60, 80, 100, 120, 140, 200],
		[html.WIDTH, 180, 200, 240, 280, 400],
	],
	play: function(event, can, value) { var list = can.onexport.travel(can, value, value._root); var sub = value._toolkit_plugin.sub
		can.core.List(list, function(item) { item._line && item._line.Value(html.CLASS, ""), item._rect.Value(html.CLASS, ""), item._text.Value(html.CLASS, "") })
		can.core.Next(list, function(item, next, index, list) {
			item._line && item._line.Value(html.CLASS, "done"), item._rect.Value(html.CLASS, "done"), item._text.Value(html.CLASS, "done")
			can.user.toast(can, list[index].index), can.ondetail._select(event, can, item, value), can.onmotion.delay(can, next, sub.Action("delay"))
		}, function() { can.user.toastSuccess(can) })
	},
	prev: function(event, can, value) { if (!can.db.current) { can.db.current = value._root }
		can.ondetail._select(event, can, can.db.current.prev || can.db.current.from, value)
	},
	next: function(event, can, value) { if (!can.db.current) { can.db.current = value._root } var sub = value._toolkit_plugin.sub
		can.ondetail._select(event, can, sub.Action("travel") == "wide" && can.db.current.next || can.db.current.to, value)
	},

	create: function(event, can) { can.user.input(event, can, can.Conf("feature.create"), function(data) {
		can.runAction(can.request(event, data), mdb.CREATE, [], function(msg) { can.db.hash = can.onexport.hash(can, data.zone)
			msg = can.request(), msg.Push(data), can.onimport._project(can, msg)
		})
	}) },
	addnext: function(event, can) { can.onaction._insert(event, can, "prev") },
	addto: function(event, can) { can.onaction._insert(event, can, "from") },
	toggles: function(event, can) { var msg = can.request(event)
		can.db.value._list[msg.Option(mdb.HASH)]._close = !can.db.value._list[msg.Option(mdb.HASH)]._close
		can.onimport._flows(can, can.db.value)
	},
	rename: function(event, can) { can.onaction._modify(event, can, [mdb.NAME]) },
	plugin: function(event, can) { can.onaction._modify(event, can, [ctx.INDEX, ctx.ARGS]) },
	_insert: function(event, can, from) { var msg = can.request(event), zone = msg.Option(mdb.ZONE)
		can.user.input(event, can, can.Conf("feature.insert"), function(data) {
			can.runAction(can.request({}, data, kit.Dict(mdb.ZONE, zone, from, msg.Option(mdb.HASH))), mdb.INSERT, [], function(msg) {
				can.db.hash = can.onexport.hash(can, zone, msg.Result())
				can.run(event, [zone], function(msg) { can.onimport._content(can, msg, can.db.value) })
			})
		})
	},
	_modify: function(event, can, list) { var msg = can.request(event), zone = msg.Option(mdb.ZONE)
		can.user.input(event, can, list, function(args) {
			can.runAction(can.request({}, {zone: zone, hash: msg.Option(mdb.HASH)}), mdb.MODIFY, args, function() {
				can.run(event, [zone], function(msg) { can.onimport._content(can, msg, can.db.value) })
			})
		})
	},
})
Volcanos(chat.ONDETAIL, {
	_select: function(event, can, item, value) { can.onimport._profile(can, item, value)
		var sub = value._toolkit_plugin.sub, _sub = value._content_plugin.sub
		can.page.Select(can, _sub.ui.svg, "rect", function(target) { var _class = (target.Value(html.CLASS)||"").split(lex.SP)
			if (can.base.isIn(target, item._line, item._rect, item._text)) {
				if (_class.indexOf(html.SELECT) == -1) { target.Value(html.CLASS, _class.concat([html.SELECT]).join(lex.SP).trim()) }
			} else {
				if (_class.indexOf(html.SELECT) > -1) { target.Value(html.CLASS, _class.filter(function(c) { return c != html.SELECT }).join(lex.SP).trim()) }
			}
		})
		can.page.Select(can, item._tr.parentNode, "", function(target) { can.page.ClassList.set(can, target, html.SELECT, target == item._tr)
			can.onmotion.scrollIntoView(can, item._tr, 45, can.ui.display)
		}), can.isCmdMode() && can.user.isChrome && item._rect.scrollIntoViewIfNeeded()
	},
	onclick: function(event, can, item, value) { can.request(event, item, {zone: value.zone})
		var sub = value._toolkit_plugin.sub, _sub = value._content_plugin.sub
		switch (_sub.ui.svg.style.cursor) {
			case "e-resize": if (sub.Action("direct") == "horizon") { can.onaction.addnext(event, can) } else { can.onaction.addto(event, can) } break
			case "s-resize": if (sub.Action("direct") == "horizon") { can.onaction.addto(event, can) } else { can.onaction.addnext(event, can) } break
			default: can.ondetail._select(event, can, item, value)
		}
	},
	oncontextmenu: function(event, can, item, value) {
		item.action? can.user.carteItem(event, can, can.base.CopyStr({action: item.action, zone: value.zone}, item)): can.user.carte(can.request(event, item, value), can, can.onaction, ["toggles"], function(event, button) {
			can.request(event, item, value), can.onaction[button](event, can, button)
		})
	},
})
Volcanos(chat.ONEXPORT, {
	margin: function(can) { var margin = can.Action(html.MARGIN); return parseFloat(margin)||10 },
	height: function(can) { var height = can.Action(html.HEIGHT); return parseFloat(height)||60 },
	width: function(can) { var width = can.Action(html.WIDTH); return parseFloat(width)||200 },
	travel: function(can, value, root, main, deep) { if (!root) { return [] } root._deep = deep||0
		var list = [root]; if (root._close) { return list } var sub = value._toolkit_plugin.sub
		if (sub.Action("travel") == "deep") { main == undefined && (main = true), root._main = main
			var horizon = sub.Action("direct") == "horizon"
			var next = 0, to = 1; if (horizon) { var next = 1, to = 0 }
			if (main) {
				var _item = root.to; if (_item) { _item.x = root.x+to, _item.y = root.y+next, list = list.concat(can.onexport.travel(can, value, _item, false, deep+1)) }
				var _item = root.next; if (_item) { _item.x = root.x+next, _item.y = root.y+to, list = list.concat(can.onexport.travel(can, value, _item, true, deep)) }
			} else {
				var _item = root.next; if (_item) { _item.x = root.x+next, _item.y = root.y+to, list = list.concat(can.onexport.travel(can, value, _item, true, deep+1)) }
				var _item = root.to; if (_item) { _item.x = root.x+to, _item.y = root.y+next, list = list.concat(can.onexport.travel(can, value, _item, false, deep)) }
			}
		} else { var i = 0
			while (i < list.length)  { var count = list.length
				for (i; i < count; i++) { for (var item = list[i].next; item; item = item.next) { list.push(item) } }
				if (count == 1) { i = 0 } var count = list.length
				for (i; i < count; i++) { for (var item = list[i].to; item; item = item.to) { list.push(item) } }
			}
		}
		return list
	},
})
