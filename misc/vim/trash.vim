
" 自动刷新
let ShyComeList = {}
fun! ShyCome(buf, row, action, extra)
    " 低配命令
    if !exists("appendbufline")
		execute a:extra["row"]

		if a:extra["count"] > 0
			execute "+1,+" . a:extra["count"] ."delete"
		endif

        let a:extra["count"] = 0
        for line in reverse(split(ShySend("sync", {"cmds": "trans", "arg": getline(".")}), "\n"))
			let a:extra["count"] += 1
            call append(".", line)
        endfor
        return
    endif
    if a:action == "refresh"
        " 清空历史
        if a:extra["count"] > 0 | call deletebufline(a:buf, a:row+1, a:row+a:extra["count"]) | endif
        let a:extra["count"] = 0
    endif
    " 刷新命令
    for line in reverse(split(ShySend({"cmd": "trans", "arg": getbufline(a:buf, a:row)[0]}), "\n"))
        call appendbufline(a:buf, a:row, line)
        let a:extra["count"] += 1
    endfor
    " 插入表头
    call appendbufline(a:buf, a:row, strftime(" ~~ %Y-%m-%d %H:%M:%S"))
    let a:extra["count"] += 1
endfun
fun! ShyUpdate(timer)
    let what = g:ShyComeList[a:timer]
    call ShyLog("timer", a:timer, what)
    call ShyCome(what["buf"], what["row"], what["action"], what)
endfun
fun! ShyComes(action)
    if !exists("b:timer") | let b:timer = -1 | endif
    " 清除定时
    if b:timer > 0 | call timer_stop(b:timer) | let b:timer = -2 | return | endif
    " 添加定时
    let b:timer = timer_start(1000, funcref('ShyUpdate'), {"repeat": -1})
    let g:ShyComeList[b:timer] = {"buf": bufname("."), "row": line("."), "pre": getline("."), "action": a:action, "count": 0}
    call ShyLog("new timer", b:timer)
endfun
fun! ShyCheck(target)
    if a:target == "cache"
        call ShySync("bufs")
        call ShySync("regs")
        call ShySync("marks")
        call ShySync("tags")
    elseif a:target == "fixs"
        let l = len(getqflist())
        if l > 0
            execute "copen " . (l > 10? 10: l + 1)
            call ShySync("fixs")
		else
            cclose
        end
    end
endfun


" nnoremap <C-G><C-R> :call ShyCheck("cache")<CR>
nnoremap <C-G><C-K> :call ShyComes("refresh")<CR>
