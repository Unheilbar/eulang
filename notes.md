go compiler parse and build - go/src/go/build/read
go/src/cmd/compile/README.md guide into go compiler


### TODOs
##### high priority
 - [x] Add scopes and support for local variables
 - [x] Add support for compiling function arguments (parsing already introduced)
 - [x] Add types bytes32, address
 - [x] Add mapping for version storage write/read (without state backend yet)
 - [ ] Add return value for functions
 - [x] Add escape analysis for defining  storing variables to /version storage/permanent storage?/stack/memory
 - [x] Add binary operations add/sub/div/multi/mod
 - [x] Add comparison operations for strings, bytes32, etc.
 - [x] Add function visibility identifier (start with all function internal by default, external functions can be called only from outside)
 - [x] Add multiple arguments for writef native function
 - [ ] Add forward func declaration
 - [ ] Add params to external functions
 - [x] Add var assignment after declaration
 ##### low priotiy
 - [ ] Add i8, i64, i128
 - [ ] Add for loops
 - [ ] Add enums
 - [ ] Add choice to examples
 - [ ] Introduce test utility (run functions with prefix Test)
 - [x] Add 'import' directive
 - [ ] Add branching tokens support 'break', 'continue'
 - [x] Remove semicolons(;) in the end of lines
 - [x] Remove colon(:) before type
 ###### backlog
 - [ ] eulType should become expressions
 - [ ] add bad manners phrase generator for compile errors

