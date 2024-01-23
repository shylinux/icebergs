Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		can.ui = can.onappend.layout(can), can.onappend.style(can, "output card", can.ui.content), can.onimport._project(can, msg)
		can.onmotion.delay(can, function() { can.onimport.layout(can) })
	},
	_project: function(can, msg) { var select
		msg.Table(function(value) { if (value["client.type"] != nfs.REPOS) { return } value.name = `${value["client.name"]}`
			var _target = can.onimport.item(can, value, function(event) {
				if (can.onmotion.cache(can, function() { return value.name }, can.ui.content)) { return can.onimport.layout(can) }
				can.run(event, [value.name], function(msg) { can.onimport._content(can, msg), can.onimport.layout(can) })
			}, function() {}, can.ui.project); select = (value.name == ice.DEV? _target: select)||_target
		}), select && select.click()
	},
	_content: function(can, msg) { var year = new Date().getFullYear()+"-"
		can.page.Appends(can, can.ui.content, msg.Table(function(value) { if (value.type != web.WORKER) { return }
			var icon = value.icon; if (can.base.beginWith(value.icon, nfs.PS)) { icon = value.origin+value.icon }
			return {view: [[html.ITEM, value.status]], list: [
				{view: [wiki.TITLE, html.DIV], list: [{img: icon}, {view: mdb.NAME, list: [
					{view: mdb.NAME, list: [{text: [value.name, "", mdb.NAME]}]},
					{view: "label", list: [
						// {icon: "bi bi-file-earmark-code"}, {text: value.language||"None"},
						// {icon: "bi bi-share"}, {text: value.forks_count||"0"},
						// {icon: "bi bi-star"}, {text: value.stars_count||"0"},
						{icon: "bi bi-folder2"}, {text: value.version.split("-").slice(0, 2).join("-")},
						{icon: "bi bi-clock"}, {text: can.base.trimPrefix(value.time.split(":").slice(0, 2).join(":"), year)},
					]}
				]}]}, {view: [wiki.CONTENT, html.DIV, value.description]},
				{view: html.ACTION, inner: value.action, _init: function(target) { can.onappend.mores(can, target, value, 5) }},
			]}
		}))
	},
	layout: function(can) {
		can.page.style(can, can.ui.project, html.HEIGHT, can.ConfHeight())
		can.page.style(can, can.ui.content, html.HEIGHT, can.ConfHeight())
		can.page.style(can, can._output, html.HEIGHT, can.ConfHeight(), html.WIDTH, can.ConfWidth())
		can.onlayout.expand(can, can.ui.content, can.user.isMobile && !can.user.isLandscape()? can.ConfWidth(): 320)
	},
}, [""])
