%{
package jsonpath

import (
  "fmt"
  "regexp"
)

type jsonpathSymUnion struct {
  val interface{}
}

%}

%union {
  expr jsonPathExpr
  pred jsonPathPred
  vals []jsonPathNode
  regexp *regexp.Regexp
  ranges []RangeSubscriptNode
  rangeNode RangeSubscriptNode
  accessor accessor
  str string
}

%token <val> AND
%token <val> EQ EXISTS
%token <val> FALSE FLAG
%token <val> FUNC_DATETIME FUNC_KEYVALUE
%token <val> FUNC_TYPE FUNC_SIZE FUNC_DOUBLE FUNC_CEILING FUNC_FLOOR FUNC_ABS
%token <val> GTE
%token <str> IDENT IS
%token <val> LAST LAX LTE LIKE_REGEX
%token <val> NEQ
%token <expr> NUMBER NULL
%token <val> OR
%token <str> STR
%token <val> STRICT STARTS
%token <val> TRUE TO
%token <val> UNKNOWN UNOT
%token <val> WITH
%token <val> EOF

%type <expr> root
%type <expr> expr
%type <expr> primary
%type <expr> variable
%type <expr> literal
%type <expr> accessor_expr
%type <accessor> accessor
%type <accessor> member_accessor
%type <accessor> member_accessor_wildcard
%type <accessor> array_accessor
%type <ranges> subscript_list
%type <rangeNode> subscript
%type <accessor> wildcard_array_accessor
%type <accessor> item_method
%type <accessor> method
%type <accessor> filter_expression
%type <pred> predicate_primary
%type <pred> delimited_predicate
%type <pred> non_delimited_predicate
%type <pred> comparison_pred
%type <pred> exists_pred
%type <pred> like_regex_pred
%type <str> like_regex_pattern
%type <str> like_regex_flag
%type <pred> starts_with_pred
%type <pred> is_unknown_pred

%left OR
%left AND
%left EQ '>' '<' GTE LTE
%left '+' '-'
%left '*' '/' '%'
%left UMINUS
%left '.'

%%

root:
    expr
    {
      yylex.(*tokenStream).expr = $1
    }

expr:
    '(' expr ')'
    {
      $$ = ParenExpr{$2}
    }
    | expr '+' expr
    {
      $$ = BinExpr{t: plusBinOp, left: $1, right: $3}
    }
    | expr '-' expr
    {
      $$ = BinExpr{t: minusBinOp, left: $1, right: $3}
    }
    | expr '*' expr
    {
      $$ = BinExpr{t: timesBinOp, left: $1, right: $3}
    }
    | expr '/' expr
    {
      $$ = BinExpr{t: divBinOp, left: $1, right: $3}
    }
    | expr '%' expr
    {
      $$ = BinExpr{t: modBinOp, left: $1, right: $3}
    }
    | '-' expr %prec UMINUS
    {
      $$ = UnaryExpr{t: uminus, expr: $2}
    }
    | '+' expr %prec UMINUS
    {
      $$ = UnaryExpr{t: uplus, expr: $2}
    }
    | accessor_expr

/* 6.9 */
primary:
    literal
  | variable

/* 6.9.1 */
literal:
     NUMBER
    | TRUE { $$ = BoolExpr{val: true} }
    | FALSE { $$ = BoolExpr{val: false} }
    | NULL { $$ = NullExpr{} }
    | STR { $$ = StringExpr{$1} }

/* 6.9.2 */
variable:
    IDENT  { $$ = VariableExpr{name: $1} }
    | '@' 
    {
      $$ = VariableExpr{name: "@"}
    }
    | LAST { $$ = LastExpr{} }

/* 6.10 */
accessor_expr:
        primary
      | accessor_expr accessor
      {
        $$ = AccessExpr{ left: $1, right: $2 }
      }

accessor:
        member_accessor
      | member_accessor_wildcard
      | array_accessor
      | wildcard_array_accessor
      | filter_expression
      | item_method

/* 6.10.1 */
member_accessor:
           '.' IDENT { $$ = DotAccessor{val: $2} }
           | '.' STR { $$ = DotAccessor{val: $2, quoted: true} }

/* 6.10.2 */
member_accessor_wildcard:
           '.' '*' { $$ = MemberWildcardAccessor{} }

/* 6.10.3 */
array_accessor:
        '[' subscript_list ']'
        {
          $$ = ArrayAccessor{subscripts: $2}
        }

