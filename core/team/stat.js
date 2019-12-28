Volcanos("onimport", {help: "导入数据", list: [],
    init: function(can, msg, cb, output, option) {output.innerHTML = "";
        msg.Table(function(value, index) {
            can.page.Append(can, output, [{view: ["stat", "div", "hello"]}])
        })
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

