Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		if (can.isCmdMode()) { can.onappend.style(can, html.OUTPUT) }
		can.ui = can.onappend.layout(can), can.onimport._project(can, msg)
	},
	_project: function(can, msg) { var select, current = can.db.hash[0]||ice.DEV
		can.page.insertBefore(can, [{view: wiki.TITLE, list: [
			{icon: "bi bi-three-dots", onclick: function() { can._legend.onclick(event) }},
			{text: "message"||can.ConfIndex(), onclick: function(event) { can._legend.onclick(event) }},
			{icon: "bi bi-plus-lg", onclick: function(event) { can.Update(event, [ctx.ACTION, mdb.CREATE]) }},
		]}], can.ui.project.firstChild, can.ui.project)
		msg.Table(function(value) {
			var _target = can.page.Append(can, can.ui.project, [{view: html.ITEM, list: [
				{img: can.misc.Resource(can, value.icons||"usr/icons/Messages.png")}, {view: html.CONTAINER, list: [
					{view: wiki.TITLE, list: [{text: value.name||"[未命名]"}, {text: [can.base.TimeTrim(value.time), "", mdb.TIME]}]},
					{view: wiki.CONTENT, list: [{text: value.text||"[未知消息]"}]},
				]},
			], onclick: function(event) { can.isCmdMode() && can.misc.SearchHash(can, value.name), can.onimport._switch(can, false)
				can.db.zone = value, can.db.hash = value.hash, can.onmotion.select(can, can.ui.project, html.DIV_ITEM, _target)
				if (can.onmotion.cache(can, function(save, load) {
					can.ui.message && save({title: can.ui.title, message: can.ui.message, scroll: can.ui.message.scrollTop})
					return load(value.name, function(bak) { can.ui.title = bak.title, can.ui.message = bak.message
						can.onmotion.delay(can, function() { can.ui.message.scrollTop = bak.scroll })
					})
				}, can.ui.content, can.ui.profile, can.ui.display, can._status)) { return can.onimport.layout(can) }
				can.run(can.request(event, {"cache.limit": 10}), [value.hash], function(msg) {
					can.onimport._display(can), can.onimport._content(can, msg)
				})
			}}])._target; select = (value.name == current? _target: select)||_target
		}), can.user.isMobile? can.onimport._switch(can, true): select && select.click()
	},
	_content: function(can, msg) {
		can.ui.title = can.page.Appends(can, can.ui.content, [{view: wiki.TITLE, list: [
			{icon: "bi bi-chevron-left", onclick: function() { can.onimport._switch(can, true) }},
			{text: can.db.zone.name},
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
			var myself = value.username == can.user.info.username, time = can.base.TimeTrim(value.time)
			var t = new Date(value.time); if (!last || (t - last > 3*60*1000)) { last = t
				can.page.Append(can, can.ui.message, [{view: [[html.ITEM, mdb.TIME], "", time]}])
			}
			can.page.Append(can, can.ui.message, [{view: [[html.ITEM, value.type, myself? "myself": ""]], list: [
				{img: can.misc.Resource(can, (value.avatar == can.db.zone.name? "": value.avatar)||can.db.zone.icons||"usr/icons/Messages.png")},
				{view: html.CONTAINER, list: [{text: [value.usernick, "", nfs.FROM]}, can.onfigure[value.type](can, value)]},
			]}])
		}), can.onappend._status(can, msg.Option(ice.MSG_STATUS)), can.onimport.layout(can)
		if (can.Status(mdb.TOTAL) > can.db.zone.id) { can.onimport._request(can) }
		can.onmotion.delay(can, function() { can.ui.message && (can.ui.message.scrollTop += 10000) })
	},
	_request: function(can) {
		can.run(can.request(event, {"cache.begin": parseInt(can.db.zone.id||0)+1, "cache.limit": 10}), [can.db.hash], function(msg) {
			can.onimport._message(can, msg)
		})
	},
	_insert: function(can, args) {
		can.runAction(event, mdb.INSERT, [can.db.hash].concat(args), function() {
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
	},
}, [""])
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
	text: function(can, value) { return {view: wiki.CONTENT, list: [{text: value.text||"[未知消息]"}]} },
	plug: function(can, value) { var height = can.base.Min(can.ui.content.offsetHeight-210, 240)
		var height = can.base.Max(320, height, height/(can.base.isIn(value.index, html.IFRAME)? 1: 2)), width = can.ui.content.offsetWidth-(can.user.isMobile? 60: 180)
		return {view: wiki.CONTENT, style: {height: height, width: width}, _init: function(target) { value.type = chat.STORY
			can.onappend.plugin(can, value, function(sub) {
				sub.onexport.output = function() { sub.onimport.size(sub, height, width)
					can.page.style(can, target, html.HEIGHT, sub._target.offsetHeight, html.WIDTH, sub._target.offsetWidth)
				}
			}, target)
		}}
	},
})
Volcanos(chat.ONINPUTS, {
	_show: function(event, can, msg, target, name) {
		function show(value) {
			if (target == target.parentNode.firstChild) { can.ui = can.ui||{}
				can.ui.img = can.page.insertBefore(can, [{type: html.IMG}], target)
				can.ui.span = can.page.insertBefore(can, [{type: html.SPAN}], target)
				can.onappend.style(can, mdb.ICONS, can.page.parentNode(can, target, html.TR))
				can.page.style(can, target, html.COLOR, html.TRANSPARENT)
			}
			can.ui.img.src = can.misc.Resource(can, value.icons), can.ui.span.innerText = value.name
			target.value = value.hash, can.onmotion.hidden(can, can._target)
		}
		can.page.Appends(can, can._output, msg.Table(function(value) {
			return value.name && {view: html.ITEM, list: [
				{img: can.misc.Resource(can, value.icons)}, {text: value.name},
			], onclick: function(event) { show(value) }}
		}))
	},
})
