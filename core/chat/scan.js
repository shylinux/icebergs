Volcanos(chat.ONIMPORT, {help: "导入数据", _init: function(can, msg, cb, target) {
        can.onmotion.clear(can), can.base.isFunc(cb) && cb(msg)
        can.onappend.table(can, msg)
        can.onappend.board(can, msg)
    },
})
Volcanos(chat.ONACTION, {help: "控件交互",
    scanQRCode0: function(event, can) { can.user.agent.scanQRCode() },
    scanQRCode: function(event, can, button) { can.user.agent.scanQRCode(function(text, data) {
        can.runAction(can.request(event, data), data.action||button [], function(msg) {
            can.user.toastSuccess(can, text), can.Update()
        }, true)
    }, can) },
})

