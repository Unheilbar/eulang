" Vim syntax file
" Language: Eulang

" Usage Instructions
" Put this file in .vim/syntax/eulang.vim
" and add in your .vimrc file the next line:
" autocmd BufRead,BufNewFile *.eul, set filetype=eulang

if exists("b:current_syntax")
  finish
endif

syntax keyword eulangTodos TODO XXX FIXME NOTE

" Language keywords

syntax keyword eulangLoopKeywords from to do

syntax keyword eulangFuncDef func
syntax keyword eulangCond if else elif
syntax keyword eulangBoolLiteral true false
syntax keyword eulangLoopKey while for range
syntax keyword eulangStruct var
syntax keyword eulangModifier external internal
syntax keyword eulangTypes i64 bytes32 address

" Comments
syntax region eulangCommentLine start="//" end="$"   contains=eulangTodos
syntax region eulangDirective start="%" end=" "

syntax match eulangLabel		"[a-z_][a-z0-9_]*:"he=e-1

" Numbers
syntax match eulangDecInt display "\<[0-9][0-9_]*"
syntax match eulangHexInt display "\<0[xX][0-9a-fA-F][0-9_a-fA-F]*"
syntax match eulangFloat  display "\<[0-9][0-9_]*\%(\.[0-9][0-9_]*\)"

" Strings
syntax region eulangString start=/\v"/ skip=/\v\\./ end=/\v"/
syntax region eulangString start=/\v'/ skip=/\v\\./ end=/\v'/

" Set highlights
highlight default link eulangTodos Todo
highlight default link eulangKeywords Identifier
highlight default link eulangCommentLine Comment
highlight default link eulangDirective PreProc
highlight default link eulangLoopKeywords PreProc
highlight default link eulangDecInt Number
highlight default link eulangHexInt Number
highlight default link eulangFloat Float
highlight default link eulangString String
highlight default link eulangLabel Label
highlight default link eulangFuncDef Function
highlight default link eulangCond Conditional
highlight default link eulangBoolLiteral Boolean
highlight default link eulangloopkey Repeat
highlight default link eulangStruct  Structure
highlight default link eulangModifier  StorageClass
highlight default link eulangTypes  Type


let b:current_syntax = "eulang"
