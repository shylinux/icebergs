Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		msg.Echo("hello world")
	},
})
Volcanos(chat.ONACTION, {
	list: ["刷新", "扫码", "清屏", "登录"],
})
