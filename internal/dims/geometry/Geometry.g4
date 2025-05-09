grammar Geometry;

// Parser Rules
start    : geometry EOF;
geometry : dimension (offset)? flags?;
dimension : (width ('x' height?)?) | ('x' height) ;
width    : NUMBER (PERCENT)? ;
height   : NUMBER (PERCENT)? ;
offset   : PLUS (offsetx) (PLUS offsety)? ;
offsetx  : NUMBER (PERCENT)? ;
offsety  : NUMBER (PERCENT)? ;
flags    : BANG | GT | LT ;

// Lexer Rules
NUMBER   : INT | FLOAT;

// Fragments
INT : MINUS? [0-9]+ ;
FLOAT : MINUS? INT DOT INT ;

GT : '>';
LT : '<' ;
BANG : '!' ;
PLUS : '+' ;
PERCENT : '%' ;
MINUS : '-' ;
DOT : '.' ;