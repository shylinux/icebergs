" 基础函数{{{
" 变量定义
func! ShyDefine(name, value)
	if !exists(a:name) | exec "let " . a:name . " = \"" . a:value . "\"" | endif
endfunc

" 后端通信
call ShyDefine("g:ctx_url", (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9020") . "/code/vim/")
fun! ShySend(cmd, arg)
    if has_key(a:arg, "sub") && a:arg["sub"] != "" | let temp = tempname()
        call writefile(split(a:arg["sub"], "\n"), temp, "b")
        let a:arg["sub"] = "@" . temp
    endif

    let a:arg["pwd"] = getcwd() | let a:arg["buf"] = bufname("%") | let a:arg["row"] = line(".") | let a:arg["col"] = col(".")
    let args = "" | for k in sort(keys(a:arg)) | let args = args . " -F '" . k . "=" . a:arg[k] . "' " | endfor
    return system("curl -s " . g:ctx_url . a:cmd . args . " 2>/dev/null")
endfun
" }}}
" 功能函数{{{
" 数据同步
fun! ShySync(target)
    if bufname("%") == "ControlP" | return | end

    if a:target == "read" || a:target == "write"
        call ShySend("sync", {"cmds": a:target, "arg": expand("<afile>")})
    elseif a:target == "insert"
        call ShySend("sync", {"cmds": a:target, "sub": getreg(".")})
    elseif a:target == "exec"
        call ShySend("sync", {"cmds": a:target, "arg": getcmdline()})
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
        if match(line, '\s*ice ') >= 0 | return match(line, "ice ") | endif
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
    return ShyInput(a:base)
endfun
set completefunc=ShyComplete

" 收藏列表
call ShyDefine("g:favor_note", "")
fun! ShyFavor()
    let tab_list = split(ShySend("favor", {"cmds": "select"}), "\n")
    let tab = tab_list[inputlist(tab_list)-1]
    let g:favor_note = input("note: ", g:favor_note)
    call ShySend("favor", {"cmds": "insert", "tab": tab, "note": g:favor_note, "arg": getline("."), "row": getpos(".")[1], "col": getpos(".")[2]})
endfun
fun! ShyFavors()
    let tab_list = split(ShySend("favor", {"cmds": "topic"}), "\n")
    let tab = tab_list[inputlist(tab_list)-1]

    let res = split(ShySend("favor", {"tab": tab}), "\n")
    let page = "" | let note = ""
    for i in range(0, len(res)-1, 2)
        if res[i] != page
            if note != "" | lexpr note | lopen | let note = "" | endif
            execute exists(":TabooOpen")? "TabooOpen " . res[i]: "tabnew"
        endif
        let page = res[i] | let note .= res[i+1] . "\n"
    endfor
    if note != "" | lexpr note | let note = "" | endif

    let view = inputlist(["列表", "默认", "垂直", "水平"])
    for i in range(0, len(res)-1, 2) | if i < 5
        if l:view == 4 | split | lnext | elseif l:view == 3 | vsplit | lnext | endif
    endif | endfor
    botright lopen | if l:view  == 1 | only | endif
endfun

" 文件搜索
call ShyDefine("g:grep_dir", "./")
fun! ShyGrep(word)
    let g:grep_dir = input("dir: ", g:grep_dir, "file")
    " execute "grep -rn --exclude tags --exclude '*.tags' '\\<" . a:word . "\\>' " . g:grep_dir
    execute "grep -rn '\\<" . a:word . "\\>' " . g:grep_dir
    copen
endfun
" }}}
" 事件回调{{{
autocmd! BufReadPost * call ShySync("bufs")
autocmd! BufReadPost * call ShySync("read")
autocmd! BufWritePre * call ShySync("write")
autocmd! InsertLeave * call ShySync("insert")
autocmd! CmdlineLeave * call ShySync("exec")
"}}}
" 按键映射{{{
nnoremap <C-G><C-G> :call ShyGrep(expand("<cword>"))<CR>
nnoremap <C-G><C-F> :call ShyFavor()<CR>
nnoremap <C-G>f :call ShyFavors()<CR>
inoremap <C-K> <C-X><C-U>
"}}}

