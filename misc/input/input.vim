
" 变量定义
func! InputDefine(name, value)
	if !exists("name") | exec "let " . a:name . " = \"" . a:value . "\"" | endif
endfunc

" 输出日志
call InputDefine("g:InputLog", "input.log")
fun! InputLog(txt)
	call writefile([strftime("%Y-%m-%d %H:%M:%S ") . join(a:txt, "")], g:InputLog, "a")
endfun

" 输入转换
call InputDefine("g:InputTrans", "localhost:9020/code/input/")
fun! InputTrans(code)
    let res = []
    for line in split(system("curl -s " . g:InputTrans . a:code), "\n")
        let word = split(line, " ")
        if len(word) > 1 | call extend(res, [word[1]]) | endif
    endfor
    let res = extend(res, [a:code])
    return l:res
endfun

" 输入补全
fun! InputComplete(firststart, base)
    call InputLog(["complete", a:base, "(", col("."), ",", line("."), ")", getline(".")])
    if a:firststart
        " locate the start of the word
        let line = getline('.')
        let start = col('.') - 1
        while start > 0 && line[start - 1] =~ '\a'
            let start -= 1
        endwhile
        return start
    else
        " find months matching with "a:base"
        " retu
        return InputTrans(a:base)
    endif
endfun
set completefunc=InputComplete

" autocmd InsertEnter * call ShySync("insert")
" autocmd InsertCharPre * call InputCheck()
