Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) { if (can.isCmdMode()) { can.onappend.style(can, html.OUTPUT), can.ConfHeight(can.page.height()) } can.onimport.layout(can)
		can.ui = {}, can.base.isFunc(cb) && cb(msg), can.onmotion.clear(can), can.onlayout.background(can, can.user.info.background, can._fields)
		can.onimport._menu(can), can.onimport._dock(can), can.onimport._searchs(can), can.onimport._notifications(can)
	},
	_menu: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.menu", style: html.OUTPUT}, function(sub) { can.ui.menu = sub
		sub.onexport.record = function(_, value, key, item) { delete(can.onfigure._path)
			switch (value) {
				case "create":
					can.onaction.create(event, can)
					break
				case "desktop": var carte = can.user.carte(event, can, {}, can.core.Item(can.onfigure), function(event, button, meta, carte) { can.onfigure[button](event, can, carte) }); break
				case "searchs": can.onmotion.toggle(can, can.ui.searchs._target); break
				case "notifications": can.onmotion.toggle(can, can.ui.notifications._target); break
			}
		}
		sub.onexport.output = function() { can.onimport._desktop(can, can._msg) }
	}) },
	_searchs: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.searchs"}, function(sub) { can.ui.searchs = sub
		can.page.style(can, sub._target, html.LEFT, can.ConfWidth()/4, html.TOP, can.ConfHeight()/4), sub.onimport.size(sub, can.ConfHeight()/2, can.ConfWidth()/2, true)
		sub.onexport.record = function(sub, value, key, item) {
			if (item.cmd == ctx.COMMAND) { can.onimport._window(can, {index: can.core.Keys(item.type, item.name.split(lex.SP)[0])}) }
		}
	}) },
	_notifications: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.notifications", style: html.OUTPUT}, function(sub) { can.ui.notifications = sub
		sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
	}) },
	_dock: function(can) { can.onappend.plugin(can, {index: "web.chat.macos.dock", style: html.OUTPUT}, function(sub) { can.ui.dock = sub
		sub.onexport.output = function(sub, msg) { can.page.style(can, sub._target, html.LEFT, (can.ConfWidth()-msg.Length()*80)/2) }
		sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
	}) },
	_item: function(can, item) { can.runAction(can.request(event, item), mdb.CREATE, [], function() { can.run(event, [], function(msg) {
		can.page.SelectChild(can, can.ui.desktop, html.DIV_ITEM, function(target) { can.page.Remove(can, target) }), can.onimport.__item(can, msg, can.ui.desktop)
	}) }) },
	__item: function(can, msg, target) { msg = msg||can._msg, msg.Table(function(item, index) {
		can.page.Append(can, target, [{view: html.ITEM, list: [{view: html.ICON, list: [{img: can.misc.PathJoin(item.icon)}]}, {view: [mdb.NAME, "", item.name]}],
			onclick: function(event) { can.onimport._window(can, item) }, style: can.onexport.position(can, index),
			oncontextmenu: function(event) { var carte = can.user.carteRight(event, can, {
				remove: function() { can.runAction(event, mdb.REMOVE, [item.hash]) },
			}); can.page.style(can, carte._target, html.TOP, event.y) },
		}])
	}) },
	_desktop: function(can, msg) { var target = can.page.Append(can, can._output, [{view: html.DESKTOP}])._target; can.onimport.__item(can, msg, target), can.ui.desktop = target
		target._tabs = can.onimport.tabs(can, [{name: "Desktop"+(can.page.Select(can, can._output, html.DIV_DESKTOP).length-1)}], function() { can.onmotion.select(can, can._output, "div.desktop", target), can.ui.desktop = target }, function() { can.page.Remove(can, target) }, can.ui.menu._output), target._tabs._desktop = target
		target.ondragend = function() { can.onimport._item(can, window._drag_item) }
	},
	_window: function(can, item) { item.height = can.base.Min(can.ConfHeight()-400, 320, 800), item.width = can.base.Min(can.ConfWidth()-400, 480, 1000)
		can.onappend.plugin(can, item, function(sub) { can.ondetail.select(can, sub._target)
			var index = 0; can.core.Item({
				"#f95f57": function(event) { sub.onaction.close(event, sub) },
				"#fcbc2f": function(event) {
					var dock = can.page.Append(can, can.ui.dock._output, [{view: html.ITEM, list: [{view: html.ICON, list: [{img: can.misc.PathJoin(item.icon)}]}], onclick: function() {
						can.onmotion.toggle(can, sub._target, true), can.page.Remove(can, dock)
					}}])._target; sub.onmotion.hidden(sub, sub._target)
				},
				"#32c840": function(event) { sub.onaction.full(event, sub) },
			}, function(color, cb) { can.page.insertBefore(can, [{view: [[html.ITEM, html.BUTTON]], style: {"background-color": color, right: 10+20*index++}, onclick: cb}], sub._output) })
			sub.onimport.size(sub, item.height, item.width, true), can.onmotion.move(can, sub._target, {"z-index": 10, top: 125, left: 100})
			sub.onmotion.resize(can, sub._target, function(height, width) { sub.onimport.size(sub, height, width) }, 25)
			sub.onexport.record = function(sub, value, key, item) { can.onimport._window(can, item) }
			sub.onexport.actionHeight = function(sub) { return can.page.ClassList.has(can, sub._target, html.OUTPUT)? 0: html.ACTION_HEIGHT+20 },
			sub.onexport.marginTop = function() { return 25 }
			sub.onappend.desktop = function(item) { can.onimport._item(can, item) }
			sub.onappend.dock = function(item) { can.ui.dock.runAction(can.request(event, item), mdb.CREATE, [], function() { can.ui.dock.Update() }) }
		}, can.ui.desktop)
	},
	session: function(can, list) { can.page.Select(can, can._output, html.DIV_DESKTOP, function(target) { can.page.Remove(can, target) }), can.onmotion.clear(can, can._action)
		can.core.List(list, function(item) { can.onimport._desktop(can), can.core.List(item.list, function(item) { can.onimport._window(can, item) }) })
	},
	layout: function(can) { can.page.styleHeight(can, can._output, can.ConfHeight()) },
}, [""])
Volcanos(chat.ONACTION, {list: ["full"],
	create: function(event, can, button) { can.onimport._desktop(can) },
	full: function(event, can) { document.body.requestFullscreen() },
})
Volcanos(chat.ONDETAIL, {
	select: function(can, target) { can.page.SelectChild(can, can.ui.desktop, html.FIELDSET, function(fieldset) {
		can.page.style(can, fieldset, "z-index", fieldset == target? "10": "9"), fieldset == target && can.onmotion.toggle(can, fieldset, true)
	}) },
})
Volcanos(chat.ONFIGURE, {
	"session\t>": function(event, can, carte) { can.runActionCommand(event, "session", [], function(msg) {
		var _carte = can.user.carteRight(event, can, {}, [{view: [html.ITEM, "", mdb.CREATE], onclick: function(event) {
			can.user.input(event, can, [mdb.NAME], function(list) {
				var args = can.page.SelectChild(can, can._output, html.DIV_DESKTOP, function(target) {
					return {name: can.page.Select(can, target._tabs, html.SPAN_NAME).innerText, list: can.page.SelectChild(can, target, html.FIELDSET, function(target) {
						return {index: target._can._index, args: target._can.onexport.args(target._can), style: {left: target.offsetLeft, top: target.offsetTop}}
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
			can.user.input(event, can, [ctx.INDEX, ctx.ARGS], function(data) {
				can.onimport._window(can, data)
			})
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
Volcanos(chat.ONEXPORT, {
	position: function(can, index) { var top = 25, margin = 20, height = 100, width = 80
		var n = parseInt((can.ConfHeight()-top)/(height+margin))
		return {top: index%n*height+top+margin/2, left: parseInt(index/n)*(width+margin)+margin/2}
	}
})
