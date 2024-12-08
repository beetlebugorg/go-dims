grammar Geometry;

// Parser Rules
start    : geometry EOF;
geometry : dimension (offset)? flags?;
dimension : (width ('x' height?)?) | ('x' height) ;
width    : NUMBER (percent)? ;
height   : NUMBER (percent)?  ;
percent  : PERCENT ;
offset   : PLUS (offsetx) (PLUS offsety)? ;
offsetx  : NUMBER (percent)? ;
offsety  : NUMBER (percent)? ;
flags    : BANG | GT | LT ;

// Lexer Rules
NUMBER   : INT ;

// Fragments
fragment INT : [0-9]+ ;

GT : '>' ;
LT : '<' ;
BANG : '!' ;
PLUS : '+' ;
PERCENT : '%' ;