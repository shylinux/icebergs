# {{title "hello world"}}

{{shell "开机时长" "" "uptime"}}

{{title "premenu"}}

## {{chapter "项目总览"}}
{{order "总览" `
volcano iceberg
context toolkit
preload appframe
`}}

{{table "总览" `
volcano iceberg
context toolkit
preload appframe
`}}

## {{chapter "项目详情"}}
{{chain "详情" `
context
    volcanos
        proto.js
        frame.js bg blue
            Page
            Pane
            Plugin
            Inputs
            Output
        order.js
    icebergs bg blue
        type.go
        base.go bg red
            code
            wiki
            chat
            team
            mall
        conf.go
    toolkits
        type.go
        core.go bg blue
            Split
            Parse
            Value
            Fetch
            Favor
        misc.go
`}}

{{title "endmenu"}}

