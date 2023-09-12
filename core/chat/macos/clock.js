Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) { can.page.requireDraw(can, function() {
		can.onmotion.hidden(can, can._status), can.base.isFunc(cb) && cb(msg)
		can.onmotion.hidden(can, can._action), can.onimport._show(can)
	}) },
	_show: function(can) { can.svg.Value("dominant-baseline", "middle")
		can.onmotion.clear(can, can.svg), can.svg.Val(html.HEIGHT, can.ConfHeight()), can.svg.Val(html.WIDTH, can.ConfWidth())
		var x = can.ConfWidth()/2, y = can.ConfHeight()/2, r = can.base.Max(can.ConfHeight(), can.ConfWidth())/2-80, c = {x: x, y: y}
		function pos(r, angle) { angle -= 90; return {x: x + r * Math.cos(angle * Math.PI / 180), y: y + r * Math.sin(angle * Math.PI / 180)} }
		function line(g, c, p) { return can.onimport.draw(can, {shape: svg.LINE, points: [c, p]}, g) }
		function group(name) { return can.onimport.group(can, name) }
		var number = group("number"), second = group("second"), minute = group("minute"), hour = group("hour")
		for (var i = 1; i <= 12; i++) { var p = pos(r, 360/12*i); can.onimport.draw(can, {shape: svg.TEXT, points: [p], style: {inner: i+""}}, number) }
		can.core.Timer({internal: 100}, function(){ var t = new Date()
			can.onmotion.clear(can, second), can.onmotion.clear(can, minute), can.onmotion.clear(can, hour)
			line(hour, c, pos(r*0.6, t.getHours()%12*360/12+t.getMinutes()*30/60))
			line(minute, c, pos(r*0.8, t.getMinutes()*360/60))
			line(second, c, pos(r, t.getSeconds()*360/60))
		})
	},
}, [""])
