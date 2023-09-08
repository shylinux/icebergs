(function() {
Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.onmotion.clear(can), can.isCmdMode() && can.onappend.style(can, html.OUTPUT)
		can.onlayout.background(can, can.user.info.background||"/require/usr/icons/background.jpg", can._fields)
		can.onimport._menu(can), can.onimport._notifications(can), can.onimport._searchs(can), can.onimport._dock(can)
	},
	_menu: function(can) { can.onappend.plugin(can, {_space: can.ConfSpace(), index: "web.chat.macos.menu", style: html.OUTPUT}, function(sub) { can.ui.menu = sub
		var cache = can.misc.sessionStorage(can, [can.ConfIndex(), "tabs"])
		sub.onexport.output = function() { can.onimport._desktop(can, can._msg)
			can.Conf("session") && can.runActionCommand(event, "session", [can.Conf("session")], function(msg) {
				var item = msg.TableDetail(); can.onimport.session(can, can.base.Obj(item.args))
			}), can.onimport.session(can, cache)
		}
		sub.onexport.record = function(sub, value, key, item) { delete(can.onfigure._path)
			switch (value) {
				case "create": can.onaction.create(event, can); break
				case "desktop": var carte = can.user.carte(event, can, {}, can.core.Item(can.onfigure), function(event, button, meta, carte) {
					can.onfigure[button](event, can, carte)
					return true
				}); break
				case "notifications": can.ui.notifications._output.innerHTML && can.onmotion.toggle(can, can.ui.notifications._target); break
				case "searchs": can.onaction._search(can); break
				default: can.onimport._window(can, value)
			}
		}
	}) },
	_notifications: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.notifications", style: html.OUTPUT}, function(sub) { can.ui.notifications = sub
		sub.onaction._close = function() { can.onmotion.hidden(can, sub._target) }, can.onmotion.hidden(can, sub._target)
		sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
	}) },
	_searchs: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.searchs"}, function(sub) { can.ui.searchs = sub
		can.page.style(can, sub._target, html.LEFT, can.ConfWidth()/4, html.TOP, can.ConfHeight()/4), sub.onimport.size(sub, can.ConfHeight()/2, can.ConfWidth()/2, true)
		sub.onaction._close = function() { can.onmotion.hidden(can, sub._target) }, can.onmotion.hidden(can, sub._target)
		sub.onexport.record = function(sub, value, key, item, event) { switch (item.type) {
			case ice.CMD: can.onimport._window(can, {index: item.name, args: can.base.Obj(item.text) }); break
			case nfs.FILE: can.onimport._window(can, {index: web.CODE_VIMER, args: can.misc.SplitPath(can, item.text) }); break
			case ssh.SHELL: can.onimport._window(can, {index: web.CODE_XTERM, args: [item.text]}); break
			case web.LINK:
			case web.WORKER:
			case web.SERVER:
			case web.GATEWAY: can.onimport._window(can, {index: web.CHAT_IFRAME, args: [item.text]}), can.onkeymap.prevent(event); break
			default: item.cmd == ctx.COMMAND && can.onimport._window(can, {index: can.core.Keys(item.type, item.name.split(lex.SP)[0])})
		} }
	}) },
	_dock: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.dock", style: html.OUTPUT}, function(sub) { can.ui.dock = sub
		sub.onexport.output = function(sub, msg) { can.onimport.layout(can) }
		sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
	}) },
	_desktop: function(can, msg, name) { var target = can.page.Append(can, can._output, [html.DESKTOP])._target; can.ui.desktop = target
		target._tabs = can.onimport.tabs(can, [{name: name||"Desktop"+(can.page.Select(can, can._output, html.DIV_DESKTOP).length-1)}], function() {
			can.onmotion.select(can, can._output, "div.desktop", target), can.ui.desktop = target
			can.onexport.tabs(can)
		}, function() { can.page.Remove(can, target) }, can.ui.menu._output), target._tabs._desktop = target
		target.ondragend = function() { can.onimport._item(can, window._drag_item) }
		can.onimport.__item(can, msg, target)
		return target._tabs
	},
	_item: function(can, item) { can.runAction(can.request(event, item), mdb.CREATE, [], function() { can.run(event, [], function(msg) {
		can.page.SelectChild(can, can.ui.desktop, html.DIV_ITEM, function(target) { can.page.Remove(can, target) }), can.onimport.__item(can, msg, can.ui.desktop)
	}) }) },
	__item: function(can, msg, target) { var index = 0; can.onimport.icon(can, msg = msg||can._msg, target, function(target, item) { can.page.Modify(can, target, {
		onclick: function(event) { can.onimport._window(can, item) }, style: can.onexport.position(can, index++),
		oncontextmenu: function(event) { var carte = can.user.carteRight(event, can, {
			remove: function() { can.runAction(event, mdb.REMOVE, [item.hash]) },
		}); can.page.style(can, carte._target, html.TOP, event.y) },
	}) }) },
	_window: function(can, item, cb) { if (!item.index) { return } item._space = can.ConfSpace()
		item.height = can.ConfHeight()-125, item.width = can.ConfWidth()-200, item.left = (can.ConfWidth()-item.width)/2, item.top = 25
		if (can.ConfWidth() > 1400) { item.width = can.base.Min(can.ConfWidth()-600, 640, 1400), item.left = (can.ConfWidth()-item.width)/2 }
		if (can.ConfHeight() > 800) { item.height = can.base.Min(can.ConfHeight()-200, 320, 800), item.top = 50 }
		if (can.user.isMobile) { item.left = 0, item.top = 25, item.height = can.ConfHeight()-125, item.width = can.ConfWidth() }
		can.onappend.plugin(can, item, function(sub) { can.ondetail.select(can, sub._target)
			var index = 0; can.core.Item({
				"#f95f57": function(event) { sub.onaction._close(event, sub) },
				"#fcbc2f": function(event) { var dock = can.page.Append(can, can.ui.dock._output, [{view: html.ITEM, list: [{view: html.ICON, list: [{img: can.misc.PathJoin(item.icon)}]}], onclick: function() {
					can.onmotion.toggle(can, sub._target, true), can.page.Remove(can, dock)
				}}])._target; sub.onmotion.hidden(sub, sub._target) },
				"#32c840": function(event) { sub.onaction.full(event, sub) },
			}, function(color, cb) { can.page.insertBefore(can, [{view: [[html.ITEM, html.BUTTON]], style: {"background-color": color, right: 10+20*index++}, onclick: cb}], sub._output) })
			sub.onexport.marginTop = function() { return 25 }, sub.onexport.marginBottom = function() { return 100 }
			sub.onexport.actionHeight = function(sub) { return can.page.ClassList.has(can, sub._target, html.OUTPUT)? 0: html.ACTION_HEIGHT+20 }
			sub.onexport.output = function() { item.index == "web.chat.macos.opens" && can.page.Remove(can, sub._target)
				sub.onimport.size(sub, item.height, can.base.Min(sub._target.offsetWidth, item.width), true)
				sub._target._meta.args = can.base.trim(can.page.SelectArgs(can, sub._option, "", function(target) { return target.value }))
				can.onexport.tabs(can)
			}, sub.onimport.size(sub, item.height, can.base.Min(sub._target.offsetWidth, item.width), true)
			sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
			sub.onimport._open = function(sub, msg, arg) { can.onimport._window(can, {index: web.CHAT_IFRAME, args: [arg]}) }
			sub.onappend.dock = function(item) { can.ui.dock.runAction(can.request(event, item), mdb.CREATE, [], function() { can.ui.dock.Update() }) }
			sub.onaction._close = function() { can.page.Remove(can, sub._target), can.onexport.tabs(can) }
			sub.onappend.desktop = function(item) { can.onimport._item(can, item) }
			can.onmotion.move(can, sub._target, {top: item.top, left: item.left})
			sub.onmotion.resize(can, sub._target, function(height, width) {
				sub.onimport.size(sub, item.height = height, item.width = width)
				sub._target._meta.height = height, sub._target._meta.width = width, can.onexport.tabs(can)
			}, 25)
			sub._target.onclick = function(event) { can.ondetail.select(can, sub._target) }
			sub._target._meta = {index: sub.ConfIndex(), args: sub.Conf(ctx.ARGS)}, can.onexport.tabs(can)
			cb && cb(sub)
		}, can.ui.desktop)
	},
	session: function(can, list) { if (!list || list.length == 0 || !can.isCmdMode() || window != window.parent) { return }
		can.page.Select(can, can._output, html.DIV_DESKTOP, function(target) { can.page.Remove(can, target) })
		can.page.Select(can, can.ui.menu._output, html.DIV_TABS, function(target) { can.page.Remove(can, target) })
		var _select; can.core.Next(list, function(item, next) {
			var _tabs = can.onimport._desktop(can, null, item.name); _select = (!_select || item.select)? _tabs: _select
			can.core.Next(item.list, function(item, next) {
				can.onimport._window(can, item, function(sub) { can.onmotion.delay(can, function() { next() }, 500) })
			}, function() { next() })
		}, function() { _select && _select.click() })
	},
	layout: function(can) { can.page.style(can, can._output, html.HEIGHT, can.ConfHeight(), html.WIDTH, can.ConfWidth())
		can.ui.notifications && can.page.style(can, can.ui.notifications._target, html.LEFT, can.ConfWidth()-250)
		can.ui.dock && can.page.style(can, can.ui.dock._target, html.LEFT, can.base.Min((can.ConfWidth()-(can.ui.dock._target.offsetWidth||502))/2, 0))
	},
}, [""])
Volcanos(chat.ONACTION, {list: ["full"],
	_search: function(can) { if (can.onmotion.toggle(can, can.ui.searchs._target)) {
		can.page.Select(can, can.ui.searchs._option, "input[name=keyword]", function(target) { can.onmotion.focus(can, target) })
	} },
	create: function(event, can) { can.onimport._desktop(can) },
	full: function(event, can) { document.body.requestFullscreen() },
})
Volcanos(chat.ONKEYMAP, {	
	escape: function(event, can) { can.onmotion.hidden(can, can.ui.searchs._target) },
	space: function(event, can) { can.onaction._search(can), can.onkeymap.prevent(event) },
	enter: function(event, can) { can.page.Select(can, can.ui.desktop, "fieldset.select", function(target) { target._can.Update(event) }) },
	ctrln: function(event, can) { can.onkeymap.selectCtrlN(event, can, can.ui.menu._output, html.DIV_TABS) },
	tabx: function(event, can) { can.page.Select(can, can.ui.menu._output, html.DIV_TABS_SELECT, function(target) { target._close() }) },
	tabs: function(event, can) { can.onaction.create(event, can) },
})
Volcanos(chat.ONDETAIL, {
	select: function(can, target) { can.onmotion.select(can, can.ui.desktop, html.FIELDSET, target) },
})
Volcanos(chat.ONEXPORT, {
	position: function(can, index) { var top = 25, margin = 20, height = 100, width = 80
		var n = parseInt((can.ConfHeight()-top)/(height+margin))
		return {top: index%n*height+top+margin/2, left: parseInt(index/n)*(width+margin)+margin/2}
	},
	tabs: function(can) {
		var list = can.page.Select(can, can.ui.menu._output, html.DIV_TABS, function(target) { return {
			select: can.page.ClassList.has(can, target, html.SELECT),
			name: can.page.SelectOne(can, target, html.SPAN_NAME).innerHTML,
			list: can.page.SelectChild(can, target._desktop, html.FIELDSET, function(target) { return target._meta })
		} }); can.misc.sessionStorage(can, [can.ConfIndex(), html.TABS], JSON.stringify(list))
	},
})
Volcanos(chat.ONFIGURE, {
	"session\t>": function(event, can, carte) { can.runActionCommand(event, "session", [], function(msg) {
		var _carte = can.user.carteRight(event, can, {}, [{view: [html.ITEM, "", mdb.CREATE], onclick: function(event) {
			can.user.input(event, can, [mdb.NAME], function(list) {
				var args = can.page.SelectChild(can, can._output, html.DIV_DESKTOP, function(target) {
					return {name: can.page.SelectOne(can, target._tabs, html.SPAN_NAME).innerText, list: can.page.SelectChild(can, target, html.FIELDSET, function(target) {
						return {index: target._can._index, args: target._can.onexport.args(target._can), style: {left: target.offsetLeft||100, top: target.offsetTop||25}}
					})}
				})
				can.runActionCommand(event, "session", [ctx.ACTION, mdb.CREATE, mdb.NAME, list[0], ctx.ARGS, JSON.stringify(args)], function(msg) {
					can.user.toastSuccess(can, "session created")
				})
			})
		}}].concat("", msg.Table(function(value) {
			return {view: [html.ITEM, "", value.name],
				onclick: function() { can.onimport.session(can, can.base.Obj(value.args, [])) },
				oncontextmenu: function(event) { can.user.carteRight(event, can, {
					open: function() { can.user.open(can.misc.MergePodCmd(can, {cmd: "desktop", session: value.name})) },
					remove: function() { can.runActionCommand(event, "session", [mdb.REMOVE, value.name], function() { can.user.toastSuccess(can, "session removed") }) },
				}, [], function() {}, _carte) },
			}
		})), function(event) {}, carte)
	}) },
	"desktop\t>": function(event, can, carte) {
		var _carte = can.user.carteRight(event, can, {}, [{view: [html.ITEM, "", mdb.CREATE], onclick: function(event) {
			can.onaction.create(event, can), can.user.toastSuccess(can, "desktop created")
		}}].concat("", can.page.Select(can, can.ui.menu._output, "div.tabs>span.name", function(target) {
			return {view: [html.ITEM, "", target.innerText+(can.page.ClassList.has(can, target.parentNode, html.SELECT)? " *": "")],
				onclick: function(event) { target.click() },
				oncontextmenu: function(event) { can.user.carteRight(event, can, {
					remove: function() { target.parentNode._close(), can.user.toastSuccess(can, "desktop removed") },
				}, [], function() {}, _carte) },
			}
		})), function(event) {}, carte)
	},
	"window\t>": function(event, can, carte) {
		can.user.carteRight(event, can, {}, [{view: [html.ITEM, "", mdb.CREATE], onclick: function(event) {
			can.user.input(event, can, [ctx.INDEX, ctx.ARGS], function(data) { can.onimport._window(can, data) })
		}}, ""].concat(can.page.Select(can, can.ui.desktop, "fieldset>legend", function(legend) {
			return {view: [html.ITEM, "", legend.innerText+(legend.parentNode.style["z-index"] == "10"? " *": "")], onclick: function(event) {
				can.ondetail.select(can, legend.parentNode)
			}}
		})), function(event) {}, carte)
	},
	"layout\t>": function(event, can, carte) { var list = can.page.SelectChild(can, can.ui.desktop, html.FIELDSET)
		can.user.carteRight(event, can, {
			grid: function(event) { for (var i = 0; i*i < list.length; i++) {} for (var j = 0; j*i < list.length; j++) {}
				var height = (can.ConfHeight()-25)/j, width = can.ConfWidth()/i; can.core.List(list, function(target, index) {
					can.page.style(can, target, html.TOP, parseInt(index/i)*height+25, html.LEFT, index%i*width)
					target._can.onimport.size(target._can, height, width)
				})
			},
			free: function(event) { can.core.List(list, function(target, index) {
				can.page.style(can, target, html.TOP, can.ConfHeight()/2/list.length*index+25, html.LEFT, can.ConfWidth()/2/list.length*index)
			}) },
			full: function(event) { can.onaction.full(event, can) },
		}, [], function(event) {}, carte)
	},
})
})()
