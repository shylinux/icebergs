Volcanos("onimport", {help: "导入数据", list: [], _init: function(can, msg, list, cb, target) {
        can.onmotion.clear(can)
        can.onappend.table(can, msg)
        can.onappend.board(can, msg)
        can.base.isFunc(cb) && cb(msg)
    },
})
Volcanos("onaction", {help: "控件交互", list: [],
    scanQRCode: function(event, can, button) { can.user.agent.scanQRCode(function(text, data) {
        var msg = can.request(event, data)
        can.run(event, can.base.Simple(ctx.ACTION, data.action||button, data), function(msg) {
            can.user.toast(can, text, "添加成功"), can.Update()
        }, true)
    }, can) },
    scanQRCode0: function(event, can) { can.user.agent.scanQRCode() },
})
Volcanos("onexport", {help: "导出数据", list: []})

