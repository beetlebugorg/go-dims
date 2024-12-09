#!/bin/sh -x

antlr -Dlanguage=Go -no-visitor -package parser -o parser *.g4