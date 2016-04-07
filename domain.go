package onto

import (
	"strings"
)

type VarType int

const (
	Unknown VarType = iota
	Int
	String
	Bool
	Float
)

type IntRange struct {
	From int
	To   int
	Up   bool
	Down bool
}

func Up(from int) IntRange {
	return IntRange{
		From: from,
		Up:   true,
	}
}

func Down(from int) IntRange {
	return IntRange{
		To:   from,
		Down: true,
	}
}

func Between(from, to int) IntRange {
	return IntRange{
		From: from,
		To:   to,
	}
}

type DVar struct {
	Type       VarType
	Whole      *DomainClass
	Other      *DomainClass
	Comment    string
	Name       string
	Column     string
	Default    string
	Range      IntRange
	IsAuto     bool
	IsEditable bool
}

func (v DVar) GoType() string {
	switch {
	case v.Whole != nil || v.Other != nil:
		return "int"
	case v.Type == Int:
		return "int"
	case v.Type == String:
		return "string"
	case v.Type == Bool:
		return "bool"
	case v.Type == Float:
		return "float64"
	}
	return ""
}

func (v DVar) UpName() string {
	if len(v.Name) < 1 {
		return ""
	}
	return strings.ToUpper(v.Name[:1]) + v.Name[1:]
}

func (v DVar) IsDefault() (bool, string) {
	if len(v.Default) == 0 {
		return false, ""
	}
	switch v.Default[0] {
	case '$':
		return true, v.Default[0:]
	case '#':
		return true, `"` + v.Default[0:] + `"`
	default:
		return true, v.Default
	}
}

type DomainClass struct {
	Name      string
	Table     string
	Comment   string
	Arguments []DVar
	Editables []DVar
	External  []DVar
	Autos     []DVar
}

func joinDVars(arrs ...[]DVar) []DVar {
	retv := []DVar{}
	for _, a := range arrs {
		retv = append(retv, a...)
	}
	return retv
}

func joinDVarsCond(p func(DVar) bool, arrs ...[]DVar) []DVar {
	retv := []DVar{}
	for _, a := range arrs {
		for _, x := range a {
			if p(x) {
				retv = append(retv, x)
			}
		}
	}
	return retv
}

func (d *DomainClass) AllVars() []DVar {
	return joinDVars(d.Arguments, d.Editables, d.External, d.Autos)
}

func (d *DomainClass) DefaultArgs() []DVar {
	return joinDVarsCond(func(v DVar) bool {
		b, _ := v.IsDefault()
		return b
	}, d.Arguments)
}

func (d *DomainClass) InitvalVars() []DVar {
	return joinDVarsCond(func(v DVar) bool {
		b, _ := v.IsDefault()
		return b
	}, d.AllAutos(), d.AllEditables())
}

func (d *DomainClass) AllAutos() []DVar {
	return joinDVars(d.Autos,
		joinDVarsCond(func(v DVar) bool {
			return v.IsAuto
		}, d.Arguments, d.Editables))
}

func (d *DomainClass) AllEditables() []DVar {
	return joinDVars(d.Editables,
		joinDVarsCond(func(v DVar) bool {
			return v.IsEditable
		}, d.Arguments, d.Autos))
}

func (d *DomainClass) Range(over func() []DVar, do func(v DVar, cmt, name, tp string) bool) bool {
	for i, v := range over() {
		t := v.GoType()
		cmt := "//" + v.Comment
		if i == 0 {
			cmt = "\n" + cmt
		}
		if len(t) > 0 {
			if !do(v, cmt, v.Name, t) {
				return false
			}
		}
	}
	return true
}
