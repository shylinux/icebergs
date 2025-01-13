(function() {
Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.isCmdMode() && can.onappend.style(can, html.OUTPUT)
		can.onlayout.background(can, can.misc.ResourceIcons(can, can.user.info.background||"/p/usr/icons/background.jpg"), can._fields)
		can.onimport._menu(can), can.onimport._notifications(can), can.onimport._searchs(can), can.onimport._dock(can)
		can.sup.onexport.link = function() { return can.misc.MergeURL(can, {pod: can.ConfSpace()||can.misc.Search(can, ice.POD), cmd: web.DESKTOP}) }
		can.onexport.title(can, can.ConfHelp(), can.user.titles)
	},
	_menu: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.menu", style: html.OUTPUT}, function(sub) { can.ui.menu = sub, sub._desktop = can
		var tabs = can.misc.sessionStorage(can, [can.ConfIndex(), html.TABS])
		sub.onexport.output = function() { can.onimport._desktop(can, can._msg)
			var sess = can.misc.SearchHash(can)[0]||can.Conf("session")
			sess? can.runActionCommand(event, "web.chat.macos.session", [sess], function(msg) {
				var item = msg.TableDetail(); can.onimport.session(can, can.base.Obj(item.args))
			}): !window.parent && can.isCmdMode() && can.onimport.session(can, tabs)
		}
		sub.onexport.record = function(sub, value, key, item) { delete(can.onfigure._path)
			switch (value) {
				case "notifications": can.ui.notifications._output.innerHTML && can.onmotion.toggle(can, can.ui.notifications._target); break
				case "searchs": can.onaction._search(can); break
				case "reload": can.Update(); break
				case cli.QRCODE: can.sup.onaction["生成链接"]({}, can.sup); break
				case mdb.CREATE: can.onaction.create(event, can); break
				case html.DESKTOP: var carte = can.user.carte(event, can, {}, can.core.Item(can.onfigure), function(event, button, meta, carte) { can.onfigure[button](event, can, carte); return true }); break
				default: can.onimport._window(can, value)
			}
		}
	}) },
	_notifications: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.notifications", style: html.OUTPUT}, function(sub) { can.ui.notifications = sub
		sub.onaction._close = function() { can.onmotion.hidden(can, sub._target) }, can.onmotion.hidden(can, sub._target)
		sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
	}) },
	_searchs: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.searchs"}, function(sub) {
		can.ui.searchs = sub, can.onmotion.hidden(can, sub._target)
		var width = can.base.Max(can.ConfWidth()/2, can.ConfWidth(), html.FLOAT_WIDTH)
		var height = can.base.Max(can.ConfHeight()/2, can.ConfHeight(), html.FLOAT_HEIGHT)
		can.page.style(can, sub._target, html.LEFT, (can.ConfWidth()-width)/2, html.TOP, (can.ConfHeight()-height)/2)
		sub.onimport.size(sub, height, width, true)
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
		can.onimport.layout(can)
		sub.onexport.output = function(sub, msg) {
			can.onimport.layout(can)
			can.onmotion.delay(can, function() {
				can.onimport.layout(can)
			})
		}
		sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
	}) },
	_desktop: function(can, msg, name) { var target = can.page.Append(can, can._output, [html.DESKTOP])._target; can.ui.desktop = target
		target._tabs = can.onimport.tabs(can, [{name: name||"Desktop"+(can.page.Select(can, can._output, html.DIV_DESKTOP).length-1)}], function() {
			can.onmotion.select(can, can._output, "div.desktop", target), can.ui.desktop = target, can.onexport.tabs(can)
		}, function() { can.page.Remove(can, target) }, can.ui.menu._output), target._tabs._desktop = target
		target.ondragend = function() { can.onimport._item(can, window._drag_item) }
		can.onimport.__item(can, msg, target)
		return target._tabs
	},
	_item: function(can, item) { can.runAction(can.request(event, item), mdb.CREATE, [], function() { can.run({}, [], function(msg) {
		can.page.SelectChild(can, can.ui.desktop, html.DIV_ITEM, function(target) { can.page.Remove(can, target) }), can.onimport.__item(can, msg, can.ui.desktop)
	}) }) },
	__item: function(can, msg, target) { var index = 0; can.onimport.icon(can, msg = msg||can._msg, target, function(target, item) { can.page.Modify(can, target, {
		onclick: function(event) { can.onimport._window(can, item) },
		oncontextmenu: function(event) { var carte = can.user.carteRight(event, can, {
			remove: function() { can.runAction(event, mdb.REMOVE, [item.hash]) },
		}); can.page.style(can, carte._target, html.TOP, event.y) },
	}) }) },
	_window: function(can, item, cb) { if (!item.index) { return }
		item.height = can.base.Max(html.DESKTOP_HEIGHT, can.ConfHeight()-125), item.width = can.base.Max(html.DESKTOP_WIDTH, can.ConfWidth())
		item.left = (can.ConfWidth()-item.width)/2, item.top = (can.ConfHeight()-item.height-125)/4+25
		item.type = html.PLUGIN, item.style = {left: item.left, top: item.top, height: item.height, width: item.width}
		can.onappend.plugin(can, item, function(sub) {
			var index = 0; can.core.Item({
				close: {color: "#f95f57", inner: "x", onclick: function(event) { sub.onaction._close(event, sub) }},
				small: {color: "#fcbc2f", inner: "-", onclick: function(event) { var dock = can.page.Append(can, can.ui.dock._output, [{view: html.ITEM, list: [{view: html.ICON, list: [{img: can.misc.PathJoin(item.icon)}]}], onclick: function() {
					can.onmotion.toggle(can, sub._target, true), can.page.Remove(can, dock)
				}}])._target; sub.onmotion.hidden(sub, sub._target) }},
				full: {color: "#32c840", inner: "+", onclick: function(event) { sub.onaction.full(event, sub) }},
			}, function(name, item) {
				can.page.insertBefore(can, [{view: [[html.ITEM, html.BUTTON, "window", name], ""], title: name, list: [{text: item.inner}], style: {"background-color": item.color, right: 10+25*index++}, onclick: item.onclick}], sub._output)
			})
			sub.onimport._open = function(sub, msg, arg) { can.onimport._window(can, {title: msg.Option(html.TITLE), index: web.CHAT_IFRAME, args: [arg]}) }
			sub.onimport._field = function(sub, msg) { msg.Table(function(item) { can.onimport._window(can, item) }) }
			sub.onaction._close = function() { can.page.Remove(can, sub._target), can.onexport.tabs(can) }
			sub.onappend.dock = function(item) { can.ui.dock.runAction(can.request(event, item), mdb.CREATE, [], function() { can.ui.dock.Update() }) }
			sub.onappend.desktop = function(item) { can.onimport._item(can, item) }
			sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
			sub.onexport.marginTop = function() { return 25 }, sub.onexport.marginBottom = function() { return 100 }
			sub.onexport.actionHeight = function(sub) { return can.page.ClassList.has(can, sub._target, html.OUTPUT)? 0: html.ACTION_HEIGHT+20 }
			sub.onexport.output = function() { sub.onimport.size(sub, item.height, can.base.Min(sub._target.offsetWidth, item.width), false)
				sub._target._meta.args = can.base.trim(can.page.SelectArgs(can, sub._option, "", function(target) { return target.value })), can.onexport.tabs(can)
			}
			can.onappend.style(can, html.FLOAT, sub._target), can.ondetail.select(can, sub._target, sub)
			sub.onimport.size(sub, item.height, can.base.Min(sub._target.offsetWidth, item.width), false)
			can.page.style(can, sub._target, html.HEIGHT, item.height, html.WIDTH, item.width)
			can.onmotion.move(can, sub._target, {top: item.top, left: item.left})
			sub.Conf("style.left", ""), sub.Conf("style.top", "")
			sub.onmotion.resize(can, sub._target, function(height, width) {
				sub.onimport.size(sub, item.height = height, item.width = width, false)
				can.page.style(sub, sub._target, html.HEIGHT, height, html.WIDTH, width)
				sub._target._meta.height = height, sub._target._meta.width = width, can.onexport.tabs(can)
			}, 25, 0, can.ui.desktop)
			sub._target.onclick = function(event) { can.ondetail.select(can, sub._target, sub) }
			sub._target._meta = {index: sub.ConfIndex(), args: sub.Conf(ctx.ARGS)}, can.onexport.tabs(can)
			cb && cb(sub)
		}, can.ui.desktop)
	},
	session: function(can, list) { if (!list || list.length == 0) { return }
		can.page.Select(can, can._output, html.DIV_DESKTOP, function(target) { can.page.Remove(can, target) })
		can.page.Select(can, can.ui.menu._output, html.DIV_TABS, function(target) { can.page.Remove(can, target) })
		var _select; can.core.Next(list, function(item, next) {
			var _tabs = can.onimport._desktop(can, null, item.name); _select = (!_select || item.select)? _tabs: _select
			can.core.Next(item.list, function(item, next) {
				can.onimport._window(can, item, function(sub) { can.onmotion.delay(can, function() { next() }, 300) })
			}, function() { next() })
		}, function() { _select && _select.click() })
	},
	layout: function(can) {
		can.page.style(can, can._output, html.HEIGHT, can.ConfHeight(), html.WIDTH, can.ConfWidth())
		can.ui.menu && can.ui.menu.onimport.size(can.ui.menu, html.DESKTOP_MENU_HEIHGT, can.ConfWidth(), false)
		can.ui.dock && can.page.style(can, can.ui.dock._target, html.LEFT, can.base.Min((can.ConfWidth()-can.ui.dock._target.offsetWidth)/2, 0))
		can.ui.dock && can.page.style(can, can.ui.dock._output, "position", "")
	},
}, [""])
Volcanos(chat.ONACTION, {
	_search: function(can) { var sub = can.ui.searchs; if (can.onmotion.toggle(can, sub._target)) {
		var height = can.ConfHeight()-115, top = 25, width = can.base.Min(sub._target.offsetWidth, can.ConfWidth()/2, can.ConfWidth()), left = (can.ConfWidth()-width)/2
		if (can.ConfHeight() > 600) { var height = can.ConfHeight()/2, top = can.ConfHeight()/4 }
		if (can.ConfWidth() < 800) { var width = can.ConfWidth(), left = 0 }
		can.page.style(can, sub._target, html.LEFT, left, html.TOP, top), sub.onimport.size(sub, height, width, true)
	} },
	create: function(event, can) { can.onimport._desktop(can) },
})
Volcanos(chat.ONDETAIL, {
	select: function(can, target, sub) {
		can.onmotion.select(can, can.ui.desktop, html.FIELDSET, target)
		can.onexport.title(can, sub.ConfHelp())
	},
})
Volcanos(chat.ONEXPORT, {
	tabs: function(can) {
		var list = can.page.Select(can, can.ui.menu._output, html.DIV_TABS, function(target) { return {
			select: can.page.ClassList.has(can, target, html.SELECT),
			name: can.page.SelectOne(can, target, html.SPAN).innerHTML,
			list: can.page.SelectChild(can, target._desktop, html.FIELDSET, function(target) { return target._meta })
		} }); can.misc.sessionStorage(can, [can.ConfIndex(), html.TABS], JSON.stringify(list))
	},
})
Volcanos(chat.ONKEYMAP, {
	escape: function(event, can) { can.onmotion.hidden(can, can.ui.searchs._target) },
	space: function(event, can) { can.onaction._search(can), can.onkeymap.prevent(event) },
	enter: function(event, can) { can.page.Select(can, can.ui.desktop, "fieldset.select", function(target) { target._can.Update(event) }) },
	ctrln: function(event, can) { can.onkeymap.selectCtrlN(event, can, can.ui.menu._output, html.DIV_TABS) },
	tabx: function(event, can) { can.page.Select(can, can.ui.menu._output, html.DIV_TABS_SELECT, function(target) { target._close() }) },
	tabs: function(event, can) { can.onaction.create(event, can) },
})
Volcanos(chat.ONFIGURE, {
	"session\t>": function(event, can, carte) { can.runActionCommand(event, "session", [], function(msg) {
		var hash = can.misc.SearchHash(can)
		var _carte = can.user.carteRight(event, can, {}, [{view: [html.ITEM, "", mdb.CREATE], onclick: function(event) {
			can.user.input(event, can, [mdb.NAME], function(list) {
				var args = can.page.SelectChild(can, can._output, html.DIV_DESKTOP, function(target) {
					return {name: can.page.SelectOne(can, target._tabs, html.SPAN_NAME).innerText, list: can.page.SelectChild(can, target, html.FIELDSET, function(target) {
						return {index: target._can._index, args: target._can.onexport.args(target._can), style: {left: target.offsetLeft||100, top: target.offsetTop||25}}
					})}
				})
				can.runActionCommand(event, "session", [ctx.ACTION, mdb.CREATE, mdb.NAME, list[0], ctx.ARGS, JSON.stringify(args)], function(msg) {
					can.user.toastSuccess(can, "session created"), can.misc.SearchHash(can, list[0])
				})
			})
		}}].concat("", msg.Table(function(value) {
			return {view: [html.ITEM, "", value.name+(value.name == hash[0]? " *": "")],
				onclick: function() { can.onimport.session(can, can.base.Obj(value.args, [])), can.misc.SearchHash(can, value.name) },
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
			return {view: [html.ITEM, "", legend.innerText+(can.page.ClassList.has(can, legend.parentNode, html.SELECT)? " *": "")], onclick: function(event) {
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
