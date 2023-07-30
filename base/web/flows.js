Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.onmotion.clear(can), can.ui = can.onappend.layout(can), can.onmotion.hidden(can, can.ui.profile)
		if (can.Option(mdb.ZONE)) { return can.onmotion.hidden(can, can.ui.project), can.onimport._content(can, msg) } can.onimport._project(can, msg)
	},
	_project: function(can, msg) { var item; msg.Table(function(value) {
		var _item = can.onimport.item(can, value, function(event) {
			if (can.onmotion.cache(can, function(data, old) {
				if (old) { data[old] = {_content_plugin: can._content_plugin, _profile_plugin: can._profile_plugin} }
				var back = data[value.zone]; if (back) { can._content_plugin = back._content_plugin, can._profile_plugin = back._profile_plugin }
				return value.zone
			}, can.ui.content)) { return }
			can.run(event, [value.zone], function(msg) { msg.Table(function() { msg.Push(mdb.ZONE, value.zone) }), can.onimport._content(can, msg) })
		}, null, can.ui.project); item = can.Option(mdb.ZONE) == value.zone? _item: item||_item
	}), item && item.click() },
	_content: function(can, msg, zone) { can.onappend.plugin(can, {index: web.WIKI_DRAW, display: "/plugin/local/wiki/draw.js", style: "output"}, function(sub) {
		sub.onexport.output = function(_sub, _msg) { can._content_plugin = _sub, can.onimport.layout(can)
			sub.Action(svg.GO, "manual")
			sub.Action(ice.MODE, web.RESIZE)
			var list = {}; msg.Table(function(value) { list[value.hash] = value })
			var root = {}; can.core.Item(list, function(key, item) { if (!item.prev && !item.from) { root = item } item.prev && (list[item.prev].next = item), item.from && (list[item.from].to = item) })
			var margin = 20, width = 200, height = 100
			function show(root, _next, x, y) { can.onimport._block(can, _sub, root, x, y)
				var to = {x: 0, y: 0}, next = {x: 0, y: 0}
				if (_next) {
					if (root.to) {
						can.onimport._flows(can, _sub, [{x: x+width/2, y: y+height-margin}, {x: x+width/2, y: y+height+margin}])
						var to = show(root.to, false, x, y+height)
					}
					if (root.next) {
						can.onimport._flows(can, _sub, [{x: x+width-margin, y: y+height/2}, {x: x+width+margin, y: y+height/2}])
						var next = show(root.next, true, x+(to.y+1)*width, y)
					}
				} else {
					if (root.next) {
						can.onimport._flows(can, _sub, [{x: x+width-margin, y: y+height/2}, {x: x+width+margin, y: y+height/2}])
						var next = show(root.next, true, x, y+height)
					}
					if (root.to) {
						can.onimport._flows(can, _sub, [{x: x+width/2, y: y+height-margin}, {x: x+width/2, y: y+height+margin}])
						var to = show(root.to, false, x, y+(next.y+1)*height)
					}
				}
				return {x: next.x, y: to.y}
			} show(root, true, 0, 0)
			can.onappend.table(can, msg, null, can.ui.display), can.onmotion.hidden(can, _sub._action), can.onimport.layout(can)
		}, sub.run = function(event, cmds, cb) { cb(can.request(event)) }
	}, can.ui.content) },
	_flows: function(can, _sub, points) { _sub.onimport.draw(_sub, {shape: svg.LINE, points: points}) },
	_block: function(can, _sub, item, x, y) { var margin = 20, width = 200, height = 100
		var rect = _sub.onimport.draw(_sub, {shape: svg.RECT, points: [{x: x+margin, y: y+margin}, {x: x+width-margin, y: y+height-margin}]}); item.status && rect.Value("class", item.status)
		var text = _sub.onimport.draw(_sub, {shape: svg.TEXT, points: [{x: x+width/2, y: y+height/2}], style: {inner: item.index}}); item.status && text.Value("class", item.status)
		rect.onclick = text.onclick = function(event) { switch (_sub.svg.style.cursor) {
			case "e-resize": can.Update(can.request(event, {prev: item.hash, zone: item.zone||can.Option(mdb.ZONE)}), [ctx.ACTION, mdb.INSERT]); break
			case "s-resize": can.Update(can.request(event, {from: item.hash, zone: item.zone||can.Option(mdb.ZONE)}), [ctx.ACTION, mdb.INSERT]); break
		} can.onkeymap.prevent(event) }
		rect.oncontextmenu = text.oncontextmenu = function(event) { can.user.carteItem(event, can, item) }
	},
	layout: function(can) {
		if (can.page.isDisplay(can.ui.profile)) { var profile = can._profile_plugin
			can.page.style(can, profile._output, html.MAX_WIDTH, "")
			var width = can.base.Max(profile._target.offsetWidth+1, (can.ConfWidth()-can.ui.project.offsetWidth)/2)
			can.page.styleWidth(can, can.ui.profile, width)
		}
		if (can.page.isDisplay(can.ui.display)) {
			can.page.styleHeight(can, can.ui.display, 0)
			can.page.SelectChild(can, can.ui.display, "table", function(target) {
				can.page.styleHeight(can, can.ui.display, can.base.Max(target.offsetHeight, can.ConfHeight()/2))
			})
		}
		can.ui.layout(can.ConfHeight(), can.ConfWidth()), profile && profile.onimport.size(profile, can.ui.profile.offsetHeight, width-1, true)
		can._content_plugin && can._content_plugin.ui.layout(can.ConfHeight()-can.ui.display.offsetHeight-0*html.ACTION_HEIGHT, can.ConfWidth()-can.ui.project.offsetWidth)
	},
}, [""])
Volcanos(chat.ONACTION, {
	plugin: function(event, can, msg) {
		if (can.onmotion.cache(can, function() { return can.core.Keys(msg.Option(mdb.ZONE), msg.Option(mdb.HASH)) }, can.ui.profile)) { return }
		can.onappend.plugin(can, {index: msg.Option(ctx.INDEX), args: msg.Option(ctx.ARGS)}, function(sub) {
			can._profile_plugin = sub
			sub.onexport.output = function() { can.onmotion.toggle(can, can.ui.profile, true), can.onimport.layout(can) }
			sub.onaction._close = function() { can.onmotion.hidden(can, can.ui.profile), can.onimport.layout(can)  }
		}, can.ui.profile)
	},
})
