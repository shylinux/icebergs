Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		// if (can.isCmdMode()) { can.onappend.style(can, html.OUTPUT) }
		can.ui = can.onappend.layout(can), can.onimport._project(can, msg)
	},
	_project: function(can, msg) { var select, current = can.db.hash[0]||can.sup.db.current||ice.DEV
		can.page.insertBefore(can, [{view: wiki.TITLE, list: [
			{icon: "bi bi-three-dots", onclick: function() { can._legend.onclick(event) }},
			{text: "message"||can.ConfIndex(), onclick: function(event) { can._legend.onclick(event) }},
			{icon: "bi bi-plus-lg", onclick: function(event) { can.Update(event, [ctx.ACTION, mdb.CREATE]) }},
		]}], can.ui.project.firstChild, can.ui.project)
		msg.Table(function(value) {
			var _target = can.page.Append(can, can.ui.project, [{view: html.ITEM, list: [
				{img: can.misc.Resource(can, value.icons||"usr/icons/Messages.png")}, {view: html.CONTAINER, list: [
					{view: wiki.TITLE, list: [{text: value.title||can.base.trimPrefix(value.zone, "ops.")||"[未命名]"}, {text: [can.base.TimeTrim(value.time), "", mdb.TIME]}]},
					{view: wiki.CONTENT, list: [{text: value.target||"[未知消息]"}]},
				]},
			], onclick: function(event) { can.isCmdMode() && can.misc.SearchHash(can, value.zone), can.onimport._switch(can, false)
				can.sup.db.current = value.zone
				can.db.zone = value, can.db.hash = value.hash, can.onmotion.select(can, can.ui.project, html.DIV_ITEM, _target)
				if (can.onmotion.cache(can, function(save, load) {
					can.ui.message && save({title: can.ui.title, message: can.ui.message, scroll: can.ui.message.scrollTop})
					return load(value.zone, function(bak) { can.ui.title = bak.title, can.ui.message = bak.message
						can.onmotion.delay(can, function() { can.ui.message.scrollTop = bak.scroll })
					})
				}, can.ui.content, can.ui.profile, can.ui.display, can._status)) { return can.onimport.layout(can) }
				can.run(can.request(event, {"cache.limit": 10}), [value.hash], function(msg) {
					can.onimport._display(can), can.onimport._content(can, msg)
				})
			}, oncontextmenu: function(event) {
				can.user.carteRight(event, can, {}, can.page.parseAction(can, value), function(event, button) {
					can.runAction(can.request(event, value), button)
				})
			}}])._target; select = (value.zone == current? _target: select)||_target
		}), can.user.isMobile? can.onimport._switch(can, true): select && select.click()
		can.onmotion.orderShow(can, can.ui.project)
	},
	_content: function(can, msg) {
		can.ui.title = can.page.Appends(can, can.ui.content, [{view: wiki.TITLE, list: [
			{icon: "bi bi-chevron-left", onclick: function() { can.onimport._switch(can, true) }},
			{text: can.db.zone.title||can.base.trimPrefix(can.db.zone.zone, "ops.")},
			{icon: "bi bi-three-dots", onclick: function() { can.onmotion.toggle(can, can.ui.profile), can.onimport.layout(can) }},
		]}])._target
		can.ui.message = can.page.Append(can, can.ui.content, [{view: html.LIST}])._target, can.onimport._message(can, msg)
	},
	_display: function(can, msg) {
		can.page.Appends(can, can.ui.display, [
			{view: "toolkit", list: can.core.Item(can.ondetail, function(icon, cb) { return {icon: icon, onclick: function(event) { cb(event, can) }} }) },
			{type: html.TEXTAREA, onkeyup: function(event) { if (event.key == "Enter" && event.ctrlKey) {
				can.onimport._insert(can, [mdb.TYPE, "text", mdb.TEXT, event.target.value])
			} }},
		]), can.onmotion.toggle(can, can.ui.display, true)
	},
	_message: function(can, msg) { var now = new Date(), last = ""
		msg.Table(function(value) { can.db.zone.id = value.id
			// value.space = value.space||can.base.trimPrefix(can.db.zone.target, "ops.")
			var myself = value.username == can.user.info.username, time = can.base.TimeTrim(value.time)
			var t = new Date(value.time); if (!last || (t - last > 3*60*1000)) { last = t
				can.page.Append(can, can.ui.message, [{view: [[html.ITEM, mdb.TIME], "", time]}])
			}
			can.page.Append(can, can.ui.message, [{view: [[html.ITEM, value.direct, value.type]], list: [
				{img: can.misc.Resource(can, value.direct == "recv"? (
					(can.base.isIn(value.avatar, can.db.zone.zone, mdb.TYPE)? "": value.avatar)||can.db.zone.icons||"usr/icons/Messages.png"
				): (can.user.info.avatar)||"usr/icons/Messages.png")},
				{view: html.CONTAINER, list: [{text: [
					value.direct == "recv"? value.usernick||can.db.zone.title||can.db.zone.zone: value.usernick||value.username
					, "", nfs.FROM]}, can.onfigure[value.type||"text"](can, value)]},
			]}])
		}), can.onappend._status(can, msg.Option(ice.MSG_STATUS)), can.onimport.layout(can)
		if (can.Status(mdb.TOTAL) > can.db.zone.id) { can.onimport._request(can) }
		can.onmotion.delay(can, function() { can.ui.message && (can.ui.message.scrollTop += 10000) }, 300)
	},
	_request: function(can) {
		can.Update(can.request({}, {"cache.begin": parseInt(can.db.zone.id||0)+1, "cache.limit": 10}), [can.db.hash], function(msg) {
			can.onimport._message(can, msg)
		})
	},
	_insert: function(can, args) {
		can.runAction(event, tcp.SEND, [can.db.hash].concat(args), function() {
			can.onimport._request(can)
		})
	},
	_switch: function(can, project) { if (!can.user.isMobile) { return }
		can.page.style(can, can.ui.project, html.WIDTH, can.ConfWidth())
		can.page.style(can, can.ui.project, html.FLEX, "0 0 "+can.ConfWidth()+"px")
		can.onmotion.toggle(can, can.ui.project, project)
		can.onmotion.toggle(can, can.ui.content, !project)
		can.onmotion.toggle(can, can.ui.display, !project)
		can.onimport.layout(can)
	},
	layout: function(can) { can.ui.layout(can.ConfHeight(), can.ConfWidth())
		can.ui.title && can.page.style(can, can.ui.message, html.HEIGHT, can.ui.content.offsetHeight-can.ui.title.offsetHeight)
		can.page.style(can, can._output, html.HEIGHT, can.ConfHeight())
	},
}, [""])
Volcanos(chat.ONDAEMON, {
	refresh: function(can, msg, sub, arg) { sub.sub.onimport._request(sub.sub) },
})
Volcanos(chat.ONEXPORT, {
	plugHeight: function(can, value) { var height = can.base.Min(can.ui.content.offsetHeight-240, 240)
		return can.base.Max(html.STORY_HEIGHT, height, height/(can.base.isIn(value.index, html.IFRAME)? 1: 2))
	},
	plugWidth: function(can, value) {
		return can.ui.content.offsetWidth-(can.user.isMobile? 80: 180)
	},
})
Volcanos(chat.ONDETAIL, {
	"bi bi-mic": function(event, can) {},
	"bi bi-card-image": function(event, can) {
		can.user.input(event, can, [mdb.ICONS], function(args) {
			can.onimport._insert(can, [mdb.TYPE, html.IMAGE].concat([mdb.TEXT, args[1]]))
		})
	},
	"bi bi-camera": function(event, can) {},
	"bi bi-camera-video": function(event, can) {},
	"bi bi-file-earmark": function(event, can) {},
	"bi bi-geo-alt": function(event, can) {},
	"bi bi-window": function(event, can) {
		can.user.input(event, can, [web.SPACE, ctx.INDEX, ctx.ARGS], function(args) {
			can.onimport._insert(can, [mdb.TYPE, html.PLUG].concat(args))
		})
	},
})
Volcanos(chat.ONFIGURE, {
	image: function(can, value) { return {view: wiki.CONTENT, list: [{img: can.misc.Resource(can, value.text)}]} },
	text: function(can, value) {
		return {view: wiki.CONTENT, list: [{text: value.text||"[未知消息]"}], _init: function(target) {
			if (value.display) { var msg = can.request(); msg.Echo(value.text), can.onmotion.clear(can, target)
				var height = can.onexport.plugHeight(can, value), width = can.onexport.plugWidth(can, value)
				can.onappend.plugin(can, {title: value.name, index: "can._plugin", args: value.args, height: height, display: value.display, msg: msg}, function(sub) {
					sub.onimport.size(sub, height, width)
					delete(sub._legend.onclick)
				}, target)
			}
		}}
	},
	plug: function(can, value) { var height = can.onexport.plugHeight(can, value), width = can.onexport.plugWidth(can, value)
		return {view: wiki.CONTENT, style: {height: height+2}, _init: function(target) { value.type = chat.STORY
			var list = can.core.Split(can.ConfSpace()||can.misc.Search(can, ice.POD)||"", ".")
			var _list = can.core.Split(value.direct == "recv"? can.db.zone.target: "", ".")
			can.base.isIn(_list[0], "ops", "dev") && (list.pop(), _list.shift())
			value._space = list.concat(_list).join(".").replaceAll("..", "."), value._commands = {direct: value.direct, target: can.db.zone.target}
			value.title = value.name; if (value.text) { var msg = can.request(); msg._xhr = {responseText: value.text}, value.msg = msg, msg.Copy(JSON.parse(value.text)) }
			can.onappend.plugin(can, value, function(sub) { sub.onimport.size(sub, height, width, false)
				sub.Conf("_plugin_action", [{view: "item.button.localCreate.icons.state", _init: function(target) {
					can.page.Append(can, target, [{icon: icon.localCreate, title: "localCreate", onclick: function(event) {
						can.core.Next(sub._msg.IsDetail()? [sub._msg.TableDetail()]: sub._msg.Table(), function(value, next, index, list) { can.user.toastProcess(can, "create "+index+"/"+list.length, sub.ConfIndex())
							can.runAction(can.request(event, sub.Option(), value), ctx.RUN, ["", sub.ConfIndex(), mdb.CREATE], function() { next() })
						}, function() {
							can.user.toastSuccess(can, mdb.CREATE)
							can.onappend._float(can, sub.ConfIndex(), [])
						})
					}}])
				}}])
				sub.onexport.output = function() { sub.onimport.size(sub, height, width, false)
					can.page.style(can, target, html.HEIGHT, sub._target.offsetHeight+2)
					can.page.style(can, target, html.HEIGHT, sub._target.offsetHeight+2, html.WIDTH, sub._target.offsetWidth)
				}
			}, target)
		}}
	},
})
Volcanos(chat.ONINPUTS, {
	_show: function(event, can, msg, target, name) {
		function show(value) {
			can.showIcons(value.hash, value.icons||"usr/icons/Messages.png", value.title||can.base.trimPrefix(value.zone, "ops."))
		}
		can.page.Appends(can, can._output, msg.Table(function(value) {
			return {view: html.ITEM, list: [
				{img: can.misc.Resource(can, value.icons||"usr/icons/Messages.png")},
				{text: value.title||can.base.trimPrefix(value.zone, "ops.")},
			], onclick: function(event) { show(value) }}
		}))
	},
})
