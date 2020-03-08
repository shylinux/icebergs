" 变量定义
func! ShyDefine(name, value)
	if !exists(a:name) | exec "let " . a:name . " = \"" . a:value . "\"" | endif
endfunc

" 输出日志
call ShyDefine("g:ShyLog", "/dev/null")
fun! ShyLog(...)
    call writefile([strftime("%Y-%m-%d %H:%M:%S ") . join(a:000, " ")], g:ShyLog, "a")
endfun

" 后端通信
call ShyDefine("g:ctx_sid", "")
call ShyDefine("g:ctx_url", (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9020") . "/code/vim/")
fun! ShySend(cmd, arg)
    if has_key(a:arg, "sub") && a:arg["sub"] != ""
        let temp = tempname()
        call writefile(split(a:arg["sub"], "\n"), temp, "b")
        let a:arg["sub"] = "@" . temp
    endif

    let a:arg["sid"] = g:ctx_sid
    let a:arg["pwd"] = getcwd()
    let a:arg["buf"] = bufname("%")
    let a:arg["row"] = line(".")
    let a:arg["col"] = col(".")
    let args = ""
    for k in sort(keys(a:arg))
        let args = args . " -F '" . k . "=" . a:arg[k] . "' "
    endfor
    return system("curl -s " . g:ctx_url . a:cmd . args . " 2>/dev/null")
endfun

" 用户登录
fun! ShyLogout()
    if g:ctx_sid != "" | call ShySend("logout", {}) | endif
endfun
fun! ShyLogin()
    let g:ctx_sid = ShySend("login", {"share": $ctx_share, "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER})
endfun
fun! ShyHelp()
    echo ShySend("help", {})
endfun
call ShyLogin()

" 数据同步
fun! ShySync(target)
    if bufname("%") == "ControlP" | return | end

    if a:target == "read" || a:target == "write"
        call ShySend("sync", {"cmds": a:target, "arg": expand("<afile>")})
    elseif a:target == "exec"
        call ShySend("sync", {"cmds": a:target, "arg": getcmdline()})
    elseif a:target == "insert"
        call ShySend("sync", {"cmds": a:target, "sub": getreg(".")})
    else
        let cmd = {"bufs": "buffers", "regs": "registers", "marks": "marks", "tags": "tags", "fixs": "clist"}
        call ShySend("sync", {"cmds": a:target, "sub": execute(cmd[a:target])})
    endif
endfun

" 输入补全
fun! ShyInput(code)
    return split(ShySend("input", {"cmds": a:code, "pre": getline("."), "row": line("."), "col": col(".")}), "\n")
endfun
fun! ShyComplete(firststart, base)
    if a:firststart | let line = getline('.') | let start = col('.') - 1
        " 命令位置
        if match(line, '\s*ice ') == 0 | return match(line, "ice ") | endif
        " 符号位置
        if line[start-1] !~ '\a' | return start - 1 | end
        " 单词位置
        while start > 0 && line[start - 1] =~ '\a' | let start -= 1 | endwhile
        return start
    endif

    " 符号转换
    if a:base == "," | return ["，", ","] | end
    if a:base == "." | return ["。", "."] | end
    if a:base == "\\" | return ["、", "\\"] | end
    " 单词转换
    let list = ShyInput(a:base)
    call ShyLog("trans", a:base, list)
    return list
endfun
set completefunc=ShyComplete
set encoding=utf-8
colorscheme torte
highlight Pmenu ctermfg=cyan ctermbg=darkblue
highlight PmenuSel ctermfg=darkblue ctermbg=cyan
highlight Comment ctermfg=cyan ctermbg=darkblue

" 收藏列表
call ShyDefine("g:favor_tab", "")
call ShyDefine("g:favor_note", "")
fun! ShyFavor()
    let g:favor_tab = input("tab: ", g:favor_tab)
    let g:favor_note = input("note: ", g:favor_note)
    call ShySend("favor", {"tab": g:favor_tab, "note": g:favor_note, "arg": getline("."), "row": getpos(".")[1], "col": getpos(".")[2]})
endfun
fun! ShyFavors()
    let res = split(ShySend("favor", {"tab": input("tab: ", g:favor_tab)}), "\n")
    let page = "" | let note = ""
    for i in range(0, len(res)-1, 2)
        if res[i] != page
            if note != "" | lexpr note | lopen | let note = "" | endif
            execute exists(":TabooOpen")? "TabooOpen " . res[i]: "tabnew"
        endif
        let page = res[i] | let note .= res[i+1] . "\n"
    endfor
    if note != "" | lexpr note | lopen | let note = "" | endif
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





" 任务列表
fun! ShyTask()
    call ShySend({"cmd": "tasklet", "arg": input("target: "), "sub": input("detail: ")})
endfun

" 标签列表
fun! ShyGrep(word)
    if !exists("g:grep_dir") | let g:grep_dir = "./" | endif
    let g:grep_dir = input("dir: ", g:grep_dir, "file")
    execute "grep -rn --exclude tags --exclude '*.tags' '\<" . a:word . "\>' " . g:grep_dir
endfun
fun! ShyTag(word)
    execute "tag " . a:word
endfun

" 自动刷新
let ShyComeList = {}
fun! ShyCome(buf, row, action, extra)
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
    " 低配命令
    if !exists("appendbufline")
        for line in reverse(split(ShySend({"cmd": "trans", "arg": getline(".")}), "\n"))
            call append(".", line)
        endfor
        return
    endif
    if !exists("b:timer") | let b:timer = -1 | endif
    " 清除定时
    if b:timer > 0 | call timer_stop(b:timer) | let b:timer = -2 | return | endif
    " 添加定时
    let b:timer = timer_start(1000, funcref('ShyUpdate'), {"repeat": -1})
    let g:ShyComeList[b:timer] = {"buf": bufname("."), "row": line("."), "pre": getline("."), "action": a:action, "count": 0}
    call ShyLog("new timer", b:timer)
endfun

" 事件回调
autocmd! VimLeave * call ShyLogout()
autocmd! BufReadPost * call ShySync("bufs")
autocmd! BufReadPost * call ShySync("read")
autocmd! BufWritePre * call ShySync("write")
autocmd! CmdlineLeave * call ShySync("exec")
" autocmd! QuickFixCmdPost * call ShyCheck("fixs")
autocmd! InsertLeave * call ShySync("insert")

" 按键映射
nnoremap <C-G><C-F> :call ShyFavor()<CR>
nnoremap <C-G>f :call ShyFavors()<CR>
" nnoremap <C-G><C-G> :call ShyGrep(expand("<cword>"))<CR>
" nnoremap <C-G><C-R> :call ShyCheck("cache")<CR>
" nnoremap <C-G><C-T> :call ShyTask()<CR>
nnoremap <C-G><C-K> :call ShyComes("refresh")<CR>
inoremap <C-K> <C-X><C-U>

