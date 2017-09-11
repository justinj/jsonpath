%{
package jsonpath

import "fmt"

type jsonpathSymUnion struct {
  val interface{}
}

%}

%union {
  b bool
  union jsonpathSymUnion
}

%token <b> AT AND
%token <b> DOLLAR DOT DIV
%token <b> EQ
%token <b> GT GTE
%token <b> LEFTSQ LEFTPAREN LT LTE
%token <b> MINUS
%token <b> NUMBER
%token <b> OR
%token <b> PLUS
%token <b> RIGHTSQ RIGHTPAREN
%token <b> STAR STR
%token <b> TIMES
%token <b> WORD
%token <b> EOF

%type <val> root

%%

root:
    NUMBER
    {
      $$ = Number{val: 1}
      fmt.Println("GOT NUMBER!!!")
    }

%%
