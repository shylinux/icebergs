#!/bin/sh

# 连接配置
if [ "${ctx_dev}" = "" ] || [ "${ctx_dev}" = "-" ]; then
    ctx_dev="http://localhost:9020"
fi
ctx_sid=${ctx_sid:=""}
ctx_share=${ctx_share:=""}
ctx_sh=`ps|grep $$|grep -v grep`

# 请求配置
ctx_head=${ctx_head:="Content-Type: application/json"}
ctx_curl=${ctx_curl:="curl"}
ctx_url=$ctx_dev"/code/zsh/"
ctx_cmd=${ctx_cmd:=""}

# 输出配置
ctx_silent=${ctx_silent:=""}
ctx_err=${ctx_err:="/dev/null"}
ctx_welcome=${ctx_welcome:="^_^  \033[32mWelcome to Context world\033[0m  ^_^"}
ctx_goodbye=${ctx_goodbye:="v_v  \033[31mGoodbye to Context world\033[0m  v_v"}

# 输出信息
ShyRight() {
    [ "$1" = "" ] && return 1
    [ "$1" = "0" ] && return 1
    [ "$1" = "false" ] && return 1
    return 0
}
ShyEcho() {
    ShyRight "$ctx_silent" || echo "$@"
}
ShyLog() {
    echo "$@" > $ctx_err
}

# 发送数据
ShyLine() {
    echo "$*"|sed -e 's/\"/\\\"/g' -e 's/\n/\\n/g'
}
ShyJSON() {
    [ $# -eq 1 ] && echo \"`ShyLine "$1"`\" && return
    echo -n "{"
    while [ $# -gt 1 ]; do
        echo -n \"`ShyLine "$1"`\"\:\"`ShyLine "$2"`\"
        shift 2 && [ $# -gt 1 ] && echo -n ","
    done
    echo -n "}"
}
ShyPost() {
    ctx_cmd="$1" && shift
    case $ctx_sh in
        *zsh)
            ShyJSON SHELL "${SHELL:=bash}" pwd "${PWD:=/root}" sid "${ctx_sid:=0}" cmds "$@"|read data
            ;;
        *)
            local data=`ShyJSON SHELL "${SHELL:=bash}" pwd "${PWD:=/root}" sid "${ctx_sid:=0}" cmds "$@"`
            ;;
    esac
    echo $data > $ctx_err
    ${ctx_curl} -s "${ctx_url}${ctx_cmd}" -H "${ctx_head}" -d "${data}"
}

# 终端登录
ShyHelp() {
    ShyPost help "$@"
}
ShyLogin() {
    HOST=`hostname` ctx_sid=`ShyPost login "" share "${ctx_share}" pid "$$" pane "${TMUX_PANE}" hostname "${HOST}" username "${USER}"`
    echo "${ctx_welcome}"
    echo "${ctx_dev}"
    echo "sid: ${ctx_sid}"
}
ShyLogout() {
    echo ${ctx_goodbye} && [ "$ctx_sid" != "" ] && ShyPost logout
}

# 发送文件
ShyDownload() {
    ${ctx_curl} -s "${ctx_url}download" -F "cmds=$1" \
        -F "SHELL=${SHELL}" -F "pwd=${PWD}" -F "sid=${ctx_sid}"
}
ShyUpload() {
    ${ctx_curl} -s "${ctx_url}upload" -F "upload=@$1" \
        -F "SHELL=${SHELL}" -F "pwd=${PWD}" -F "sid=${ctx_sid}"
}
ShySend() {
    local TEMP=`mktemp /tmp/tmp.XXXXXX` && "$@" > $TEMP
    ShyRight "$ctx_silent" || cat $TEMP
    ${ctx_curl} -s "${ctx_url}sync" -F "cmds=$1" -F "cmds=$*" -F "sub=@$TEMP" \
        -F "SHELL=${SHELL}" -F "pwd=${PWD}" -F "sid=${ctx_sid}"
}

# 同步数据
ShySync() {
    case "$1" in
        "history")
            ctx_end=`history|tail -n1|awk '{print $1}'`
            ctx_begin=${ctx_begin:=$ctx_end}
            ctx_count=`expr $ctx_end - $ctx_begin`
            ShyEcho "sync $ctx_begin-$ctx_end count $ctx_count to $ctx_dev"
            history|tail -n $ctx_count |while read line; do
                ShyPost history "$line"
            done
            ctx_begin=$ctx_end
            ;;
        ps) ShySend ps -ef ;;
        *) ShySend "$@"
    esac
}
ShySyncs() {
    case "$1" in
        "base")
            ShySync df &>/dev/null
            ShySync ps &>/dev/null
            ShySync env &>/dev/null
            ShySync free &>/dev/null
            ShySync history
            ;;
        *)
    esac
}
ShyInput() {
    if [ "$1" = "line" ] ; then
        READLINE_LINE=`ShyPost input "$1" line "$READLINE_LINE" point "$READLINE_POINT"`
    else
        COMPREPLY=(`ShyPost input "$COMP_WORDS" line "$COMP_LINE" index "$COMP_CWORD" break "$COMP_WORDBREAKS"`)
    fi
}
ShyFavor() {
    cmd=$1; [ "$READLINE_LINE" != "" ] && set $READLINE_LINE && READLINE_LINE=""
    if [ "$cmd" = "sh" ] ; then
        # 查看收藏
        ctx_word="sh"
        shift && [ "$1" != "" ] && ctx_tab="$1"
        shift && [ "$1" != "" ] && ctx_note="$1"
    else
        # 添加收藏
        [ "$1" != "" ] && ctx_word="$*" || ctx_word=`history|tail -n1|head -n1|sed -e 's/^[\ 0-9]*//g'`
    fi
    ShyPost favor "${ctx_word}" tab "${ctx_tab}" note "${ctx_note}"
}

ShyUpgrade() {
    file=auto.sh && [ "$1" != "" ] && file=$1
    ${ctx_curl} -s $ctx_dev/publish/$file > $file && source auto.sh
}
ShyInit() {
    [ "$ctx_begin" = "" ] && ctx_begin=`history|tail -n1|awk '{print $1}'`

    case "${SHELL##*/}" in
        "zsh") PROMPT='%![%*]%c$ ' ;;
        *) PS1="\!-$$-\u@\h[\t]\W\$ " ;;
    esac

    if bind &>/dev/null; then
        # bash
        # bind 'TAB:complete' 
        bind 'TAB:menu-complete' 
        complete -F ShyInput word
        bind -x '"\C-K":ShyInput line' 
        bind -x '"\C-G\C-F":ShyFavor'
        bind -x '"\C-GF":ShyFavor sh'
        bind -x '"\C-Gf":ShyFavor sh'
        bind -x '"\C-G\C-G":ShySync history'

    elif bindkey &>/dev/null; then
        # zsh
        bindkey -s '\C-G\C-R' 'ShySyncs base\n'
        setopt nosharehistory
    fi

    echo "url: ${ctx_url}"
    echo "pid: $$"
    echo "begin: ${ctx_begin}"
    echo "share: ${ctx_share}"
    echo "pane: $TMUX_PANE"
}
ShyInit && trap ShyLogout EXIT