subscript_list:
        subscript
        {
          $$ = []RangeSubscriptNode{$1}
        }
        | subscript_list ',' subscript
        {
          $$ = append($1, $3)
        }

subscript:
     expr
    {
      $$ = RangeSubscriptNode{start: $1, end: nil}
    }
    | expr TO expr
    {
      $$ = RangeSubscriptNode{start: $1, end: $3}
    }

/* 6.10.4 */
wildcard_array_accessor:
  '[' '*' ']'
  {
    $$ = WildcardArrayAccessor{}
  }

/* 6.11 */
item_method:
    '.' method
    {
      $$ = $2
    }

method:
      FUNC_TYPE '(' ')' { $$ = FuncNode{typeFunction, nil} }
      | FUNC_SIZE '(' ')' { $$ = FuncNode{sizeFunction, nil} }
      | FUNC_DOUBLE '(' ')' { $$ = FuncNode{doubleFunction, nil} }
      | FUNC_CEILING '(' ')' { $$ = FuncNode{ceilingFunction, nil} }
      | FUNC_FLOOR '(' ')' { $$ = FuncNode{floorFunction, nil} }
      | FUNC_ABS '(' ')' { $$ = FuncNode{absFunction, nil} }
      | FUNC_DATETIME '(' STR ')' { $$ = FuncNode{datetimeFunction, StringExpr{$3}} }
      | FUNC_KEYVALUE '(' ')' { $$ = FuncNode{keyvalueFunction, nil} }

/* 6.13 */
filter_expression:
   '?' '(' predicate_primary ')'
    {
      $$ = FilterNode{pred: $3}
    }
  | '?' '(' expr ')'
    {
      n := FormatNode($3)
      yylex.(*tokenStream).err = fmt.Errorf("filter expressions cannot be raw json values - if you expect `%s` to be boolean true, write `%s == true`", n, n)
      return 0
    }


predicate_primary:
    delimited_predicate
  | non_delimited_predicate

delimited_predicate:
  exists_pred
  | '(' predicate_primary ')'
  {
    $$ = ParenPred{$2}
  }

non_delimited_predicate:
  comparison_pred
  | like_regex_pred
  | starts_with_pred
  | is_unknown_pred

exists_pred:
  EXISTS '(' expr ')'
  {
    $$ = ExistsNode{expr: $3}
  }

comparison_pred:
    expr EQ expr
    {
      $$ = BinPred{t: eqBinOp, left: $1, right: $3}
    }
    | expr NEQ expr
    {
      $$ = BinPred{t: neqBinOp, left: $1, right: $3}
    }
    | expr '>' expr
    {
      $$ = BinPred{t: gtBinOp, left: $1, right: $3}
    }
    | expr '<' expr
    {
      $$ = BinPred{t: ltBinOp, left: $1, right: $3}
    }
    | expr GTE expr
    {
      $$ = BinPred{t: gteBinOp, left: $1, right: $3}
    }
    | expr LTE expr
    {
      $$ = BinPred{t: lteBinOp, left: $1, right: $3}
    }
    | predicate_primary AND predicate_primary
    {
      $$ = BinLogic{t: andBinOp, left: $1, right: $3}
    }
    | predicate_primary OR predicate_primary
    {
      $$ = BinLogic{t: orBinOp, left: $1, right: $3}
    }
    | UNOT predicate_primary %prec UMINUS
    {
      $$ = UnaryNot{expr: $2}
    }

like_regex_pred:
  expr LIKE_REGEX like_regex_pattern FLAG like_regex_flag
  {
    pattern, err := regexp.Compile($3)
    if err != nil {
      yylex.(*tokenStream).err = err
      return 1
    }
    $$ = LikeRegexNode{left: $1, rawPattern: $3, pattern: pattern, flag: &$5}
  }
  | expr LIKE_REGEX like_regex_pattern
  {
    pattern, err := regexp.Compile($3)
    if err != nil {
      yylex.(*tokenStream).err = err
      return 1
    }
    $$ = LikeRegexNode{left: $1, pattern: pattern, rawPattern: $3}
  }

like_regex_pattern:
  STR

like_regex_flag: STR

starts_with_pred:
  expr STARTS WITH expr
  {
    $$ = StartsWithNode{left: $1, right: $4}
  }

is_unknown_pred:
  '(' predicate_primary ')' IS UNKNOWN
  {
    $$ = IsUnknownNode{expr: $2}
  }

%%
