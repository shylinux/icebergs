Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { if (can.isCmdMode()) { can.onappend.style(can, html.OUTPUT) }
		can.ui = can.onappend.layout(can), can.onimport._project(can, msg)
	},
	_project: function(can, msg) { var select, current = can.db.hash[0]||ice.DEV
		can.isCmdMode() && can.page.insertBefore(can, [{view: "title", list: [
			{icon: "bi bi-three-dots", onclick: function() { can.onmotion.toggle(can, can.ui.profile), can.onimport.layout(can) }},
			{text: "message"||can.ConfIndex(), onclick: function(event) {
				can._legend.onclick(event)
			}},
			{icon: "bi bi-plus-lg", onclick: function(event) {
				can.Update(event, [ctx.ACTION, mdb.CREATE])
			}},
		]}], can.ui.project.firstChild, can.ui.project)
		msg.Table(function(value) {
			var _target = can.page.Append(can, can.ui.project, [{view: html.ITEM, list: [
				{img: can.misc.Resource(can, value.icons||"usr/icons/Messages.png")}, {view: "container", list: [
					{view: "title", list: [{text: value.name||"some"}, {text: [can.base.TimeTrim(value.time), "", mdb.TIME]}]},
					{view: "content", list: [{text: value.text||"[未知消息]"}]},
				]},
			], onclick: function(event) { can.isCmdMode() && can.misc.SearchHash(can, value.name), can.onimport._switch(can, false)
				can.db.zone = value, can.db.hash = value.hash, can.onmotion.select(can, can.ui.project, html.DIV_ITEM, _target)
				if (can.onmotion.cache(can, function() { return value.name }, can._status, can.ui.content, can.ui.profile, can.ui.display)) { return can.onimport.layout(can) }
				can.run(event, [value.hash], function(msg) { can.onappend._status(can, msg.Option(ice.MSG_STATUS))
					can.onimport._display(can), can.onimport._content(can, msg), can.onimport.layout(can)
				})
			}}])._target; select = (value.name == current? _target: select)||_target
		}), can.user.isMobile? can.onimport._switch(can, true): select && select.click()
	},
	_content: function(can, msg) {
		can.ui.title = can.page.Appends(can, can.ui.content, [{view: "title", list: [
			{icon: "bi bi-chevron-left", onclick: function() { can.onimport._switch(can, true) }},
			{text: can.db.zone.name},
			{icon: "bi bi-three-dots", onclick: function() { can.onmotion.toggle(can, can.ui.profile), can.onimport.layout(can) }},
		]}])._target
		can.ui.message = can.page.Append(can, can.ui.content, [{view: "list"}])._target
		msg.Table(function(value) { var myself = value.username == can.user.info.username
			can.page.Append(can, can.ui.message, [{view: [[html.ITEM, mdb.TIME], "", can.base.TimeTrim(value.time)]}])
			can.page.Append(can, can.ui.message, [{view: [[html.ITEM, value.type, myself? "myself": ""]], list: [
				{img: can.misc.Resource(can, (value.avatar == can.db.zone.name? "": value.avatar)||can.db.zone.icons||"usr/icons/Messages.png")},
				{view: "container", list: [{text: [value.usernick, "", "from"]}, can.onfigure[value.type](can, value)]},
			]}])
		})
	},
	_display: function(can, msg) {
		can.page.Appends(can, can.ui.display, [
			{view: "toolkit", list: can.core.Item(can.ondetail, function(icon, cb) { return {icon: icon, onclick: function(event) { cb(event, can) }} }) },
			{type: html.TEXTAREA, onkeyup: function(event) { if (event.key == "Enter" && event.ctrlKey) {
				can.onimport._insert(can, [mdb.TYPE, "text", mdb.TEXT, event.target.value])
			} }},
		]), can.onmotion.toggle(can, can.ui.display, true)
	},
	_insert: function(can, args) {
		can.runAction(event, mdb.INSERT, [can.db.hash].concat(args), function() {
			can.run(event, [can.db.hash], function(msg) { can.onappend._status(can, msg.Option(ice.MSG_STATUS))
				can.onimport._content(can, msg), can.onimport.layout(can)
			})
		})
	},
	_switch: function(can, project) { if (!can.user.isMobile) { return }
		can.page.style(can, can.ui.project, html.WIDTH, can.ConfWidth())
		can.page.style(can, can.ui.project, "flex", "0 0 "+can.ConfWidth()+"px")
		can.onmotion.toggle(can, can.ui.project, project)
		can.onmotion.toggle(can, can.ui.content, !project)
		can.onmotion.toggle(can, can.ui.display, !project)
		can.onimport.layout(can)
	},
	layout: function(can) { can.ui.layout(can.ConfHeight(), can.ConfWidth())
		can.ui.title && can.page.style(can, can.ui.message, html.HEIGHT, can.ui.content.offsetHeight-can.ui.title.offsetHeight)
			can.ui.message && (can.ui.message.scrollTop += 10000)
	},
}, [""])
Volcanos(chat.ONDETAIL, {
	"bi bi-mic": function(event, can) {},
	"bi bi-card-image": function(event, can) {
		can.user.input(event, can, [mdb.ICONS], function(args) {
			can.onimport._insert(can, [mdb.TYPE, "image"].concat([mdb.TEXT, args[1]]))
		})
	},
	"bi bi-camera": function(event, can) {},
	"bi bi-camera-video": function(event, can) {},
	"bi bi-file-earmark": function(event, can) {},
	"bi bi-geo-alt": function(event, can) {},
	"bi bi-window": function(event, can) {
		can.user.input(event, can, [web.SPACE, ctx.INDEX, ctx.ARGS], function(args) {
			can.onimport._insert(can, [mdb.TYPE, "plug"].concat(args))
		})
	},
})
Volcanos(chat.ONFIGURE, {
	text: function(can, value) {
		return {view: "content", list: [{text: value.text||"[未知消息]"}]}
	},
	image: function(can, value) {
		return {view: "content", list: [{img: can.misc.Resource(can, value.text)}]}
	},
	plug: function(can, value) { value.type = "story"
		var height = can.base.Min(can.ui.content.offsetHeight-210, 240)
		var height = can.base.Max(320, height, height/(can.base.isIn(value.index, "iframe")? 1: 2)),
			width = can.ui.content.offsetWidth-(can.user.isMobile? 60: 180)
		return {view: "content", style: {height: height, width: width}, _init: function(target) {
			can.onappend._plugin(can, value, {height: height, width: width}, function(sub) {
				sub.onexport.output = function() { sub.onimport.size(sub, height, width)
					can.page.style(can, target, html.HEIGHT, sub._target.offsetHeight, html.WIDTH, sub._target.offsetWidth)
				}
			}, target)
		}}
	},
})
