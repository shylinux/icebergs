Volcanos("onimport", {help: "导入数据", list: [], _init: function(can, msg, list, cb, target) {
        can.onmotion.clear(can)
        can.onappend.table(can, msg)
        can.onappend.board(can, msg)
        can.base.isFunc(cb) && cb(msg)
    },
}, [""])
Volcanos("onaction", {help: "控件交互", list: ["play"],
    play: function(event, can) {
        can.page.Select(can, can._output, "div.item", function(item) {
            can.onmotion.hidden(can, item)
        })
        can.core.Next(can.page.Select(can, can._output, "div.item"), function(item, next) {
            can.onmotion.show(can, 300, next, item)
        }, function() {
            can.user.toast(can, "播放结束")
        })
    },
})
Volcanos("onexport", {help: "导出数据", list: []})
