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
"syntax keyword eulangKeywords nop push drop dup
"syntax keyword eulangKeywords plusi minusi multi divi modi
"syntax keyword eulangKeywords multu divu modu
"syntax keyword eulangKeywords plusf minusf multf divf
"syntax keyword eulangKeywords jmp jmp_if halt swap not
"syntax keyword eulangKeywords eqi gei gti lei lti nei
"syntax keyword eulangKeywords equ geu gtu leu ltu neu
"syntax keyword eulangKeywords eqf gef gtf lef ltf nef
"syntax keyword eulangKeywords ret call native
"syntax keyword eulangKeywords andb orb xor shr shl notb
"syntax keyword eulangKeywords read8u read16u read32u read64u
"syntax keyword eulangKeywords read8i read16i read32i read64i
"syntax keyword eulangKeywords write8 write16 write32 write64
"syntax keyword eulangKeywords i2f u2f f2i f2u

syntax keyword eulangLoopKeywords from to do

syntax keyword eulangFuncDef func

" Comments
syntax region eulangCommentLine start=";" end="$"   contains=eulangTodos
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

let b:current_syntax = "eulang"
