Volcanos("onimport", {help: "导入数据", list: [], _init: function(can, msg, list, cb, target) {
        can.onmotion.clear(can, target)
        can.ui = can.onlayout.profile(can)
        can.onmotion.hidden(can, can._action)
        can.onimport[can.Option("scale")||"week"](can, msg)
        can.page.Modify(can, can.ui.project, {style: {"max-height": can.ui.content.offsetHeight}})
        can.page.Modify(can, can.ui.profile, {style: {"max-height": can.ui.content.offsetHeight}})
        can.page.Modify(can, can.ui.profile, {style: {display: "block"}})
        can.base.isFunc(cb) && cb(msg)
    },
    _content: function(can, msg, head, list, key, get, set) {
        var hash = {}; msg.Table(function(value, index) {
            var k = key(can.base.Date(value.begin_time)); hash[k] = (hash[k]||[]).concat([value])
        })

        can.sup.task && (can.sup.task._target = null)
        var begin_time = can.base.Date(can.Option("begin_time"))
        can.page.Append(can, can.ui.content, [{view: [chat.CONTENT, html.TABLE], list: can.core.List(list, function(hour, row) {
            return {type: html.TR, list: can.core.List(head, function(week, col) {
                if (row == 0) { return {text: [week, html.TH]} }
                if (col == 0) { return {text: [hour, html.TD]} }
                return can.onimport._task(can, msg, get(begin_time, col, row, hash), set(begin_time, col, row))
            })}
        }) }])
        can.core.Timer(10, function() { if (!can.sup.task) { return }
            var target = can.sup.task._target||can.task._target
            can.sup.task = null, target.click()
            can.Status("count", msg.Length())
        })
    },
    _task: function(can, msg, list, time) { return {text: ["", html.TD],
        ondblclick: function(event) {
            can.onaction.insertTask(event, can, time+can.base.Time(null, "%y-%m-%d %H:%M:%S").slice(time.length))
        },
        ondrop: function(event) { event.preventDefault()
            can.drop(event, event.target, time)
        },
        ondragover: function(event) { event.preventDefault()
            can.page.Select(can, can.ui.content, html.TD, function(item) {
                can.page.ClassList[event.target == item? "add": "del"](can, item, "over")
            })
        },
        list: can.core.List(list, function(task) { return can.base.isString(task)? {text: [task, html.DIV, "date"]}:
            {text: [can.onexport[can.Action("view")||"text"](can, task), html.DIV, can.onexport.style(can, task)],
                ondragstart: function(event) { var target = event.target; can.drop = function(event, td, time) { td.append(target)
                    can.onaction.modifyTask(event, can, task, "begin_time", time+task.begin_time.slice(time.length), task.begin_time)
                } }, draggable: time != undefined,

                title: can.onexport.title(can, task), _init: function(target) {
                    var item = can.onappend.item(can, html.ITEM, {nick: task.name+":"+task.text}, function() {
                        can.core.Timer(10, function() { can.onmotion.select(can, can.ui.content, html.TD, target.parentNode) })
                        can.onimport._profile(can, task)
                    }, function() {

                    }, can.ui.project); task._target = target, can.task = can.task||task, can.sup.task = can.sup.task||task
                    can.sup.task.zone == task.zone && can.sup.task.id == task.id && (can.sup.task._target = target)
                    target.onclick = function(event) { item.click() }
                },
            }
        }),
    } },
    _profile: function(can, task) {
        function keys(task, key) { return [task.pod, task.zone, task.id, key].join(".") }

        if (can.sup.task) { if (can.sup.task.id == task.id) { return }
            can.page.Cache(keys(can.sup.task, chat.PROFILE), can.ui.profile, can.sup.task.id)
            can.page.Cache(keys(can.sup.task, chat.DISPLAY), can.ui.display, can.sup.task.id)
        }

        can.sup.task = task, can.Status(task)
        var profile = can.page.Cache(keys(task, "profile"), can.ui.profile)
        var display = can.page.Cache(keys(task, "display"), can.ui.display)
        if (profile || display) { return }

        task.extra && can.core.Item(can.base.Obj(task.extra), function(key, value) { task["extra."+key] = value }), delete(task.extra)
        var table = can.page.Appends(can, can.ui.profile, [{view: [chat.CONTENT, html.TABLE], list: [{th: ["key", "value"]}]}]).first
        can.core.Item(task, function(key, value) { key != "_target" && can.page.Append(can, table, [{
            td: [key, key == "pod" && value != ""? can.page.Format("a", can.user.MergeURL(can, {pod: value}), value): value],
            onclick: function(event) { if (event.target.type == "button") { var msg = can.request(event, can.sup.task)
                can.run(event, [ctx.ACTION, event.target.name], function(msg) { can.Update() }, true)
            } },
            ondblclick: function(event) { can.onmotion.modify(can, event.target, function(ev, value, old) {
                can.onaction.modifyTask(event, can, task, key, value)
            }) },
        }]) }), can.onimport._display(can, task)
    },
    _display: function(can, task) { if (!task["extra.cmd"]) { return }
        can.onappend.plugin(can, {type: "story", ctx: task["extra.ctx"], cmd: task["extra.cmd"], arg: task["extra.arg"]}, function(sub, meta) {
            sub.run = function(event, cmds, cb) { var msg = can.request(event, {"task.zone": task.zone, "task.id": task.id})
                can.run(event, can.misc.Concat([ctx.ACTION, cli.RUN, task["zone"], task["id"]], cmds), cb, true)
            }
        }, can.ui.display)
        can.page.Modify(can, can.ui.display, {style: {display: "block"}})
    },

    day: function(can, msg) {
        var head = ["hour", "task"]
        var list = [0]; for (var i = 7; i < 24; i++) { list.push(can.base.Number(i, 2)) }

        function key(time) { return can.base.Number(time.getHours(), 2) }
        function get(begin_time, col, row, hash) { return hash[list[row]] }
        function set(begin_time, col, row) { return can.base.Time(begin_time, "%y-%m-%d ")+list[row] }

        can.onimport._content(can, msg, head, list, key, get, set)
    },
    week: function(can, msg) {
        var head = ["hour"].concat(["周日", "周一", "周二", "周三", "周四", "周五", "周六"])
        var list = [0]; for (var i = 7; i < 24; i++) { list.push(can.base.Number(i, 2)) }

        function key(time) { return time.getDay()+" "+can.base.Number(time.getHours(), 2) }
        function get(begin_time, col, row, hash) { return hash[col-1+" "+list[row]] }
        function set(begin_time, col, row) { return can.base.Time(can.base.TimeAdd(begin_time, -begin_time.getDay()+col-1), "%y-%m-%d ")+list[row] }

        can.onimport._content(can, msg, head, list, key, get, set)
    },
    month: function(can, msg) {
        var head = ["order"].concat(["周日", "周一", "周二", "周三", "周四", "周五", "周六"])
        var list = [0]; for (var i = 1; i < 6; i++) { list.push(i) }

        function key(time) { return can.base.Time(time, "%y-%m-%d") }
        function get(begin_time, col, row, hash) {
            var begin = can.base.TimeAdd(begin_time, -(begin_time.getDate()-1))
            var last = can.base.TimeAdd(begin_time, -(begin_time.getDate()-1)-begin.getDay())
            var day = can.base.TimeAdd(last, (row-1)*7+col)
            return [day.getDate()+""].concat(hash[key(day)]||[])
        }
        function set(begin_time, col, row) {
            var begin = can.base.TimeAdd(begin_time, -(begin_time.getDate()-1))
            var last = can.base.TimeAdd(begin_time, -(begin_time.getDate()-1)-begin.getDay())
            var day = can.base.TimeAdd(last, (row-1)*7+col)
            return key(day)
        }

        can.onimport._content(can, msg, head, list, key, get, set)
    },
    year: function(can, msg) {
        var head = ["month"].concat(["周日", "周一", "周二", "周三", "周四", "周五", "周六"]);
        var list = [0]; for (var i = 1; i < 13; i++) { list.push(i) }

        function key(time) { return can.base.Time(time, "%y-%m ")+time.getDay() }
        function get(begin_time, col, row, hash) { return hash[begin_time.getFullYear()+"-"+can.base.Number(row, 2)+" "+(col-1)] }
        function set(begin_time, col, row) { return begin_time.getFullYear()+"-"+can.base.Number(list[row], 2) }

        can.onimport._content(can, msg, head, list, key, get, set)
    },
    long: function(can, msg) {
        var begin_time = can.base.Date(can.base.Time(can.Option("begin_time"), "%y-%m-%d %H:%M:%S"))
        var begin = begin_time.getFullYear() - 5

        var head = ["month"]; for (var i = 0; i < 10; i++) { head.push(begin+i) }
        var list = [0]; for (var i = 1; i < 13; i++) { list.push(i) }

        function key(time) { return can.base.Time(time, "%y-%m") }
        function get(begin_time, col, row, hash) { return hash[begin+col-1+"-"+can.base.Number(row, 2)] }
        function set(begin_time, col, row) { return begin+col-1+"-"+can.base.Number(row, 2) }

        can.onimport._content(can, msg, head, list, key, get, set)
    },
}, [""])
Volcanos("onaction", {help: "组件交互", list: [
        ["level", "all", "l1", "l2", "l3", "l4", "l5"],
        ["status", "all", "prepare", "process", "cancel", "finish"],
        ["score", "all", "s1", "s2", "s3", "s4", "s5"],
        ["view", "", "name", "text", "level", "score"],
    ],
    insertTask: function(event, can, time) { var msg = can.sup.request(event, {begin_time: time})
        can.user.input(event, can, can.Conf("feature.insert"), function(event, button, data, list) {
            can.run(event, can.base.Simple(ctx.ACTION, mdb.INSERT, "begin_time", time, data))
        })
    },
    modifyTask: function(event, can, task, key, value) { var msg = can.request(event, task)
        can.run(event, [ctx.ACTION, mdb.MODIFY, key, value, task[key]])
    },

    _filter: function(event, can, key, value) { var count = 0
        if (value == "all") {
            can.page.Select(can, can.ui.content, "div.item", function(item) {
                can.page.ClassList.del(can, item, "hidden"), count++
            })
        } else {
            can.page.Select(can, can.ui.content, "div.item", function(item) {
                can.page.ClassList.add(can, item, "hidden")
            })
            can.page.Select(can, can.ui.content, "div."+value, function(item) {
                can.page.ClassList.del(can, item, "hidden"), count++
            })
        }
        can.Action(key, value), can.Status("count", count)
    },
    level: function(event, can, key, value) { can.onaction._filter(event, can, key, value) },
    status: function(event, can, key, value) { can.onaction._filter(event, can, key, value) },
    score: function(event, can, key, value) { can.onaction._filter(event, can, key, value) },
    view: function(event, can, key, value) {
        can.Action(key, value)
        can.onmotion.clear(can, can.ui.project)
        can.onmotion.clear(can, can.ui.content)
        can.onimport[can.Option("scale")](can, can._msg)
    },
})
Volcanos("onexport", {help: "导出数据", list: ["count", "begin_time", "zone", "id", "type", "name", "text"],
    name: function(can, task) { return task.name },
    text: function(can, task) { return task.name+": "+(task.text||"") },
    level: function(can, task) { return "l-"+(task.level||3)+": "+(task.name||"") },
    score: function(can, task) { return "s-"+(task.level||3)+": "+(task.name||"") },
    title: function(can, task) { return task.zone+": "+(task.type||"") },
    style: function(can, task) { return ["item", task.status, "id"+task.id, "l"+(task.level||""), "s"+(task.score||"")].join(" ") },
})

