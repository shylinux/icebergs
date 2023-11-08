Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) {
		// window.initialize = function() {
		// 	can.require(["location.js"], function() { can.onimport._layout_init(can, msg, function() { cb && cb(msg), can.onimport._content(can) }) })
		// }, can.require([can.Conf(nfs.SCRIPT)])
		can.require(["location.js"], function() { can.onimport._layout_init(can, msg, function() { cb && cb(msg), can.onimport._content(can) }) })
	},
	_content: function(can, item) {
		var map = new BMapGL.Map(can.ui.content)
		var point = new BMapGL.Point(116.404, 39.915)
		map.centerAndZoom(point, 15)
		map.enableScrollWheelZoom(true)
		map.setMapType(BMAP_EARTH_MAP)
		var scaleCtrl = new BMapGL.ScaleControl();  // 添加比例尺控件
		map.addControl(scaleCtrl);
		var zoomCtrl = new BMapGL.ZoomControl();  // 添加缩放控件
		map.addControl(zoomCtrl);
		var cityCtrl = new BMapGL.CityListControl();  // 添加城市列表控件
		map.addControl(cityCtrl);
		can.ui.map = map
	},
})
