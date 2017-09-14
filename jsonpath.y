%{
package jsonpath

type jsonpathSymUnion struct {
  val interface{}
}

%}

%union {
  val jsonPathNode
  vals []jsonPathNode
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
%token <val> NEQ NUMBER NULL
%token <val> OR
%token <str> STR
%token <val> STRICT STARTS
%token <val> TRUE TO
%token <val> UNKNOWN UNOT
%token <val> WITH
%token <val> EOF

%type <val> root
%type <val> expr
%type <val> primary
%type <val> variable
%type <val> literal
%type <val> member_accessor
%type <val> accessor_expr
%type <val> accessor
%type <val> member_accessor
%type <val> member_accessor_wildcard
%type <val> array_accessor
%type <vals> subscript_list
%type <val> subscript
%type <val> wildcard_array_accessor
%type <val> item_method
%type <val> method
%type <val> filter_expression
%type <val> predicate_primary
%type <val> delimited_predicate
%type <val> non_delimited_predicate
%type <val> comparison_pred
%type <val> exists_pred
%type <val> like_regex_pred
%type <str> like_regex_pattern
%type <str> like_regex_flag
%type <val> starts_with_pred
%type <val> is_unknown_pred

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
          $$ = []jsonPathNode{$1}
        }
        | subscript_list ',' subscript
        {
          $$ = append($1, $3)
        }

subscript:
     expr
    | expr TO expr
    {
      $$ = RangeNode{start: $1, end: $3}
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

predicate_primary:
    delimited_predicate
  | non_delimited_predicate

delimited_predicate:
  exists_pred
  | '(' predicate_primary ')'
  {
    $$ = ParenExpr{$2}
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
      $$ = BinExpr{t: eqBinOp, left: $1, right: $3}
    }
    | expr NEQ expr
    {
      $$ = BinExpr{t: neqBinOp, left: $1, right: $3}
    }
    | expr '>' expr
    {
      $$ = BinExpr{t: gtBinOp, left: $1, right: $3}
    }
    | expr '<' expr
    {
      $$ = BinExpr{t: ltBinOp, left: $1, right: $3}
    }
    | expr GTE expr
    {
      $$ = BinExpr{t: gteBinOp, left: $1, right: $3}
    }
    | expr LTE expr
    {
      $$ = BinExpr{t: lteBinOp, left: $1, right: $3}
    }
    | predicate_primary AND predicate_primary
    {
      $$ = BinExpr{t: andBinOp, left: $1, right: $3}
    }
    | predicate_primary OR predicate_primary
    {
      $$ = BinExpr{t: orBinOp, left: $1, right: $3}
    }
    | UNOT predicate_primary %prec UMINUS
    {
      $$ = UnaryExpr{t: unot, expr: $2}
    }

like_regex_pred:
  expr LIKE_REGEX like_regex_pattern FLAG like_regex_flag
  {
    $$ = LikeRegexNode{left: $1, pattern: $3, flag: &$5}
  }
  | expr LIKE_REGEX like_regex_pattern
  {
    $$ = LikeRegexNode{left: $1, pattern: $3}
  }

like_regex_pattern: STR

like_regex_flag: STR

starts_with_pred:
  expr STARTS WITH expr
  {
    $$ = StartsWithNode{left: $1, right: $4}
  }

is_unknown_pred:
  expr IS UNKNOWN
  {
    $$ = IsUnknownNode{expr: $1}
  }

%%
