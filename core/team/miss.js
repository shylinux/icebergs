Volcanos("onimport", {help: "导入数据", list: [],
    init: function(can, msg, cb, output, option) {output.innerHTML = "";
        if (!msg.append || msg.append.length == 0) {
            var code = can.page.Append(can, output, [{view: ["code", "div", msg.Result()]}]).code;
            return typeof cb == "function" && cb(msg), code;
        }

        var table = can.page.AppendTable(can, output, msg, msg.append);
        table.oncontextmenu = function(event) {var target = event.target; var data = target.dataset;
            switch (event.target.tagName) {
                case "SPAN":
                    can.user.carte(event, shy("", can.ondetail, can.feature.detail || can.ondetail.list, function(event, cmd, meta) {
                        var id = data.id;
                        can.run(event, [id, cmd], function(msg) {
                            can.onimport.init(can, msg, cb, output, option)
                            can.user.toast(cmd+"成功");
                        }, true)
                    }))
                    event.stopPropagation()
                    event.preventDefault()
                    break
                case "TH":
                case "TR":
                case "TABLE":
            }
        }
        return typeof cb == "function" && cb(msg), table;
    },
})
Volcanos("onaction", {help: "组件交互", list: [],
})
Volcanos("onchoice", {help: "组件菜单", list: [],
})
Volcanos("ondetail", {help: "组件详情", list: [],
})
Volcanos("onexport", {help: "导出数据", list: [],
})

