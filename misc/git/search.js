Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		if (can.isZoneMode()) { return can.onimport._vimer_zone(can, msg, can._output) }
		can.onappend.style(can, "card", can._output)
		can.page.Appends(can, can._output, msg.Table(function(value) {
			return {view: [[html.ITEM, value.status]], list: [
				{view: [wiki.TITLE, html.DIV], list: [{img: value.avatar_url}, {view: "name", list: [
					{view: "name", list: [
						{text: [value.name, "", "name"]},
					]},
					{view: "label", list: [
						{icon: "bi bi-file-earmark-code"}, {text: value.language},
						{icon: "bi bi-share"}, {text: value.forks_count},
						{icon: "bi bi-star"}, {text: value.stars_count},
						{icon: "bi bi-folder2"}, {text: value.size},
						{icon: "bi bi-clock"}, {text: value.updated_at},
					]}
				]}]}, {view: [wiki.CONTENT, html.DIV, value.description]},
				{view: html.ACTION, inner: value.action, _init: function(target) { can.onappend.mores(can, target, value, 5) }},
			]}
		})), can.onappend.board(can, msg)
		can.onimport.layout(can)
	},
	layout: function(can) { can.onlayout.expand(can, can._output, 360) },
}, [""])