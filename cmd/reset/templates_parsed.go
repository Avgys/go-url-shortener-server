package main

import "text/template"

var (
	mainTmpl        = template.Must(template.New("main").Parse(templateStr))
	arrayTmpl       = template.Must(template.New("array").Parse(arrayResetTemplate))
	arrayZeroTmpl   = template.Must(template.New("arrayZero").Parse(arrayZeroTemplate))
	mapTmpl         = template.Must(template.New("map").Parse(mapResetTemplate))
	chanCloseTmpl   = template.Must(template.New("chan").Parse(chanCloseResetTemplate))
	valueTmpl       = template.Must(template.New("value").Parse(valueResetTemplate))
	callResetTmpl   = template.Must(template.New("callReset").Parse(callResetTemplate))
	zeroStructTmpl  = template.Must(template.New("zeroStruct").Parse(zeroStructTemplate))
	pointerResetTmpl = template.Must(template.New("pointer").Parse(pointerResetTemplate))
)
