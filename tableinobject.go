package object

import (
	"errors"
	"sort"
	"strings"

	"github.com/svcbase/base"
	"github.com/tidwall/gjson"
)

type CodesetPropertyT struct {
	Tablename       string
	Objectcaption   string
	Propertyname    string /*language_id*/
	Propertytype    string /*id,code*/
	Propertycaption string
}

type TableInObject struct {
	Identifier       string
	Tables           []string
	Objects          []string
	Codesets         []string
	CodesetinObjects map[string][]CodesetPropertyT
}

func (tio *TableInObject) ParseO(o *gjson.Result) (e error) {
	if o.Exists() {
		tio.CodesetinObjects = make(map[string][]CodesetPropertyT)
		tio.getTable(o, []string{tio.Identifier})
		sort.Strings(tio.Codesets)
		tio.Tables = append(tio.Tables, tio.Objects...)
		tio.Tables = append(tio.Tables, tio.Codesets...)
	}
	return
}

func (tio *TableInObject) Parse(definition string) (e error) {
	if gjson.Valid(definition) {
		o := gjson.Parse(definition)
		e = tio.ParseO(&o)
	} else {
		e = errors.New("definition error!")
	}
	return
}

func (tio *TableInObject) getTable(o *gjson.Result, roadmap []string) {
	o_type := o.Get("type").String()
	if strings.HasPrefix(o_type, "object") || o_type == "codeset" {
		table := strings.Join(roadmap, "_")
		exists, _ := base.In_array(table, tio.Objects)
		if !exists {
			tio.Objects = append(tio.Objects, table)
		}
		caption := o.Get("caption").String()
		o.ForEach(func(k, v gjson.Result) bool {
			key := k.String()
			if v.Type.String() == "JSON" {
				o_type := v.Get("type").String()
				if strings.HasPrefix(o_type, "object") {
					tio.getTable(&v, append(roadmap, key))
				} else {
					codeset := v.Get("codeset").String()
					if len(codeset) == 0 {
						codeset = v.Get("options").String()
					}
					if len(codeset) > 0 {
						otype := "id"
						if o_type != "int" {
							otype = "code"
						}
						cp := CodesetPropertyT{Tablename: table, Objectcaption: caption, Propertyname: key, Propertytype: otype, Propertycaption: v.Get("caption").String()}

						if cio, ok := tio.CodesetinObjects[codeset]; ok {
							cio = append(cio, cp)
						} else {
							tio.CodesetinObjects[codeset] = []CodesetPropertyT{cp}
						}
						exists, _ := base.In_array(codeset, tio.Codesets)
						if !exists {
							tio.Codesets = append(tio.Codesets, codeset)
						}
					}
				}
			}
			return true
		})
	}
	return
}

func (tio *TableInObject) CodesetRelatedTables(codeset string) (codesetntables []CodesetPropertyT) {
	codesetntables = tio.CodesetinObjects[codeset]
	return
}
