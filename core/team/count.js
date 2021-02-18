Volcanos("onimport", {help: "导入数据", list: [], _init: function(can, msg, list, cb, target) {
        can.onmotion.clear(can)
        can.onappend.table(can, msg)
        can.onappend.board(can, msg.Result())
        typeof cb == "function" && cb(msg)
    },
}, ["count.css"])
Volcanos("onaction", {help: "控件交互", list: ["播放"],
    "播放": function(event, can) {
        can.page.Select(can, can._output, "div.item", function(item) {
            can.onmotion.hidden(can, item)
        })
        can.core.Next(can.page.Select(can, can._output, "div.item"), function(item, next) {
            can.onmotion.show(can, 300, next, item)
        }, function() {
            can.user.toast(can, "播放结束", can.Option("zone"))
        })
    },
})
Volcanos("onexport", {help: "导出数据", list: []})
