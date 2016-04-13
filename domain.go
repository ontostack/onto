package onto

import (
	"strings"
)

type VarType int

const (
	Unknown VarType = iota
	Int
	String
	Date
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
	Type        VarType
	Whole       *DomainClass
	Other       *DomainClass
	Comment     string
	Name        string
	Column      string
	Default     string
	Range       IntRange
	IsAuto      bool
	IsEditable  bool
	IsUpdatable bool
}

func (v DVar) GoType() string {
	switch {
	case v.Whole != nil || v.Other != nil:
		return "int"
	case v.Type == Int:
		return "int"
	case v.Type == String:
		return "string"
	case v.Type == Date:
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

func joinDVars(arrs ...[]DVar) []DVar {
	retv := []DVar{}
	for _, a := range arrs {
		if a != nil {
			retv = append(retv, a...)
		}
	}
	return retv
}

func joinDVarsCond(p func(DVar) bool, arrs ...[]DVar) []DVar {
	retv := []DVar{}
	for _, a := range arrs {
		if a != nil {
			for _, x := range a {
				if p(x) {
					retv = append(retv, x)
				}
			}
		}
	}
	return retv
}

type DomainClass struct {
	Name       string
	Table      string
	Comment    string
	Arguments  []DVar
	Editables  []DVar
	Updatables []DVar
	External   []DVar
	Autos      []DVar

	allVars       []DVar
	defaultArgs   []DVar
	initvalVars   []DVar
	allAutos      []DVar
	allEditables  []DVar
	allUpdatables []DVar
}

func (d *DomainClass) Init() {
	if d.Arguments == nil {
		d.Arguments = []DVar{}
	}
	if d.Editables == nil {
		d.Editables = []DVar{}
	}
	if d.Updatables == nil {
		d.Updatables = []DVar{}
	}
	if d.External == nil {
		d.External = []DVar{}
	}
	if d.Autos == nil {
		d.Autos = []DVar{}
	}
}

func (d *DomainClass) Args() []DVar {
	return d.Arguments
}

func (d *DomainClass) allVars_() []DVar {
	return joinDVars(d.Arguments, d.Editables, d.Updatables, d.External, d.Autos)
}

func (d *DomainClass) AllVars() []DVar {
	if d.allVars == nil {
		d.allVars = d.allVars_()
	}
	return d.allVars
}

func (d *DomainClass) defaultArgs_() []DVar {
	return joinDVarsCond(func(v DVar) bool {
		b, _ := v.IsDefault()
		return b
	}, d.Arguments)
}

func (d *DomainClass) DefaultArgs() []DVar {
	if d.defaultArgs == nil {
		d.defaultArgs = d.defaultArgs_()
	}
	return d.defaultArgs
}

func (d *DomainClass) initvalVars_() []DVar {
	return joinDVarsCond(func(v DVar) bool {
		b, _ := v.IsDefault()
		return b
	}, d.Autos, d.Editables, d.External)
}

func (d *DomainClass) InitvalVars() []DVar {
	if d.initvalVars == nil {
		d.initvalVars = d.initvalVars_()
	}
	return d.initvalVars
}

func (d *DomainClass) allAutos_() []DVar {
	return joinDVars(d.Autos,
		joinDVarsCond(func(v DVar) bool {
			return v.IsAuto
		}, d.Arguments, d.Editables, d.Updatables))
}

func (d *DomainClass) AllAutos() []DVar {
	if d.allAutos == nil {
		d.allAutos = d.allAutos_()
	}
	return d.allAutos
}

func (d *DomainClass) allEditables_() []DVar {
	return joinDVars(d.Editables,
		joinDVarsCond(func(v DVar) bool {
			return v.IsEditable
		}, d.Arguments, d.Autos))
}

func (d *DomainClass) AllEditables() []DVar {
	if d.allEditables == nil {
		d.allEditables = d.allEditables_()
	}
	return d.allEditables
}

func (d *DomainClass) allUpdatables_() []DVar {
	return joinDVars(d.Updatables,
		joinDVarsCond(func(v DVar) bool {
			return v.IsUpdatable
		}, d.Arguments, d.Autos))
}

func (d *DomainClass) AllUpdatables() []DVar {
	if d.allUpdatables == nil {
		d.allUpdatables = d.allUpdatables_()
	}
	return d.allUpdatables
}

func (d *DomainClass) IsEditable() bool {
	return len(d.AllEditables()) > 0
}

func (d *DomainClass) IsUpdatable() bool {
	return len(d.AllUpdatables()) > 0
}

func (d *DomainClass) IsExternal() bool {
	return len(d.AllVars()) == len(d.External)
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
