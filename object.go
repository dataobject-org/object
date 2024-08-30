package object

import (
	"base"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

const (
	CRLF = "\r\n"
)

type indexT struct {
	idx_type       string
	idx_properties string
}

type propertyT struct {
	mapKV map[string]string
}

func mapKV_simple(val, readability string) (property propertyT) {
	property.mapKV = make(map[string]string)
	property.mapKV["type"] = "-"
	property.mapKV["default"] = val
	property.mapKV["comment"] = readability
	return
}

func quote(ss string) (txt string) {
	txt = strconv.Quote(ss)
	return
}

func em_quote(emphasis, ss string) (txt string) {
	txt = strconv.Quote(fmt.Sprintf(emphasis, ss))
	return
}

/*
"type": "int",
"index": "single",
"language_adaptive": true
"default": "0",
"comment": "global unique in the data set",
"caption": "en:base language;zh:基础语言"
*/
func mapKV_full(i_type, i_options, i_default, i_comment, i_caption, i_pattern, i_size, i_capacity, i_joinsuperiors string, i_language_adaptive bool) (property propertyT) {
	property.mapKV = make(map[string]string)
	property.mapKV["type"] = quote(i_type)
	if i_type == "string" || i_type == "password" {
		size := base.Str2int(i_size)
		if size > 0 {
			property.mapKV["size"] = i_size
		} else {
			property.mapKV["size"] = base.DEFAULT_STRING_SIZE
		}
	}
	if i_type == "ipv4" {
		size := base.Str2int(i_size)
		if size > 0 {
			property.mapKV["size"] = i_size
		} else {
			property.mapKV["size"] = base.DEFAULT_IPV4_SIZE
		}
	}
	if i_type == "ipv6" {
		size := base.Str2int(i_size)
		if size > 0 {
			property.mapKV["size"] = i_size
		} else {
			property.mapKV["size"] = base.DEFAULT_IPV6_SIZE
		}
	}
	if i_type == "dotids" {
		size := base.Str2int(i_size)
		if size > 0 {
			property.mapKV["size"] = i_size
		} else {
			property.mapKV["size"] = base.DEFAULT_DOTIDS_SIZE
		}
	}
	if i_type == "text" || i_type == "blob" {
		if len(i_capacity) > 0 {
			property.mapKV["capacity"] = i_capacity
		}
	}
	if i_language_adaptive {
		property.mapKV["language_adaptive"] = "true"
	}
	if len(i_options) > 0 {
		property.mapKV["options"] = quote(i_options)
	}
	if len(i_default) > 0 {
		property.mapKV["default"] = quote(i_default)
	} else {
		if i_type == "string" || i_type == "text" {
			property.mapKV["default"] = quote(i_default)
		}
	}
	if len(i_comment) > 0 {
		property.mapKV["comment"] = quote(i_comment)
	}
	if len(i_caption) > 0 {
		property.mapKV["caption"] = quote(i_caption)
	}
	if len(i_pattern) > 0 {
		property.mapKV["pattern"] = quote(i_pattern)
	}
	if len(i_joinsuperiors) > 0 {
		property.mapKV["joinsuperiors"] = quote(i_joinsuperiors)
	}
	return
}

func UserAccess(user_type int, limit gjson.Result) (access string) {
	//fmt.Println("UserAccess:", limit.String())
	//fmt.Println("UserType:", user_type, base.GetUsertype(user_type))
	access = ""
	if limit.IsArray() {
		access = "0=1" //no access authority by default
		for _, row := range limit.Array() {
			utut := strings.Split(row.Get("usertype").String(), ",")
			ac := row.Get("access").String()
			ac = strings.ReplaceAll(ac, "TODAYUNIX", base.TodayUnixtime())
			exists, _ := base.In_array("all", utut)
			if exists {
				access = ac
				break
			} else {
				exists, _ = base.In_array(base.GetUsertype(user_type), utut)
				if exists {
					access = ac
					break
				}
			}
		}
	}
	if access == "*" {
		access = ""
	}
	return access
}

func textProperty(id, EMPHASIS string, mpKV map[string]string) (txt, txtread string) {
	txt = quote(id) + ": "
	txtread = em_quote(EMPHASIS, id) + ": "
	if o_type, ok := mpKV["type"]; ok {
		if o_type == "-" {
			txt += mpKV["default"]
			txtread += mpKV["comment"]
		} else {
			mm := []string{}
			mm = append(mm, quote("type")+": "+o_type)
			if o_type == quote("string") || o_type == quote("password") {
				if o_size, ok := mpKV["size"]; ok && len(o_size) > 0 {
					mm = append(mm, quote("size")+": "+o_size)
				}
			}
			keys := []string{"options", "capacity", "unitofmeasure", "set_exclusive", "decimal_places", "encoding", "labeling", "default", "comment", "caption", "pattern", "index", "language_adaptive", "joinsuperiors"} //====****
			n := len(keys)
			for i := 0; i < n; i++ {
				key := keys[i]
				if ss, ok := mpKV[key]; ok && len(ss) > 0 {
					mm = append(mm, quote(key)+": "+ss)
				}
			}
			txt += "{" + strings.Join(mm, ",") + "}"
			txtread += "{" + strings.Join(mm, ",") + "}"
		}
	}
	return
}

func languageadaptiveObject(roadmap, language_adaptivee, language_adaptiver []string, NEWLINE, OFFSET, TAB, EMPHASIS string) (definition, readability string) {
	definition = "{"
	readability = "{" + NEWLINE
	ss := []string{quote("type") + ": " + quote("object")} //object_extension 2024-03-27
	ss = append(ss, quote("extension")+": "+quote("language"))
	tt := []string{em_quote(EMPHASIS, "type") + ": " + quote("object")} //object_extension
	tt = append(tt, em_quote(EMPHASIS, "extension")+": "+quote("language"))
	d, r := textProperty("id", EMPHASIS, mapKV_full("int", "", "", "", "", "", "", "", "", false).mapKV)
	ss = append(ss, d)
	tt = append(tt, r)
	n := len(roadmap)
	for i := 0; i < n; i++ {
		d, r = textProperty(strings.Join(roadmap[0:i+1], "_")+"_id", EMPHASIS, mapKV_full("int", "", "0", "", "", "", "", "", "", false).mapKV)
		ss = append(ss, d)
		tt = append(tt, r)
	}
	d, r = textProperty("language_id", EMPHASIS, mapKV_full("int", "", "0", "", "", "", "", "", "", false).mapKV)
	ss = append(ss, d)
	tt = append(tt, r)
	d, r = textProperty("language_tag", EMPHASIS, mapKV_full("string", "", "", "", "", "", "", "", "", false).mapKV)
	ss = append(ss, d)
	tt = append(tt, r)
	d, r = textProperty("time_created", EMPHASIS, mapKV_full("time", "", base.ZERO_TIME, "", "", "", "", "", "", false).mapKV)
	ss = append(ss, d)
	tt = append(tt, r)
	d, r = textProperty("time_updated", EMPHASIS, mapKV_full("time", "", base.ZERO_TIME, "", "", "", "", "", "", false).mapKV)
	ss = append(ss, d)
	tt = append(tt, r)
	ss = append(ss, language_adaptivee...)
	tt = append(tt, language_adaptiver...)
	ds, rs := quote("indexes"), em_quote(EMPHASIS, "indexes")
	txt := ": ["
	txt += "{" + quote("name") + ": " + quote("id") + ","
	txt += quote("properties") + ": " + quote("id") + ","
	txt += quote("type") + ": " + quote("primary")
	txt += "},"
	ds += txt
	rs += txt + NEWLINE + OFFSET + TAB + TAB
	for i := 0; i < n-1; i++ { //2023-04-21
		o_id := strings.Join(roadmap[0:i+1], "_") + "_id"
		txt = "{" + quote("name") + ": " + quote(o_id) + ","
		txt += quote("properties") + ": " + quote(o_id) + ","
		txt += quote("type") + ": " + quote("single")
		txt += "},"
		ds += txt
		rs += txt + NEWLINE + OFFSET + TAB + TAB
	}
	identifier := strings.Join(roadmap, "_") //2023-04-21 改变了最后一层的取名逻辑，要求分级recursive_deletion。
	txt = "{" + quote("name") + ": " + quote(identifier+"_id_language") + ","
	txt += quote("properties") + ": " + quote(identifier+"_id,language_id") + ","
	txt += quote("type") + ": " + quote("composite")
	txt += "}]"
	ds += txt
	rs += txt
	ss = append(ss, ds)
	tt = append(tt, rs)

	definition += strings.Join(ss, ",") + "}"
	readability += OFFSET + TAB + strings.Join(tt, ","+NEWLINE+OFFSET+TAB) + NEWLINE
	readability += OFFSET + "}"
	return
}

func extObject(object gjson.Result, roadmap, roadmaptype []string, multi_language bool, NEWLINE, TAB, EMPHASIS string) (definition, readability string, e error) {
	o_type := object.Get("type").String()
	if strings.HasPrefix(o_type, "object") || o_type == "codeset" {
		nHier := len(roadmap)
		identifier := roadmap[nHier-1]
		TABS := strings.Repeat(TAB, nHier)
		self_relationship := object.Get("self_relationship").String()
		extension, relation := "", ""
		switch o_type {
		case "object_extension":
			extension = object.Get("extension").String() //like language
		case "object_relation":
			relation = object.Get("relation").String()
		}
		definition = "{"
		readability = "{" + NEWLINE
		indexes := make(map[string]indexT) //multi-key index
		o_keys, keys := []string{}, []string{}
		properties := make(map[string]propertyT)
		language_adaptivee, language_adaptiver := []string{}, []string{}
		major := "" //the major index key name
		key := "id"
		indexes[key] = indexT{"primary", key}
		properties[key] = mapKV_full("int", "", "", identifier+" instance id", "", "", "", "", "", false)
		keys = append(keys, key)
		key = "time_created"
		properties[key] = mapKV_full("time", "", base.ZERO_TIME, "", "", "", "", "", "", false)
		keys = append(keys, key)
		key = "time_updated"
		properties[key] = mapKV_full("time", "", base.ZERO_TIME, "", "", "", "", "", "", false)
		keys = append(keys, key)
		if nHier > 1 && o_type != "object_extension" { //2023-02-13
			for i := 0; i < nHier-1; i++ {
				key = strings.Join(roadmap[0:i+1], "_") + "_id"
				i_options := ""
				if roadmaptype[i] == "codeset" {
					i_options = strings.Join(roadmap[0:i+1], "_")
				}
				properties[key] = mapKV_full("int", i_options, "0", "", "", "", "", "", "", false)
				keys = append(keys, key)
				indexes[key] = indexT{"single", key}
			}
		}
		/*if nHier > 1 {//2023-02-13
			for i := 0; i < nHier-1; i++ {
				key = strings.Join(roadmap[0:i+1], "_") + "_id"
				properties[key] = mapKV_full("int", "0", "", "", "", "", "", "", false, false, false)
				keys = append(keys, key)
				if i == nHier-2 && len(extension) > 0 {
					indexes[key+"_"+extension] = indexT{"composite", key + "," + extension + "_id"}
				} else {
					indexes[key] = indexT{"single", key}
				}
			}
		}*/
		if len(extension) > 0 { //扩展代码集，可选
			key = extension + "_id"
			properties[key] = mapKV_full("int", "", "0", "", "", "", "", "", "", false)
			keys = append(keys, key)
			key = extension + "_tag"
			properties[key] = mapKV_full("string", "", "", "", "", "", "", "", "", false)
			keys = append(keys, key)
		}
		if len(relation) > 0 {
			key = relation + "_id"
			properties[key] = mapKV_full("int", "", "0", "", "", "", "", "", "", false)
			keys = append(keys, key)
			po := strings.Join(roadmap[0:nHier-1], "_")
			indexes["relation"] = indexT{"composite", po + "_id," + relation + "_id"}
		}

		if o_type == "codeset" {
			key = "code"
			indexes[key] = indexT{"unique", key}
			properties[key] = mapKV_full("string", "", "", "codeset uniform definition, unique identifier(UUID)", "en:code;zh:代码", "^[0-9a-zA-Z_\\-]*$", "64", "", "", false)
			keys = append(keys, key)
			key = "name"
			m := mapKV_full("string", "", "", "", "en:name;zh:名称", "", "", "", "", true)
			properties[key] = m
			keys = append(keys, key)
			d, r := textProperty(key, EMPHASIS, m.mapKV)
			language_adaptivee = append(language_adaptivee, d) //definition
			language_adaptiver = append(language_adaptiver, r) //readability

			key = "description"
			m = mapKV_full("text", "", "", "", "en:description;zh:说明", "", "", "", "", true)
			properties[key] = m
			keys = append(keys, key)
			d, r = textProperty(key, EMPHASIS, m.mapKV)
			language_adaptivee = append(language_adaptivee, d) //definition
			language_adaptiver = append(language_adaptiver, r) //readability

			key = "enableflag"
			indexes[key] = indexT{"single", key}
			properties[key] = mapKV_full("int", "", "1", "code item status[0:disable,1:enable]", "en:enable;zh:是否启用", "^[01]$", "", "", "", false)
			keys = append(keys, key)
			key = "ordinalposition"
			if self_relationship != "hierarchical" {
				indexes[key] = indexT{"single", key}
			}
			properties[key] = mapKV_full("int", "", "0", "show position in all siblings", "en:ordinal position;zh:顺序号", "^[0-9]+$", "", "", "", false)
			keys = append(keys, key)

			default_occurrences := ""
			key = "occurrences"
			properties[key] = mapKV_full("text", "", default_occurrences, "code item usage quantity", "en:occurrences;zh:使用量", "", "", "", "", false)
			keys = append(keys, key)
		}

		if self_relationship == "hierarchical" {
			key = "parentid"
			properties[key] = mapKV_full("int", "", "0", "parent instance id", "", "", "", "", "", false)
			keys = append(keys, key)
			key = "ordinalposition"
			exists, _ := base.In_array(key, keys)
			if !exists {
				properties[key] = mapKV_full("int", "", "0", "show position in all siblings", "en:ordinal position;zh:顺序号", "^[0-9]+$", "", "", "", false)
				keys = append(keys, key)
			}
			indexes["siblingorder"] = indexT{"composite", "parentid,ordinalposition"} //multi-key index
			/*if o_type == "codeset" {
				indexes["siblingorder"] = indexT{"composite", "parentid,ordinalposition"} //multi-key index
			} else {
				indexes[key] = indexT{"single", key}
			}*/
			//if o_type == "codeset" {
			key = "isleaf"
			properties[key] = mapKV_full("int", "", "1", "is hierarchy leaf node", "", "^[01]$", "", "", "", false)
			keys = append(keys, key)
			key = "depth"
			properties[key] = mapKV_full("int", "", "0", "hierarchy depth", "", "^[0-9]+$", "", "", "", false)
			keys = append(keys, key)
			//}
		}
		object.ForEach(func(k, v gjson.Result) bool {
			key = k.String()
			if v.Type.String() == "JSON" {
				if key == "indexes" { //array
					i_array := v.Array()
					for _, vv := range i_array {
						index_name := vv.Get("name").String()
						index_properties := vv.Get("properties").String()
						index_type := vv.Get("type").String()
						if len(index_name) > 0 && len(index_properties) > 0 {
							iv := indexes[index_name]
							if len(index_properties) > 0 {
								iv.idx_properties = index_properties
							}
							if len(index_type) > 0 {
								iv.idx_type = index_type
							}
							indexes[index_name] = iv
						}
					}
				} else {
					v_type := v.Get("type").String()
					if strings.HasPrefix(v_type, "object") {
						def, ra, ee := extObject(v, append(roadmap, key), append(roadmaptype, o_type), multi_language, NEWLINE, TAB, EMPHASIS)
						if ee == nil {
							properties[key] = mapKV_simple(def, ra)
							keys = append(keys, key)
						} else {
							e = ee
							return false
						}
					} else { //merge properties
						var m map[string]string
						if p, ok := properties[key]; ok {
							m = p.mapKV
						} else {
							m = make(map[string]string)
							keys = append(keys, key) //new key
						}
						mm := v.Map()
						for kk, vv := range mm {
							if kk == "index" { //indexes merge
								iv := vv.String()
								//if len(iv) == 0 { //clear exist index
								//	delete(indexes, key)
								//} else {
								namee, keyy := []string{}, []string{}
								//n = len(roadmap)
								if nHier > 1 && o_type != "object_extension" { //2023-02-13
									po := strings.Join(roadmap[0:nHier-1], "_") //parent object
									namee = append(namee, po)
									keyy = append(keyy, po+"_id")
								}
								if iv == "major" {
									namee = append(namee, key)
									keyy = append(keyy, key)
									major = key
								} else if iv == "auxiliary" {
									if len(major) > 0 {
										namee = append(namee, major)
										keyy = append(keyy, major)
									}
									namee = append(namee, key)
									keyy = append(keyy, key)
								} else { //single
									namee = append(namee, key)
									keyy = append(keyy, key)
								}
								n := len(keyy)
								if n > 1 {
									for i := 1; i <= n; i++ {
										nm := strings.Join(namee[0:i], "_")
										if _, ok := indexes[nm]; ok {
											delete(indexes, nm)
										}
									}
									indexes[strings.Join(namee, "_")] = indexT{"composite", strings.Join(keyy, ",")}
								} else {
									indexes[key] = indexT{"single", key}
								}
								//}
							} // else {
							m[kk] = vv.Raw
							//}
						}
						properties[key] = propertyT{m}
						if m["language_adaptive"] == "true" {
							o := make(map[string]string)
							for k, v := range m {
								if k != "index" && k != "language_adaptive" {
									o[k] = v
								}
							}
							d, r := textProperty(key, EMPHASIS, o)
							language_adaptivee = append(language_adaptivee, d) //definition
							language_adaptiver = append(language_adaptiver, r) //readability
						}
					}
				}
			} else {
				properties[key] = mapKV_simple(v.Raw, v.Raw) //simple node. like: "type": "codeset" / "comment": "Accept language codeset"
				o_keys = append(o_keys, key)
			}
			return true
		})
		o_keys = append(o_keys, keys...)
		if multi_language && (len(language_adaptivee) > 0) {
			key = "languages"
			properties[key] = mapKV_simple(languageadaptiveObject(roadmap, language_adaptivee, language_adaptiver, NEWLINE, TABS, TAB, EMPHASIS))
			o_keys = append(o_keys, key)
		}
		for i, kk := range o_keys {
			if i > 0 {
				definition += ","
				readability += "," + NEWLINE
			}
			d, r := textProperty(kk, EMPHASIS, properties[kk].mapKV)
			definition += d
			readability += TABS + r
		}
		if len(indexes) > 0 {
			definition += ","
			readability += "," + NEWLINE
			definition += quote("indexes") + ": ["
			readability += TABS + em_quote(EMPHASIS, "indexes") + ": ["
			nn := 0
			for k, v := range indexes {
				if nn > 0 {
					definition += ","
					readability += "," + NEWLINE + TABS + TAB
				}
				txt := "{" + quote("name") + ": " + quote(k) + ","
				txt += quote("properties") + ": " + quote(v.idx_properties)
				if len(v.idx_type) > 0 {
					txt += "," + quote("type") + ": " + quote(v.idx_type)
				}
				txt += "}"
				definition += txt
				readability += txt
				nn++
			}
			definition += "]"
			readability += "]" + NEWLINE
		}
		definition += "}"
		readability += strings.Repeat(TAB, nHier-1) + "}"
	}
	return
}

func sqlQuote(ss string) (tt string) {
	tt = "'" + ss + "'"
	return
}

type ObjectpropertyT struct {
	object_type    string
	identifier     string
	logo           string
	comment        string
	multi_language string //"0":"1"
	self_hierarchy string //"0":"1"
	public_access  string //"0":"1"
	coding_type    string //"0":"1"
	loading_mode   string //"0":"1"
	pad_format     string
	code_structure string
	authority      string
	caption        string
	definition     string
}

func GetObjectProperty(o gjson.Result, o_identifier string) (op ObjectpropertyT) {
	op.definition = o.String()
	op.object_type = o.Get("type").String()
	op.identifier = o_identifier
	op.logo = o.Get("logo").String()
	op.comment = o.Get("comment").String()
	op.public_access = o.Get("public_access").String()
	if len(op.public_access) == 0 {
		op.public_access = "0"
	}
	op.multi_language = "0" //single / multiple
	if o.Get("language").String() == "multiple" {
		op.multi_language = "1"
	}
	self_relationship := o.Get("self_relationship").String()
	op.self_hierarchy = "0"
	if self_relationship == "hierarchical" {
		op.self_hierarchy = "1"
	}
	op.coding_type = "0"
	if o.Get("coding_type").String() == "hierarchical" {
		op.coding_type = "1"
	}
	op.loading_mode = "1"
	if o.Get("loading_mode").String() == "dynamic" {
		op.loading_mode = "0"
	}
	op.pad_format = o.Get("pad_format").String()
	op.code_structure = o.Get("code_structure").String()
	op.authority = "system"
	if o.Get("authority").String() == "business" {
		op.authority = "business"
	}
	op.caption = base.Language_label(o.Get("caption").String(), base.BaseLanguage_id())
	if len(op.caption) == 0 {
		op.caption = o_identifier
	}
	return
}

func InsertEntitySQL(op ObjectpropertyT, creator string, db_type int) (esql string) {
	esql = "insert into entity(code,creator,self_hierarchy,multiple_language,metatype_id,name,authority,thumbnail,definition,description,time_created,time_updated)"
	esql += " values(" + sqlQuote(op.identifier) + "," + sqlQuote(creator) + "," + op.self_hierarchy + "," + op.multi_language
	esql += ",(select id from metatype where code='" + op.object_type + "')," + sqlQuote(op.caption) + ","
	esql += sqlQuote(op.authority) + ",'" + op.logo + "',"
	switch db_type {
	case base.SQLite:
		esql += base.SQLiteEscape(op.definition) + "," + base.SQLiteEscape(op.comment) + ",current_timestamp,current_timestamp)"
	case base.MySQL:
		esql += base.MySQLEscape(op.definition) + "," + base.MySQLEscape(op.comment) + ",now(),now())"
	}
	return
}

func UpdateEntitySQL(op ObjectpropertyT, db_type int, entity_id int64) (esql string) {
	esql = "update entity set "
	esql += "self_hierarchy=" + op.self_hierarchy + ","
	esql += "multiple_language=" + op.multi_language + ","
	esql += "metatype_id=(select id from metatype where code='" + op.object_type + "'),"
	esql += "name=" + sqlQuote(op.caption) + ","
	esql += "authority=" + sqlQuote(op.authority) + ","
	esql += "thumbnail='" + op.logo + "',"
	switch db_type {
	case base.SQLite:
		esql += "definition=" + base.SQLiteEscape(op.definition) + ","
		esql += "description=" + base.SQLiteEscape(op.comment) + ","
	case base.MySQL:
		esql += "definition=" + base.MySQLEscape(op.definition) + ","
		esql += "description=" + base.MySQLEscape(op.comment) + ","
	}
	esql += "time_updated=" + base.SQL_now()
	esql += " where id=" + strconv.FormatInt(entity_id, 10)
	return
}

func InsertCodesetSQL(op ObjectpropertyT, db_type int) (csql string) {
	if op.object_type == "codeset" && op.coding_type == "1" {
		csql = "insert into entity_codeset(id,coding_type,loading_mode,pad_format,code_structure,public_access,time_created,time_updated) values("
		csql += "(select id from entity where code=" + sqlQuote(op.identifier) + ")," + op.coding_type + "," + op.loading_mode + ",'" + op.pad_format + "',"
		csql += "'" + op.code_structure + "'," + op.public_access + "," + base.SQL_now() + "," + base.SQL_now() + ")"
	}
	return
}

func UpdateCodesetSQL(op ObjectpropertyT, db_type int, id int64) (csql string) {
	if op.object_type == "codeset" && op.coding_type == "1" {
		csql = "update entity_codeset set "
		csql += "coding_type=" + op.coding_type + ","
		csql += "loading_mode=" + op.loading_mode + ","
		csql += "pad_format='" + op.pad_format + "',"
		csql += "code_structure='" + op.code_structure + "',"
		csql += "public_access=" + op.public_access + ","
		csql += "time_updated=" + base.SQL_now()
		csql += " where id=" + strconv.FormatInt(id, 10)
	}
	return
}

func CreateIndexSQL(o gjson.Result, tablename string) (idxes, onfields []string, primary string) {
	o_indexes := o.Get("indexes")
	if o_indexes.Exists() {
		i_array := o_indexes.Array()
		for _, v := range i_array {
			index_name := v.Get("name").String()
			index_properties := v.Get("properties").String()
			index_type := v.Get("type").String()
			if len(index_name+index_properties) > 0 {
				pp := strings.Split(index_properties, ",")
				//split asc desc from property	time_updated desc -> `time_updated` DESC
				props, props_collation := []string{}, []string{}
				n := len(pp)
				for i := 0; i < n; i++ {
					ss := base.TrimBLANK(pp[i])
					if strings.Contains(ss, "`") {
						ss = strings.ReplaceAll(ss, "`", "")
					}
					if strings.Contains(ss, " ") {
						p_s := strings.Split(ss, " ")
						if len(p_s) == 2 {
							prop := "`" + p_s[0] + "`"
							props = append(props, p_s[0])
							if index_type != "primary" {
								sort := strings.ToUpper(p_s[1])
								if sort == "ASC" || sort == "DESC" {
									prop += " " + sort
								}
							}
							props_collation = append(props_collation, prop)
						}
					} else {
						props_collation = append(props_collation, "`"+ss+"`")
						props = append(props, ss)
					}
				}
				index_properties = strings.Join(props_collation, ",")
				if index_type == "primary" {
					primary = index_properties
				} else {
					xx := "CREATE"
					if index_type == "fulltext" {
						xx += " FULLTEXT"
					}
					inm := "`idx_" + tablename + "_" + index_name + "`"
					if len(inm) > 64 {
						inm = "I" + base.StrMD5(inm)
					}
					xx += " INDEX " + inm + " ON `" + tablename + "`(" + index_properties + ")"
					idxes = append(idxes, xx)
					onfields = append(onfields, strings.Join(props, ","))
				}
			}
		}
	}
	return
}

// normal: true - DEFAULT
func property2SQL(objecttype string, v gjson.Result, db_type int, field_name, primary string, normal bool) (ff string) {
	ff = "`" + field_name + "` "
	field_default := v.Get("default").String()
	switch v.Get("type").String() {
	case "time":
		ff += "datetime"
		if len(field_default) > 0 {
			if field_default != "now" {
				ff += " DEFAULT '" + field_default + "'"
			} else {
				if normal {
					if db_type == base.SQLite {
						ff += " DEFAULT current_timestamp"
					}
				}
			}
		} else {
			ff += " DEFAULT '" + base.ZERO_TIME + "'"
		}
	case "string", "password":
		size := base.DEFAULT_STRING_SIZE
		o_size := v.Get("size")
		if o_size.Exists() {
			size = o_size.String()
		}
		ff += "varchar(" + size + ")"
		ff += " DEFAULT '" + field_default + "'"
	case "ipv4":
		size := base.DEFAULT_IPV4_SIZE
		o_size := v.Get("size")
		if o_size.Exists() {
			size = o_size.String()
		}
		ff += "varchar(" + size + ")"
		ff += " DEFAULT '" + field_default + "'"
	case "ipv6":
		size := base.DEFAULT_IPV6_SIZE
		o_size := v.Get("size")
		if o_size.Exists() {
			size = o_size.String()
		}
		ff += "varchar(" + size + ")"
		ff += " DEFAULT '" + field_default + "'"
	case "dotids":
		size := base.DEFAULT_DOTIDS_SIZE
		o_size := v.Get("size")
		if o_size.Exists() {
			size = o_size.String()
		}
		ff += "varchar(" + size + ")"
		ff += " DEFAULT '" + field_default + "'"
	case "int":
		switch db_type {
		case base.SQLite:
			ff += "INTEGER"
		case base.MySQL:
			ff += "int"
		}
		if "`"+field_name+"`" == primary {
			ff += " PRIMARY KEY"
			if objecttype == "object_extension" {
				//id same as origin id,not autoincrement.
			} else {
				if normal {
					switch db_type {
					case base.SQLite:
						ff += " AUTOINCREMENT"
					case base.MySQL:
						ff += " AUTO_INCREMENT"
					}
				}
			}
			ff += " NOT NULL"
		} else {
			if len(field_default) > 0 {
				ff += " DEFAULT '" + field_default + "'"
			} else {
				ff += " DEFAULT '0'"
			}
		}
	case "float":
		switch db_type {
		case base.SQLite:
			ff += "REAL"
		case base.MySQL:
			ff += "FLOAT"
		}
		if len(field_default) > 0 {
			ff += " DEFAULT '" + field_default + "'"
		} else {
			ff += " DEFAULT '0'"
		}
	case "decimal":
		d := "2"
		d_p := v.Get("decimal_places")
		if d_p.Exists() {
			d = d_p.String()
		}
		ff += "decimal(20," + d + ")"
		if len(field_default) > 0 {
			ff += " DEFAULT '" + field_default + "'"
		} else {
			ff += " DEFAULT '0'"
		}
	case "long":
		ff += "bigint"
		if len(field_default) > 0 {
			ff += " DEFAULT '" + field_default + "'"
		} else {
			ff += " DEFAULT '0'"
		}
	case "text":
		capacity := v.Get("capacity").String()
		switch capacity {
		case "L", "long":
			ff += "LONGTEXT"
		case "M", "medium":
			ff += "MEDIUMTEXT"
		default:
			ff += "TEXT"
		}
		/*if len(field_default) > 0 {//mysql can not set default value.
			ff += " DEFAULT '" + field_default + "'"
		} else {
			ff += " DEFAULT ''"
		}*/
	case "blob":
		capacity := v.Get("capacity").String()
		switch capacity {
		case "L", "long":
			ff += "LONGBLOB"
		case "M", "medium":
			ff += "MEDIUMBLOB"
		default:
			ff += "TEXT"
		}
	}
	if normal {
		o_comment := v.Get("comment").String()
		o_pattern := v.Get("pattern").String()
		if len(o_pattern) > 0 {
			o_comment += " " + o_pattern
		}
		if len(o_comment) > 0 {
			switch db_type {
			case base.SQLite:
				ff += " /*" + strings.Trim(base.SQLiteEscape(o_comment), "'") + "*/"
			case base.MySQL:
				ff += " COMMENT " + base.MySQLEscape(o_comment)
			}
		}
	}
	return
}

func CreateTableSQL(o gjson.Result, roadmap []string, primary, creator string, db_type int, NEWLINE, TAB string) (asql string) {
	asql = "CREATE TABLE `" + strings.Join(roadmap, "_") + "`"
	comment := o.Get("comment").String()
	if len(comment) > 0 && db_type == base.SQLite {
		asql += "/*" + strings.Trim(base.SQLiteEscape(comment), "'") + "*/"
	}
	asql += "(" + NEWLINE
	ncols := 0
	objecttype := o.Get("type").String()
	o.ForEach(func(k, v gjson.Result) bool {
		field_name := k.String()
		if v.Type.String() == "JSON" && field_name != "indexes" {
			if !strings.HasPrefix(v.Get("type").String(), "object") { //codeset must not be in second level
				if ncols > 0 {
					asql += "," + NEWLINE
				}
				asql += TAB + property2SQL(objecttype, v, db_type, field_name, primary, true)
				ncols++
			}
		}
		return true
	})
	if strings.Contains(primary, ",") { //if no 'id' property, last line end ","  must handle
		asql += TAB + " PRIMARY KEY(" + primary + ")" + NEWLINE
	}
	asql += ")"
	if len(comment) > 0 && db_type == base.MySQL {
		asql += "COMMENT=" + base.MySQLEscape(comment)
	}
	/*charset*/
	switch db_type {
	case base.SQLite:
	case base.MySQL:
		asql += " DEFAULT CHARSET=utf8"
	}
	return
}

func UpdateTableSQL(o gjson.Result, roadmap []string, primary string, db_type int) (sqlsql []string) {
	tablename := strings.Join(roadmap, "_")
	//fmt.Println("UpdateTableSQL:", tablename)
	var ti base.TableInfoT
	if ti.ReadFields(tablename) == nil {
		objecttype := o.Get("type").String()
		o.ForEach(func(k, v gjson.Result) bool {
			field_name := k.String()
			if v.Type.String() == "JSON" && field_name != "indexes" {
				if !strings.HasPrefix(v.Get("type").String(), "object") { //codeset must not be in second level
					normal_propertySQL := property2SQL(objecttype, v, db_type, field_name, primary, true)
					asql := ""
					if ti.FieldExists(field_name) {
						short_propertySQL := property2SQL(objecttype, v, db_type, field_name, primary, false)
						if !ti.SameProperty(field_name, short_propertySQL) {
							switch db_type {
							case base.SQLite: /*not support MODIFY COLUMN*/
							case base.MySQL:
								asql += " modify " + normal_propertySQL
							}
						}
					} else {
						switch db_type {
						case base.SQLite:
							asql += " ADD COLUMN " + normal_propertySQL
						case base.MySQL:
							asql += " add " + normal_propertySQL
						}
					}
					if len(asql) > 0 {
						sqlsql = append(sqlsql, "alter table `"+tablename+"`"+asql)
					}
				}
			}
			return true
		})
	}
	return
}

func def2SQL(o gjson.Result, roadmap []string, creator string, db_type int, NEWLINE, TAB string) (ss, es []string) {
	o_type := o.Get("type").String()
	if strings.HasPrefix(o_type, "object") || o_type == "codeset" {
		op := GetObjectProperty(o, strings.Join(roadmap, "."))
		asql := InsertEntitySQL(op, creator, db_type)
		es = append(es, asql+";")
		asql = InsertCodesetSQL(op, db_type)
		if len(asql) > 0 {
			es = append(es, asql+";")
		}
		o.ForEach(func(k, v gjson.Result) bool {
			field_name := k.String()
			if v.Type.String() == "JSON" && field_name != "indexes" {
				o_type := v.Get("type").String()
				if strings.HasPrefix(o_type, "o") { //codeset must not be in second level
					vv, _ := def2SQL(v, append(roadmap, field_name), creator, db_type, NEWLINE, TAB) //sub-object recursive call
					ss = append(ss, vv...)
				}
			}
			return true
		})
		idxes, _, primary := CreateIndexSQL(o, strings.Join(roadmap, "_"))
		asql = CreateTableSQL(o, roadmap, primary, creator, db_type, NEWLINE, TAB)
		if len(asql) > 0 {
			if len(idxes) > 0 {
				for _, x := range idxes {
					ss = append(ss, x+";") //ss = append(ss, idxes...)
				}
			}
			ss = append(ss, asql+";")
		}
	}
	if len(roadmap) == 1 {
		for i, j := 0, len(ss)-1; i < j; i, j = i+1, j-1 { //array reverse
			ss[i], ss[j] = ss[j], ss[i]
		}
	}
	return
}

func Extend(dirRes, identifier, NEWLINE, TAB, EMPHASIS string, sys_accept_multi_language bool) (definition, readability string, e error) { //format: purifying / readability
	fname := filepath.Join(dirRes, "res", identifier+".object")
	if base.IsExists(fname) {
		bb, err := ioutil.ReadFile(fname)
		if err == nil {
			definition, readability, e = DefinitionExtend(bb, identifier, NEWLINE, TAB, EMPHASIS, sys_accept_multi_language)
		}
	} else {
		e = errors.New("error object identifier")
	}
	return
}

func DefinitionExtend(jsontxt []byte, identifier, NEWLINE, TAB, EMPHASIS string, sys_accept_multi_language bool) (definition, readability string, e error) {
	m_l := false
	if identifier != "language" { //need confirm
		m_l = sys_accept_multi_language
	}
	result := gjson.GetBytes(jsontxt, identifier)
	if result.Exists() {
		if m_l && result.Get("language").String() == "single" { //object close multiple language
			m_l = false
		}
		definition, readability, e = extObject(result, []string{identifier}, []string{result.Get("type").String()}, m_l, NEWLINE, TAB, EMPHASIS)
		if e == nil {
			definition = "{" + quote(identifier) + ": " + definition + "}"
			readability = "{" + em_quote(EMPHASIS, identifier) + ": " + readability + "}"
		}
	}
	return
}

func DefinitionTextExtend(jsontxt, identifier string, multi_language bool) (definition string, e error) {
	m_l := false
	if identifier != "language" {
		m_l = multi_language
	}
	result := gjson.Get(jsontxt, identifier)
	if result.Exists() {
		definition, _, e = extObject(result, []string{identifier}, []string{result.Get("type").String()}, m_l, "", "", "")
	}
	return
}

func Definition2SQL(definition, identifier, creator string, db_type int, NEWLINE, TAB string) (sqlsql, entitysql []string, e error) { //NEWLINE: "<br>" TAB: strings.Repeat("&nbsp;", 8)
	result := gjson.Get(definition, identifier)
	if result.Exists() {
		sqlsql, entitysql = def2SQL(result, []string{identifier}, creator, db_type, NEWLINE, TAB)
	} else {
		e = errors.New(identifier + " syntax error!")
	}
	return
}

func DefinitionExtend2SQL(definition, identifier string, multi_language bool) (sqlsql []string, definitionex string, e error) {
	m_l := false
	if identifier != "language" {
		m_l = multi_language
	}
	o := gjson.Get(definition, identifier)
	if o.Exists() {
		definitionex, _, e = extObject(o, []string{identifier}, []string{o.Get("type").String()}, m_l, "", "", "")
		if e == nil {
			o = gjson.Parse(definitionex)
			if o.Exists() {
				sqlsql, _ = def2SQL(o, []string{identifier}, "", base.DB_type, "", "")
			}
		}
	}
	return
}

type optionT struct {
	Value string
	Text  string
}
type actionT struct {
	Icon      string //"icon": "fa-trash-o",
	Action    string //"action": "remove",
	Scene     string
	Caption   string //"caption": "en:remove;zh:删除",
	Hint      string //"hint": "en:remove;zh:删除", hint for icon
	Condition string //"condition": "released=0&user_id=user.id"
}
type shortcutT struct {
	Icon    string `json:"icon"`
	Action  string `json:"action"`
	Caption string `json:"label"`
}
type hotkeyT struct {
	Icon    string //"icon": "fa-clip-o",
	Action  string //"action": "clip",
	Caption string //"caption": "en:clip;zh:引用",
	Style   string
}
type ColumnT struct {
	Name       string
	Property   string
	Query      string
	Caption    string
	Hyperlink  string
	Mark       string
	Onclick    string
	Order      string
	Format     string
	Render     string
	Align      string
	Width      string
	Transform  string
	Comparison string /*comparison with another property*/
	//	Alllanguage bool
	Visible bool
	Fixed   string
	Options []optionT
	Actions []actionT
	//SelectedOption string
}

func ReplaceChooseActions(col *ColumnT, clientlanguage_id string) {
	caption := base.GetConfigurationLanguage("TIP_CHOOSE", clientlanguage_id)
	col.Actions = append([]actionT{}, actionT{"fa-crosshairs", "choose", "", caption, "", ""})
}

func rightProperty(ss string) (flag bool) {
	flag = true
	if base.IsDigital(ss) {
		flag = false
	} else if strings.Contains(ss, "'") {
		flag = false
	} else if strings.Contains(ss, "@") {
		flag = false
	}
	return
}

/*released=0&user_id=session@user_id&recommend.ordinalposition=0*/
func extractProperty(str string) (properties []string) {
	opers := strings.Split("!=,<=,>=,<,>,=,∈", ",")
	ss := strings.Split(str, "&")
	n := len(ss)
	for i := 0; i < n; i++ {
		expression := ss[i]
		m := len(opers)
		for j := 0; j < m; j++ {
			oper := opers[j]
			if strings.Contains(expression, oper) {
				ee := strings.SplitN(expression, oper, 2)
				if len(ee) == 2 {
					if rightProperty(ee[0]) {
						exists, _ := base.In_array(ee[0], properties)
						if !exists {
							properties = append(properties, ee[0])
						}
					}
					if rightProperty(ee[1]) {
						exists, _ := base.In_array(ee[1], properties)
						if !exists {
							properties = append(properties, ee[1])
						}
					}
				}
				break
			}
		}
	}
	return
}

func isReservedWord(word string) (flag bool) {
	flag = strings.HasPrefix(word, "_") && strings.HasSuffix(word, "_")
	return
}

func ParseList(l_definition, o_definition *gjson.Result) (names, types, properties, transforms []string, txtsort string) {
	if l_definition.Get("type").String() == "list" {
		txtsort = l_definition.Get("sort").String()
		l_definition.ForEach(func(k, v gjson.Result) bool {
			name := k.String()
			if !isReservedWord(name) {
				if v.Type.String() == "JSON" {
					property := v.Get("property").String()
					if len(name) > 0 && len(property) > 0 {
						oproperty := o_definition.Get(property)
						if oproperty.Exists() {
							names = append(names, name)
							types = append(types, oproperty.Get("type").String())
							properties = append(properties, property)
							transforms = append(transforms, v.Get("transform").String())
						}
					}
				}
			}
			return true // keep iterating
		})
	}
	return
}

func ParseTable(u_definition, o_definition *gjson.Result) (names, captions, types, properties, transforms, aligns, widths, renders []string, txtsort string) {
	utype := u_definition.Get("type").String()
	if utype == "grid" {
		txtsort = u_definition.Get("sort").String()
		u_definition.ForEach(func(k, v gjson.Result) bool {
			name := k.String()
			if !isReservedWord(name) {
				if v.Type.String() == "JSON" {
					property := v.Get("property").String()
					if len(name) > 0 && len(property) > 0 {
						exactproperty := exactProperty(property)
						oproperty := o_definition.Get(exactproperty) //user_id^user.name
						if oproperty.Exists() {
							properties = append(properties, property)
							caption := v.Get("caption").String()
							if len(caption) == 0 {
								caption = oproperty.Get("caption").String()
								if len(caption) == 0 {
									caption = name
								}
							}
							names = append(names, name)
							captions = append(captions, caption)
							otype := "string"
							if property == exactproperty {
								o := oproperty.Get("type")
								if o.Exists() {
									otype = oproperty.Get("type").String()
								}
							}
							types = append(types, otype)
							align := v.Get("text-align").String()
							if len(align) == 0 {
								switch otype {
								case "string":
									align = "left"
								case "float", "decimal":
									align = "right"
								default:
									align = "center"
								}
							}
							aligns = append(aligns, align)
							widths = append(widths, v.Get("width").String())
							renders = append(renders, v.Get("render").String())
							transforms = append(transforms, v.Get("transform").String())
						}
					}
				}
			}
			return true // keep iterating
		})
	}
	return
}

func shortcut_JS(bhotkey bool, prefix, func_name, scene, caption string) (jstxt string, dependency []string) {
	switch func_name {
	case "intersect":
		dependency = base.GetWidgetDependency("popis")
		jstxt = `function intersect(){`
		jstxt += `	$.getJSON('/readintersectblock',{eid:{{.entity_id}},scene:'` + scene + `'},function(m){`
		jstxt += `		if(m.Code=="100"){`
		jstxt += `			$('body').Popintersect({`
		jstxt += `				i18n: page_i18n,`
		jstxt += `				identifier: m.Identifier,`
		jstxt += `				entity_id: m.Entity_id,`
		jstxt += `				subentity: m.Subentity,`
		jstxt += `				widget_dependency_bs64: m.Dependency_bs64,`
		jstxt += `				caption: m.Caption,`
		jstxt += `				vert_codeset: m.Vert_codeset,`
		jstxt += `				hori_codeset: m.Hori_codeset,`
		jstxt += `				vertcodeset_bs64: m.Vertcodeset_bs64,`
		jstxt += `				horicodeset_bs64: m.Horicodeset_bs64,`
		jstxt += `				data_bs64: m.Data_bs64`
		jstxt += `			});`
		jstxt += `		}else{alert(m.Msg);}`
		jstxt += `	});`
		jstxt += `}`
	case "onatom":
		dependency = base.GetWidgetDependency("popgrid")
		jstxt = `function onatom(){`
		jstxt += `	var pg_identifier='entity',pg_scene='atom',pg_roadmap='{{.entity_id}}';`
		jstxt += `	$.getJSON('/readgridblock',{idf:pg_identifier,scene:pg_scene},function(m){`
		jstxt += `		if(m.Code=="100"){`
		jstxt += `			$('body').Popgrid({`
		jstxt += `				i18n: page_i18n,`
		jstxt += `				identifier: pg_identifier,`
		jstxt += ` 				caption: '{{.entity_caption}}<i class="fa fa-adn"></i>'+m.Caption,`
		jstxt += ` 				entity_id: m.Entity_id,`
		jstxt += ` 				subentity: m.Subentity,`
		jstxt += ` 				scene: pg_scene,`
		jstxt += ` 				user_id: m.User_id,`
		jstxt += ` 				ip: m.Ip,`
		jstxt += ` 				roadmapids: pg_roadmap,`
		jstxt += ` 				filter_block_bs64: m.Filter_block_bs64,`
		jstxt += ` 				filter_inputs_bs64: m.Filter_inputs_bs64,`
		jstxt += ` 				grid_dependency_bs64: m.Grid_dependency_bs64,`
		jstxt += ` 				rows_per_page: m.Rows_per_page,`
		jstxt += ` 				column_block_bs64: m.Column_block_bs64,`
		jstxt += ` 				column_template_bs64: m.Column_template_bs64,`
		jstxt += ` 				view_scene: m.View_scene,`
		jstxt += ` 				view_dependency_bs64: m.View_dependency_bs64,`
		jstxt += ` 				form_scene: m.Form_scene,`
		jstxt += ` 				form_dependency_bs64: m.Form_dependency_bs64,`
		jstxt += ` 				onClose: function(ismodified){}`
		jstxt += ` 			});`
		jstxt += ` 		}else{alert(m.Msg);}`
		jstxt += `	  });`
		jstxt += `}`
	case "table2csv":
		jstxt = `function table2csv(){`
		jstxt += `	$.getJSON('/table2csv',{idf:'{{.identifier}}',sub:'{{.subentity}}'},function(m){`
		jstxt += `		if(m.Code=="100"){alert('success:'+m.Filename);}else{alert(m.Msg);}`
		jstxt += `	});`
		jstxt += `}`
	case "routegather":
		jstxt = `function routegather(){`
		jstxt += `	$.getJSON('/routegather',{},function(m){`
		jstxt += `		if(m.Code=="100"){alert('success! news:'+m.NewItems);gridRefresh(grid_name);}else{alert(m.Msg);}`
		jstxt += `	});`
		jstxt += `}`
	case "empty":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `function empty(){`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/instanceoperate',{eid:entity_id,ids:ids.join(','),sub:subentity,act:'remove'},function(m){`
		jstxt += `				if(m.Code=="100"){afterEmpty();}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_emptyornot}}'+'?','empty');`
		jstxt += `}`
	case "emptyentity":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `function emptyentity(){`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/instanceoperate',{eid:entity_id,sub:subentity,act:'empty'},function(m){`
		jstxt += `				if(m.Code=="100"){afterEmpty();}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_emptyornot}}'+'?','empty');`
		jstxt += `}`
	case "addnew":
		jstxt = `function addnew(){`
		jstxt += `	var rmi=$('#main_rmi').val();`
		jstxt += `	var url='/form?eid={{.entity_id}}&rmi='+rmi`
		if len(scene) > 0 {
			jstxt += `+'&scene=` + scene + `'`
		}
		jstxt += `;`
		jstxt += `	window.open(url,'_blank');`
		jstxt += `}`
	case "popnew":
		dependency = base.GetWidgetDependency("popform")
		jstxt = `function popnew(){`
		jstxt += `	$('body').Popform({`
		jstxt += `		eid:'{{.entity_id}}',rmi:$('#main_rmi').val(),`
		jstxt += `		scene:'` + scene + `',i18n:page_i18n,`
		jstxt += `		afterSave:function(instance_id){gridRefresh(grid_name);}`
		jstxt += `	}).setInstance(0);`
		jstxt += `}`
	case "popdoc":
		dependency = base.GetWidgetDependency("popdoc")
		jstxt = `function popdoc(){`
		jstxt += `	$('body').Popdoc({`
		jstxt += `		scene:'` + scene + `',caption:'` + caption + `',i18n:page_i18n`
		jstxt += `	});`
		jstxt += `}`
	case "clip":
		jstxt = `function clip(){`
		jstxt += `	window.open('/form?eid={{.entity_id}}&scene=clip','_blank');`
		jstxt += `}`
		jstxt += `$('#` + prefix + `_clip_btn').live('click',function(){`
		jstxt += `	clip();`
		jstxt += `});`
	case "pick":
		dependency = base.GetWidgetDependency("urlpicker")
		jstxt = `function pick(){`
		jstxt += `	$('body').URLpicker({`
		jstxt += `		txt_caption:'{{.t_urlinput}}',txt_run:'{{.t_run}}',`
		jstxt += `		txt_stop:'{{.t_stop}}',txt_stopped:'{{.t_stopped}}',`
		jstxt += `		txt_restart:'{{.t_restart}}',txt_close:'{{.t_close}}',`
		jstxt += `		txt_over:'{{.t_over}}',`
		jstxt += `		onOver: function(){gridRefresh(grid_name);}`
		jstxt += `	});`
		jstxt += `}`
	case "paste":
		dependency = base.GetWidgetDependency("textimport")
		jstxt = `function paste(){`
		jstxt += `	$('body').TextImport({`
		jstxt += `		onCancel: function(id){},`
		jstxt += `		onRevise: function(id,textareaid){`
		jstxt += `			var ta=$('#'+textareaid);`
		jstxt += `			var dt={'type':'GBT9704','data':$.base64.encode(ta.val())};`
		jstxt += `			$.getJSON('/hierarchyrevise',dt,function(m){`
		jstxt += `				if(m.Code=='100'){ta.val($.base64.decode(m.Data));}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `		},`
		jstxt += `		onImport: function(id,fmt,val,rmvspace){`
		jstxt += `			switch(fmt){`
		jstxt += `			case 'GBT9704':`
		jstxt += `				var dt={'eid':{{.entity_id}},'fmt':fmt,'dat':$.base64.encode(val)};`
		jstxt += `				$.getJSON('/pastearticle',dt,function(m){`
		jstxt += `					if(m.Code=='100'){`
		jstxt += `				        window.location.href='/view?eid={{.entity_id}}&iid='+m.Instance_id;`
		jstxt += `					}else{alert(m.Msg);}`
		jstxt += `				});`
		jstxt += `				break;`
		jstxt += `			}`
		jstxt += `		}`
		jstxt += `	}).show_pane('');`
		jstxt += `}`
	case "recommendorder":
		dependency = base.GetWidgetDependency("popordinal")
		jstxt = `function recommendorder(){`
		jstxt += `	recommendremoved=false;`
		jstxt += `	$('body').Popordinal({`
		jstxt += `		txt_caption:'{{.t_recommend}}',`
		jstxt += `		txt_close:'{{.t_close}}',`
		jstxt += `		support_moverow:true,`
		jstxt += `		column_data:[{{.recommendcolumns}}],`
		jstxt += `		entity_id:{{.entity_id}},`
		jstxt += `		entity_extension:'recommend',`
		jstxt += `		onClose: function(){`
		jstxt += `			if(recommendremoved){`
		jstxt += `				gridRefresh(grid_name);`
		jstxt += `				recommendremoved=false;`
		jstxt += `			}`
		jstxt += `		}`
		jstxt += `	});`
		jstxt += `}`
	}
	if bhotkey {
		jstxt += `$('#` + prefix + `_` + func_name + `_btn').live('click',function(){`
		jstxt += `	` + func_name + `();`
		jstxt += `});`
	}
	return
}

func action_JS(prefix, action_name, scene, hint string) (jstxt string, dependency []string) {
	switch action_name {
	case "reply":
		jstxt = `case 'reply':`
		jstxt += `window.open('/reply?id='+instance_id,"_blank");`
		jstxt += `break;`
	case "objectdefinition":
		jstxt = `case 'objectdefinition':`
		jstxt += `window.open('/objectdefinition?fn='+identifier+'.object',"_blank");`
		jstxt += `break;`
	case "labeling":
		jstxt = `case 'labeling':`
		jstxt += `alert('labeling eid:'+instance_id);`
		jstxt += `break;`
	case "select":
		dependency = base.GetWidgetDependency("popselector")
	case "sendmail":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'sendmail':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/sendcachemail',{id:instance_id},function(m){`
		jstxt += `				if(m.Code=="100"){`
		jstxt += `					refreshrow(id,m.Result);`
		jstxt += `					alert('{{.txt_sendmailok}}');`
		jstxt += `				}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `}).show_alertpane('','{{.txt_sendmailagain}}['+instance_name+']?',action);`
		jstxt += `break;`
	case "remove":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'remove':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/instanceoperate',{eid:entity_id,iid:instance_id,sub:subentity,act:action},function(m){`
		jstxt += `				if(m.Code=="100"){`
		jstxt += `					afterRemove(instance_id);gridRefresh(grid_name);`
		jstxt += `				}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `}).show_alertpane('','{{.txt_removeornot}}['+instance_name+']?',action);`
		jstxt += `break;`
	case "optiongrid": //scene: "[templet]" templet is a property name
		dependency = base.GetWidgetDependency("popgrid")
		jstxt = `case 'optiongrid':`
		jstxt += `	$('body').Popgrid({`
		jstxt += `		width:900,height:460,identifier:identifier,roadmapids:instance_id,`
		jstxt += `		scene:templet,caption:caption,i18n:page_i18n`
		jstxt += `	});`
		jstxt += `	break;`
	case "popgrid":
		dependency = base.GetWidgetDependency("popgrid")
		jstxt = `case 'popgrid_` + scene + `':`
		jstxt += `	$('body').Popgrid({`
		jstxt += `		width:900,height:460,identifier:identifier,roadmapids:instance_id,`
		jstxt += `		scene:'` + scene + `',caption:'` + hint + `',i18n:page_i18n`
		jstxt += `	});`
		jstxt += `	break;`
	case "view":
		jstxt = `case 'view':`
		jstxt += `	var url="/view?eid="+entity_id+"&iid="+instance_id`
		if len(scene) > 0 {
			jstxt += `+"&scene=` + scene + `"`
		}
		jstxt += `;	window.open(url,'_blank');`
		jstxt += `	break;`
	case "popview":
		dependency = base.GetWidgetDependency("popview")
		jstxt = `case 'popview':`
		jstxt += `	$('body').Popview({`
		jstxt += `		width:800,height:480,eid:entity_id,iid:instance_id,iids:rowids,`
		jstxt += `		scene:'` + scene + `',`
		jstxt += `		firstText:'{{.t_first}}',prevText:'{{.t_prev}}',nextText:'{{.t_next}}',`
		jstxt += `		lastText:'{{.t_last}}',closeText:'{{.t_close}}',copiedText:'{{.t_copied}}'`
		jstxt += `	}).show();`
		jstxt += `	break;`
	case "popchart":
		dependency = base.GetWidgetDependency("popchart")
		jstxt = `case 'popchart':`
		jstxt += `	$('body').Popchart({`
		jstxt += `		eid:entity_id,sub:subentity,iid:instance_id,`
		jstxt += `		scene:'` + scene + `',`
		jstxt += `		closeText:'{{.t_close}}'`
		jstxt += `	}).show();`
		jstxt += `	break;`
	case "form":
		jstxt = `case 'form':`
		//jstxt += `var rmi=$('#` + prefix + `_rmi').val();`
		jstxt += `var url="/form?eid="+entity_id+"&iid="+instance_id`
		//jstxt += `&rmi='+rmi`
		if len(scene) > 0 {
			jstxt += `+"&scene=` + scene + `"`
		}
		jstxt += `;`
		jstxt += `if(typeof(a_p)!='undefined'){url+='&'+a_p;}`
		jstxt += `window.open(url,'_blank');`
		jstxt += `break;`
	case "popform": //support multiple popforms by popform_scene
		dependency = base.GetWidgetDependency("popform")
		jstxt = `case 'popform_` + scene + `':`
		jstxt += `	$('body').Popform({i18n:page_i18n,`
		jstxt += `		eid:entity_id,rmi:instance_rmi,`
		jstxt += `		instance_name:instance_name,`
		jstxt += `		scene:'` + scene + `',`
		jstxt += `		afterSave:function(instance_id){gridRefresh(grid_name);}`
		jstxt += `	}).setInstance(instance_id);`
		jstxt += `	break;`
	case "popitem":
		dependency = base.GetWidgetDependency("popitem")
		jstxt = `case 'popitem':`
		jstxt += `	$('body').Popitem({eid:instance_id,language:language_id,i18n:page_i18n});`
		jstxt += `	break;`
	case "occurrence":
		jstxt = `case 'occurrence':`
		jstxt += `	window.location.href="/codeoccurrences?eid="+instance_id;`
		jstxt += `	break;`
	case "recommend":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'recommend':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			var data={id:instance_id,ordinalposition:'autoincrement'};`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/saveinstance',{eid:entity_id,sub:'recommend',iid:instance_id,dat:$.base64.encode(JSON.stringify(data))},`
		jstxt += `				function(m){`
		jstxt += `					if(m.Code=='100'){`
		jstxt += `						$.getJSON('/getgridrow',{wgt:'GM',eid:entity_id,iid:instance_id,scene:'{{.scene}}'},function(m){`
		jstxt += `							if(m.Code=='100'){`
		jstxt += `								GridManager.updateRowData(grid_name,'id',JSON.parse(m.Instance_stringify));`
		jstxt += `							}else{alert(m.Msg);}`
		jstxt += `						});`
		jstxt += `					}else{alert(m.Msg);}`
		jstxt += `				}`
		jstxt += `			);`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_recommendornot}}'+'?',action);`
		jstxt += `	break;`
	case "unrecommend":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'unrecommend':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/rmvinstance',{eid:entity_id,sub:'recommend',rmi:instance_id},`
		jstxt += `				function(m){`
		jstxt += `					if(m.Code=='100'){`
		jstxt += `						$.getJSON('/getgridrow',{wgt:'GM',eid:entity_id,iid:instance_id,scene:'{{.scene}}'},function(m){`
		jstxt += `							if(m.Code=='100'){`
		jstxt += `								GridManager.updateRowData(grid_name,'id',JSON.parse(m.Instance_stringify));`
		jstxt += `							}else{alert(m.Msg);}`
		jstxt += `						});`
		jstxt += `					}else{alert(m.Msg);}`
		jstxt += `				}`
		jstxt += `			);`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += ` }).show_alertpane('','{{.txt_unrecommendornot}}'+'?',action);`
		jstxt += ` break;`
	case "issue":
		dependency = base.GetWidgetDependency("dateinput")
		jstxt = `case 'issue':`
		jstxt += `	$('body').DateInputDialog({`
		jstxt += `		okText:'{{.t_ok}}',cancelText:'{{.t_cancel}}',language_code:'{{.clientlanguage_code}}',`
		jstxt += `		hintText:'{{.t_selectadate}}',`
		jstxt += `		doOK:function(newdate,extra){`
		jstxt += `			return saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{released:1,time_released:newdate});`
		jstxt += `		}`
		jstxt += `	}).showinput('','extra');`
		jstxt += `	break;`
	case "revoke":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'revoke':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{released:0,time_released:'0000-01-01 00:00:00'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_revokeornot}}'+'?',action);`
		jstxt += `	break;`
	case "recommendremove":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'recommendremove':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		//jstxt += `			alert('eid:'+entity_id+' iid:'+instance_id);`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/instanceoperate',{eid:entity_id,iid:instance_id,act:'remove',sub:'recommend'},function(m){`
		//jstxt += `				alert('/instanceoperate:'+m.Code);`
		jstxt += `				if(m.Code=="100"){`
		jstxt += `					gridRefresh('popupgrid');` //popordinal.js:60
		jstxt += `					recommendremoved=true;`
		jstxt += `				}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_unrecommendornot}}'+'?',action);`
		jstxt += `	break;`
	case "export":
		jstxt = `case 'export':`
		jstxt += `	$.ajaxSettings.async=false;`
		jstxt += `	$.getJSON('/instanceoperate',{eid:entity_id,iid:instance_id,act:action},function(m){`
		jstxt += `		if(m.Code=="100"){`
		jstxt += `			if(m.Url.length>0){window.open(m.Url,"_blank");}`
		jstxt += `		}else{alert(m.Msg);}`
		jstxt += `	});`
		jstxt += `	$.ajaxSettings.async=true;`
		jstxt += `	break;`
	case "issuelicense":
		dependency = base.GetWidgetDependency("actionform")
		jstxt = `case 'issuelicense':`
		jstxt += `	$('body').Actionform({i18n:page_i18n,`
		jstxt += `		eid:entity_id,scene:'` + scene + `',`
		jstxt += `		onOK:function(data){var o=JSON.parse(data);`
		jstxt += `			var dt={iid:instance_id,level:o['level'],days:o['days'],start:o['start']};`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/issuelicense',dt,function(m){`
		jstxt += `				if(m.Code=='100'){`
		jstxt += `					var rdata={id:instance_id,level:m.Level,start:m.Start,expire:m.Expire};`
		jstxt += `					GridManager.updateRowData(grid_name,'id',rdata);`
		jstxt += `					refreshPartGrid();`
		jstxt += `				}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}`
		jstxt += `	});`
		jstxt += `	break;`
	case "revokelicense":
		dependency = base.GetWidgetDependency("spinnerinput")
		jstxt = `case 'revokelicense':`
		jstxt += `	var dd=Date.parse(o['expire'])-Date.now();`
		jstxt += `	if(dd>0){`
		jstxt += `		var days=Math.round(dd/(24*60*60*1000));`
		jstxt += `		$('body').SpinnerInputDialog({doOK: function(val,extra){`
		jstxt += `			var dt={iid:instance_id,level:o['level'],days:val};`
		jstxt += `			$.ajaxSettings.async = false;`
		jstxt += `			$.getJSON('/revokelicense',dt,function(m){`
		jstxt += `				if(m.Code=='100'){`
		jstxt += `					var rdata={id:instance_id,expire:m.Expire};`
		jstxt += `					GridManager.updateRowData(grid_name,'id',rdata);`
		jstxt += `					refreshPartGrid();`
		jstxt += `				}else{alert(m.Msg);}`
		jstxt += `			});`
		jstxt += `			$.ajaxSettings.async = true;`
		jstxt += `		}}).showinput(days,'',1,days,1);`
		jstxt += `	}`
		jstxt += `	break;`
	case "enable":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'enable':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{status:1,time_switched:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_enableornot}}'+'?',action);`
		jstxt += `	break;`
	case "disable":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'disable':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{status:0,time_switched:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_disableornot}}'+'?',action);`
		jstxt += `	break;`
	case "enableword":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'enableword':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{id:instance_id,stopflag:0,time_flagswitched:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_enablewordornot}}'+'?',action);`
		jstxt += `	break;`
	case "stopword":
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'stopword':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{id:instance_id,stopflag:1,time_flagswitched:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_stopwordornot}}'+'?',action);`
		jstxt += `	break;`
	case "release": //ji application release
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'release':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{id:instance_id,status:1,time_released:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_releaseornot}}'+'?',action);`
		jstxt += `	break;`
	case "cancelrelease": //ji application cancel release
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'cancelrelease':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{id:instance_id,status:0,time_released:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_cancelreleaseornot}}'+'?',action);`
		jstxt += `	break;`
	case "review": //ji entrance review
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'review':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{id:instance_id,status:1,time_reviewed:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_reviewornot}}'+'?',action);`
		jstxt += `	break;`
	case "cancelreview": //ji entrance cancel review
		dependency = base.GetWidgetDependency("yesno")
		jstxt = `case 'cancelreview':`
		jstxt += `	$('body').YesnoAlert({`
		jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
		jstxt += `		doyes: function(id,action){`
		jstxt += `			saveandrefreshrow(grid_name,entity_id,subentity,instance_id,{id:instance_id,status:0,time_reviewed:'NOW'});`
		jstxt += `		}`
		jstxt += `	}).show_alertpane('','{{.txt_cancelreviewornot}}'+'?',action);`
		jstxt += `	break;`
	case "appmould": //ji application press mould
		jstxt = `case 'appmould':`
		jstxt += `	var dt={iid:instance_id};`
		jstxt += `	$.ajaxSettings.async = false;`
		jstxt += `	$.getJSON('/pressmould',dt,function(m){`
		jstxt += `		alert(m.Code+':'+m.Msg);`
		jstxt += `	});`
		jstxt += `	$.ajaxSettings.async = true;`
		jstxt += `	break;`
	}
	return
}

func ParseComponentAction(component *gjson.Result, roadmaps string) (htmltxt, jstxt string, widgets []string) {
	if component.Type.String() == "JSON" {
		if component.IsArray() {
			a_a := component.Array()
			for _, a := range a_a {
				icon := a.Get("icon").String()
				action := a.Get("action").String()
				scene := a.Get("scene").String()
				style := a.Get("style").String()
				htmltxt += strings.Repeat("&nbsp;", 3)
				htmltxt += `<span sub="{{sub}}" iid="{{iid}}" rmi="{{rmi}}" class="` + roadmaps + `_` + action + `">`
				htmltxt += `<i class="actionbtn `
				if strings.HasPrefix(icon, "fa-") {
					htmltxt += `fa ` + icon
				}
				htmltxt += `"`
				if len(style) > 0 {
					htmltxt += ` style="` + style + `"`
				}
				htmltxt += `></i></span>`
				switch action {
				case "popform":
					widgets = append(widgets, "popform")
					jstxt += `$('.` + roadmaps + `_` + action + `').live('click',function(){`
					jstxt += `	var instance_id=$(this).attr('iid');`
					jstxt += `	var rmi=$(this).attr('rmi');`
					jstxt += `	$('body').Popform({i18n:page_i18n,`
					jstxt += `		eid:{{.entity_id}},rmi:rmi,`
					jstxt += `		instance_name: instance_name,`
					jstxt += `		scene:'` + scene + `',`
					jstxt += `		afterSave:function(iid){cascadeRefresh(rmi,iid);}`
					jstxt += `	}).setInstance(instance_id);`
					jstxt += `});`
				case "remove":
					widgets = append(widgets, "yesno")
					jstxt += `$('.` + roadmaps + `_` + action + `').live('click',function(){`
					jstxt += `	var instance_id=$(this).attr('iid');`
					jstxt += `	var rmi=$(this).attr('rmi');`
					jstxt += `	var subentity=$(this).attr('sub');`
					jstxt += `	$('body').YesnoAlert({`
					jstxt += `		yesText:'{{.t_yes}}',noText:'{{.t_no}}',`
					jstxt += `		doyes: function(id,action){`
					jstxt += `			$.ajaxSettings.async = false;`
					jstxt += `			$.getJSON('/instanceoperate',{eid:{{.entity_id}},iid:instance_id,sub:subentity,act:action},function(m){`
					jstxt += `				if(m.Code=="100"){`
					jstxt += `					cascadeRefresh(rmi,instance_id);`
					jstxt += `				}else{alert(m.Msg);}`
					jstxt += `			});`
					jstxt += `			$.ajaxSettings.async = true;`
					jstxt += `		}`
					jstxt += `	}).show_alertpane('','{{.txt_removeornot}}?','` + action + `');`
					jstxt += `});`
				}
			}
		}
	}
	return
}

func ParseGrid(grid *gjson.Result, identifier, gridscene, clientlanguage_code string) (rows_per_page int, columns []ColumnT, shortcut_block, hot_block, shortcut_js, shortcut_case, action_js string, reference_properties, dependencies []string) {
	hotkeys := []hotkeyT{}
	shortcuts := []shortcutT{}
	g_type := grid.Get("type").String()
	if g_type == "grid" {
		rows_per_page = int(grid.Get("rows_per_page").Int())
		if rows_per_page == 0 {
			rows_per_page = base.Str2int(base.GetConfigurationSimple("UI_ROWSPERPAGE"))
		}
		refers := []string{"id"}
		grid.ForEach(func(k, v gjson.Result) bool {
			name := k.String()
			if name == "_shortcut_" {
				if v.Type.String() == "JSON" {
					if v.IsArray() {
						a_a := v.Array()
						for _, a := range a_a {
							var s shortcutT
							s.Icon = a.Get("icon").String()
							s.Action = a.Get("action").String()
							s.Caption = base.LanguageLabel(a.Get("caption").String(), clientlanguage_code)
							bHotkey := a.Get("hotkey").Bool()
							caption := s.Caption
							if len(s.Icon) > 0 {
								caption = base.Iconhtml(s.Icon) + caption
							}
							scene := a.Get("scene").String()
							if s.Action == "popdoc" {
								scene = strings.ReplaceAll(scene, "{{identifier}}", identifier)
							}
							js, dependency := shortcut_JS(bHotkey, gridscene, s.Action, scene, caption)
							base.MergeDependency(&dependencies, dependency)
							shortcut_js += js

							shortcuts = append(shortcuts, s)
							shortcut_case += `case '` + s.Action + `':` + s.Action + `();break;`
							if bHotkey {
								var hot hotkeyT
								hot.Icon = a.Get("icon").String()
								hot.Action = a.Get("action").String()
								hot.Caption = base.LanguageLabel(a.Get("caption").String(), clientlanguage_code)
								hot.Style = a.Get("style").String()
								hotkeys = append(hotkeys, hot)
							}
						}
					}
				}
			} else if !isReservedWord(name) {
				if v.Type.String() == "JSON" {
					if len(name) > 0 {
						var col ColumnT
						property := v.Get("property").String()
						col.Name = name
						col.Query = v.Get("query").String()
						col.Transform = v.Get("transform").String()
						col.Hyperlink = v.Get("hyperlink").String()
						col.Mark = v.Get("mark").String()
						col.Onclick = v.Get("onclick").String()
						col.Align = v.Get("text-align").String()
						col.Width = v.Get("width").String()
						col.Order = v.Get("order").String()
						col.Render = v.Get("render").String()
						col.Comparison = v.Get("comparison").String()
						col.Fixed = v.Get("fixed").String()
						col.Property = property
						/*if v.Get("all_language").String() == "true" { //use for GROUP_CONCAT order by o.id
							if !strings.HasSuffix(property, "*") {
								property += "*"
							}
						}*/
						col.Visible = (v.Get("visible").String() != "false")
						col.Caption = base.LanguageLabel(v.Get("caption").String(), clientlanguage_code)
						actions := v.Get("actions")
						if actions.Exists() {
							if actions.IsArray() {
								a_a := actions.Array()
								for _, a := range a_a {
									var action actionT
									scene := a.Get("scene").String()
									action.Icon = a.Get("icon").String()
									action.Action = a.Get("action").String()
									action.Scene = scene
									action.Caption = base.LanguageLabel(a.Get("caption").String(), clientlanguage_code)
									action.Hint = base.LanguageLabel(a.Get("hint").String(), clientlanguage_code)
									condition := a.Get("condition").String()
									properties := extractProperty(condition)
									for _, v := range properties {
										exists, _ := base.In_array(v, refers)
										if !exists {
											refers = append(refers, v)
										}
									}
									action.Condition = condition
									js, dependency := action_JS(scene, action.Action, scene, action.Hint)
									action_js += js
									base.MergeDependency(&dependencies, dependency)
									col.Actions = append(col.Actions, action)
								}
							}
						}
						columns = append(columns, col)
					}
				}
			}
			return true // keep iterating
		})
		n := len(columns)
		for _, v := range refers {
			flag := true
			for i := 0; i < n; i++ {
				if v == columns[i].Property {
					flag = false
					break
				}
			}
			if flag {
				reference_properties = append(reference_properties, v)
			}
		}
	}
	n := len(hotkeys)
	for i := 0; i < n; i++ {
		hot := hotkeys[i]
		hot_block += strings.Repeat("&nbsp;", 3)
		hot_block += `<i id="` + gridscene + `_` + hot.Action + `_btn" class="toolbtn `
		if strings.HasPrefix(hot.Icon, "fa-") {
			hot_block += `fa ` + hot.Icon + ` fa-lg`
		}
		hot_block += `"`
		if len(hot.Style) > 0 {
			hot_block += ` style="` + hot.Style + `"`
		}
		hot_block += `></i>`
	}
	sb, _ := json.Marshal(shortcuts)
	shortcut_block = string(sb)
	return
}

//---cascade
/*
font可以按顺序设置如下属性：
font-style:		normal/italic/oblique/inherit
font-variant:	normal/small-caps/inherit
font-weight:	normal/bold/bolder/lighter/100-400(normal)-700(bold)-900/inherit
font-size/line-height:	12px/16px	xx-smal/x-small/small/medium/large/x-large/xx-large
font-family:
*/

func getStyle(component gjson.Result, properties string) (style []string) {
	propertyy := strings.Split(properties, ",")
	n := len(propertyy)
	for i := 0; i < n; i++ {
		name := propertyy[i]
		if len(name) > 0 {
			value := component.Get(name).String()
			if len(value) > 0 {
				style = append(style, name+":"+value)
			}
		}
	}
	return
}

func add_property(properties, renderings, transforms *[]string, property, render, transform string) {
	exists, idx := base.In_array(property, *properties)
	if exists {
		if len(render) > 0 {
			(*renderings)[idx] = render
		}
		if len(transform) > 0 {
			(*transforms)[idx] = transform
		}
	} else {
		*properties = append(*properties, property)
		*renderings = append(*renderings, render)
		*transforms = append(*transforms, transform)
	}
}

func exactProperty(txt string) (property string) {
	property = txt
	if strings.Contains(txt, "^") {
		ss := strings.Split(txt, "^")
		if len(ss) == 2 {
			property = ss[0]
		}
	}
	return
}

type inputT struct {
	Property   string //language_id
	Caption    string
	FormID     string //language
	InputType  string
	InputParam string
	Default    string
}

func Filter2html(definition, object_definition gjson.Result, clientlanguage_code, sessionvalues string) (html string, properties []string, json_inputs string, input_types []string) {
	if definition.Get("type").String() == "filter" {
		subentity := definition.Get("subentity").String()
		verticalalign := definition.Get("vertical-align").String()
		html, json_inputs = "", ""
		definition.ForEach(func(k, v gjson.Result) bool {
			if v.Type.String() == "JSON" {
				txt, propertyy, inputss := filtercomponent2html(k.String(), v, verticalalign, clientlanguage_code, sessionvalues)
				html += txt
				for _, property := range propertyy { //make unique
					exists := false
					if len(properties) > 0 {
						exists, _ = base.In_array(property, properties)
					}
					if !exists {
						properties = append(properties, property)
					}
				}
				n := len(inputss)
				if n > 0 {
					for i := 0; i < n; i++ {
						if len(json_inputs) > 0 {
							json_inputs += ","
						}
						json_inputs += `{"property":"` + inputss[i].Property + `",`
						json_inputs += `"caption":"` + inputss[i].Caption + `",`
						json_inputs += `"id":"` + inputss[i].FormID + `",`
						json_inputs += `"type":"` + inputss[i].InputType + `",`
						exists, _ := base.In_array(inputss[i].InputType, input_types)
						if !exists {
							input_types = append(input_types, inputss[i].InputType)
						}
						params := []string{}
						switch inputss[i].InputType {
						case "daterange":
							//if len(inputss[i].Default) > 0 {
							json_inputs += `"default":"` + inputss[i].Default + `",`
							//}
						case "chooser":
							o := object_definition
							r := o.Get(inputss[i].Property + ".options")
							if r.Exists() {
								params = append(params, `"entity":"`+r.String()+`"`)
							}
						case "selector":
							o := object_definition
							if len(subentity) > 0 {
								o = o.Get(subentity)
							}
							//fmt.Println("============", o.String())
							r := o.Get(inputss[i].Property + ".options")
							if r.Exists() {
								params = append(params, `"codeset":"`+r.String()+`"`)
							}
							t := o.Get(inputss[i].Property + ".type")
							if t.Exists() {
								tt := "ids"
								switch t.String() {
								case "string":
									tt = "codes"
								case "dotids":
									tt = "dotids"
								}
								params = append(params, `"result_type":"`+tt+`"`)
							}
							params = append(params, `"language":"`+base.Language_id(clientlanguage_code)+`"`)
							ss := strings.Split(inputss[i].InputParam, "|") //selector:multiple
							exists, _ := base.In_array("multiple", ss)
							if exists {
								params = append(params, `"multiple_choice":true`)
							}
						}
						json_inputs += `"param":{` + strings.Join(params, ",") + `}}`
					}
				}
			}
			return true // keep iterating
		})
		json_inputs = "[" + json_inputs + "]"
	}
	return
}

func filtercomponent2html(name string, component gjson.Result, verticalalign, clientlanguage_code, sessionvalues string) (html string, properties []string, inputs []inputT) {
	if name == "html" {
		html = component.Get("text").String()
		return
	}
	mapSession := base.String2MapOper(sessionvalues, "&", "=")
	html = ""
	inline_style := []string{} //float:left影响上级div撑开及vertical-align:middle
	inline_style = append(inline_style, "vertical-align:middle")
	inline_style = append(inline_style, "display:inline-block")

	caption := base.LanguageLabel(component.Get("caption").String(), clientlanguage_code)
	if len(caption) > 0 {
		html += `<span class="caption" style="` + strings.Join(inline_style, ";") + `">` + caption + `:</span>`
	}
	class := component.Get("class").String() //"class": "f-value"
	property := component.Get("property").String()
	if len(property) > 0 {
		properties = append(properties, property)
	}
	editor := component.Get("editor").String()
	if len(editor) > 0 {
		sclass := ""
		if len(class) > 0 {
			sclass += ` class="` + class + `"`
		}
		defaultvalue := component.Get("default").String()
		autofocus := component.Get("autofocus").String()
		inputtype, inputparameter := editor, ""
		ss := strings.Split(editor, ":")
		if len(ss) == 2 {
			inputtype = ss[0]
			inputparameter = ss[1]
		}
		switch inputtype {
		case "input":
			html += `<input id="` + name + `"` + sclass
			if len(autofocus) > 0 {
				html += ` autofocus="autofocus"`
			}
			is := getStyle(component, "width,font")
			is = append(is, inline_style...)
			is = append(is, "outline:none")
			html += ` style="` + strings.Join(is, ";") + `">`
			//html += ` value="{{` + property + `}}">`
		case "chooser":
			html += `<div id="` + name + `" class="dd_chooser" tabindex="0"`
			is := getStyle(component, "width")
			is = append(is, inline_style...)
			is = append(is, "outline:none")
			html += ` style="` + strings.Join(is, ";") + `"`
			html += `></div>`
		case "selector":
			html += `<div id="` + name + `" class="h_selector" tabindex="0"`
			is := getStyle(component, "width")
			is = append(is, inline_style...)
			is = append(is, "outline:none")
			html += ` style="` + strings.Join(is, ";") + `"`
			html += `></div>`
		case "wenhao":
			html += `<span id="` + name + `" class="wenhao"></span>`
		case "daterange":
			html += `<span id="` + name + `" class="daterange" tabindex="10"`
			/*is := getStyle(component, "width")
			is = append(is, "outline:none")
			html += ` style="` + strings.Join(is, ";") + `"`*/
			html += `></span>`
			switch defaultvalue {
			case "today":
				dateformat := "yyyy-mm-dd"
				if df, ok := mapSession["dateformat_preferred"]; ok {
					dateformat = df
				}
				var today = time.Now().Format(base.DateTimeLayout(dateformat))
				defaultvalue = today + "," + today
			}
		}
		//Property,Caption,FormID,InputType,InputParam,Default
		inputs = append(inputs, inputT{property, caption, name, inputtype, inputparameter, defaultvalue})
	} else {
		if len(property) > 0 {
			html += "<div"
			html += ` style="` + strings.Join(inline_style, ";") + `"`
			html += `>{{` + property + `}}</div>`
		}
	}
	child_valign := component.Get("vertical-align").String()
	component.ForEach(func(k, v gjson.Result) bool {
		if v.Type.String() == "JSON" {
			txt, propertyy, inputss := filtercomponent2html(k.String(), v, child_valign, clientlanguage_code, sessionvalues)
			html += txt
			if len(propertyy) > 0 {
				properties = append(properties, propertyy...)
			}
			if len(inputss) > 0 {
				inputs = append(inputs, inputss...)
			}
		}
		return true //keep iterating
	})
	return
}

func articleToolbar() (txt string) {
	txt = `<div style="clear:both;width:100%;text-align:right;margin:2px 0 2px 0" rel="article toolbar">`
	txt += `<i class="fa fa-file-pdf-o fa-lg pdf_download"></i>`
	txt += `<span>&nbsp;</span>`
	txt += `<i class="fa fa-qrcode fa-lg qrcode"></i>`
	txt += `<span>&nbsp;</span>`
	txt += `<i class="fa fa-angle-double-up fa-lg ht_collapse_all"></i>`
	txt += `<span>&nbsp;</span>`
	txt += `<i class="fa fa-angle-double-down fa-lg ht_expand_all"></i>`
	txt += `<span>&nbsp;</span>`
	txt += `</div>`
	return
}

func CodesetGridTitle(o gjson.Result, identifier, clientlanguage_code string) (title, sqltxt, fields, colfields string, multiplelanguage bool, e error) {
	langs := base.AcceptLanguages(1)                                                    // (languages []LanguageT)
	multiple_language := (o.Get("language").String() == "multiple") && (len(langs) > 1) //"language": "multiple",
	self_relationship := (o.Get("self_relationship").String() == "hierarchical")        //"self_relationship": "hierarchical",
	caption := base.LanguageLabel(o.Get("caption").String(), clientlanguage_code)
	title = `[{"title":"` + caption + `","width":240,"editable":false}`
	colcol := []string{}
	sqlsql := []string{}
	lala := []string{}
	o.ForEach(func(k, v gjson.Result) bool {
		if v.Type.String() == "JSON" {
			key := k.String()
			caption = base.LanguageLabel(v.Get("caption").String(), clientlanguage_code)
			if len(caption) == 0 {
				caption = key
			}
			switch key {
			case "id":
				title += `,{"title":"` + caption + `","_key":"` + key + `","hidden":true}`
				sqlsql = append(sqlsql, "o."+key)
				colcol = append(colcol, key)
			case "parentid":
				if self_relationship {
					sqlsql = append(sqlsql, "o."+key)
				}
			case "code":
				title += `,{"title":"` + caption + `","_key":"` + key + `","width":100}`
				sqlsql = append(sqlsql, "o."+key)
				colcol = append(colcol, key)
			case "name", "description":
				title += `,{"title":"` + caption + `"`
				sqlsql = append(sqlsql, "o."+key)
				if multiple_language && v.Get("language_adaptive").Bool() {
					sqlsql = append(sqlsql, "ol."+key)
					lala = append(lala, key)
					title += `,"colModel":[`
					for i := 0; i < len(langs); i++ {
						if i > 0 {
							title += ","
						}
						title += `{"title":"` + langs[i].Language_tag + `","_key":"` + key + `.` + langs[i].Language_id + `","width":120}`
						colcol = append(colcol, key+"."+langs[i].Language_id)
					}
					title += "]"
				} else {
					title += `,"_key":"` + key + `","width":120`
					colcol = append(colcol, key)
				}
				title += `}`
			}
		}
		return true //keep iterating
	})
	title += "]"
	if self_relationship {
		sqlsql = append(sqlsql, "o.depth")
	}
	if len(lala) > 0 {
		sqlsql = append(sqlsql, "ol.language_id")
	}
	sqltxt = "SELECT " + strings.Join(sqlsql, ",")
	sqltxt += " FROM " + identifier + " o"
	if len(lala) > 0 {
		sqltxt += " LEFT JOIN " + identifier + "_languages ol ON o.id=ol." + identifier + "_id"
		multiplelanguage = true
	}
	sqltxt += " ORDER BY "
	if self_relationship {
		sqltxt += "o.parentid,"
	}
	sqltxt += "o.ordinalposition"
	fields = strings.Join(sqlsql, ",")
	colfields = strings.Join(colcol, ",")
	return
}

/*colModel: https://www.paramquery.com/demos/showhide_groups*/

func extractData(gr gjson.Result, path string) (data string) {
	var sr gjson.Result
	pp := strings.SplitN(path, ".", 2)
	n := len(pp)
	if n > 0 {
		if gr.IsArray() {
			for _, vv := range gr.Array() {
				sr = vv.Get(pp[0])
				break
			}
		} else {
			sr = gr.Get(pp[0])
		}
		if sr.IsObject() || sr.IsArray() {
			if n == 2 {
				data = extractData(sr, pp[1])
			} else {
				data = sr.String()
			}
		} else {
			data = sr.String()
		}
	}
	return
}

func TransformData(property, data, transform, sessionvalues string) (dt string) {
	if len(transform) > 0 {
		mapSession := base.String2MapOper(sessionvalues, "&", "=")
		tf := strings.SplitN(transform, "#", 2)
		if len(tf) == 2 {
			param := tf[1]
			switch tf[0] {
			case "replace": /* src/des;src1/des1;src2/des2 */
				dt = data
				sdsd := strings.Split(param, ";")
				for i := 0; i < len(sdsd); i++ {
					s_d := strings.Split(sdsd[i], "/")
					if len(s_d) == 2 {
						dt = strings.ReplaceAll(dt, s_d[0], s_d[1])
					}
				}
			case "equalto":
				dt = "false"
				if strings.Compare(data, param) == 0 {
					dt = "true"
				}
			case "fill_templet": // /view?idf=article&iid=[id]
				dt = strings.Replace(param, "["+property+"]", data, -1)
			case "cut_left":
				maxlen := base.Str2int(param)
				if maxlen > 0 {
					dt = base.FirstWords(data, maxlen)
				}
			case "data_extract":
				if len(data) > 0 && len(param) > 0 { //images.src
					if gjson.Valid(data) {
						dt = extractData(gjson.Parse(data), param)
					}
				}
			case "date_format":
				if v, ok := mapSession["dateformat_preferred"]; ok {
					param = v
				}
				tm, e := base.Str20time(data)
				if e == nil {
					if !tm.IsZero() && !base.IsZero(tm) {
						dt = tm.Format(base.DateTimeLayout(param))
					}
				}
			case "datetime_format":
				dateformat, timeformat := "2006-01-02", "15:04:05"
				if v, ok := mapSession["dateformat_preferred"]; ok {
					dateformat = v
				}
				if v, ok := mapSession["timeformat_preferred"]; ok {
					timeformat = v
				}
				tm, e := base.Str20time(data)
				if e == nil {
					if param == "humanized" {
						clientlanguage_id := base.BaseLanguage_id()
						if v, ok := mapSession["clientlanguage_id"]; ok {
							clientlanguage_id = v
						}
						dt = base.HumanizedTime(tm, dateformat, timeformat, clientlanguage_id)
					} else {
						if len(param) == 0 {
							param = dateformat + " " + timeformat
						}
						if !tm.IsZero() && !base.IsZero(tm) {
							dt = tm.Format(base.DateTimeLayout(param))
						}
					}
				}
			case "poplink":
				dt = `<a href="/poplink?z=` + base.EncodeParam(data) + `" target="_blank">` + param + `</a>`
			case "hyperlink":
				dt = `<a href="` + data + `">` + param + `</a>`
			}
		} else {
			switch transform {
			case "LF2BR":
				dt = strings.ReplaceAll(data, "\n", "<br>")
			case "base64":
				dt = "bs64:" + base64.StdEncoding.EncodeToString([]byte(data))
			}
		}
	} else {
		dt = data
	}
	return
}

func Appendix2html(b []byte) (html string, e error) {
	type appendixT struct {
		Tag string `json:"tag"`
		Src string `json:"src"`
	}
	type appendixesT struct {
		Columns    int         `json:"columns"`
		Align      string      `json:"align"`
		Valign     string      `json:"valign"`
		Appendixes []appendixT `json:"appendixes"`
	}
	var apxes appendixesT
	e = json.Unmarshal(b, &apxes)
	if e == nil {
		var ncols = apxes.Columns
		if ncols == 0 {
			ncols = 1
		}
		var n = len(apxes.Appendixes)
		if n > 0 {
			colwidths := base.CalcColumnWidth(ncols)
			nrows := (n + ncols - 1) / ncols
			html = `<table width="100%">`
			for j := 0; j < nrows; j++ {
				html += "<tr>"
				for i := 0; i < ncols; i++ {
					style := []string{"width:" + colwidths[i] + "%"}
					class := ""
					var apx appendixT
					k := j*ncols + i
					if k < n {
						class = "used"
						apx = apxes.Appendixes[k]
						align := apxes.Align
						if len(align) == 0 {
							align = "left"
						}
						style = append(style, "text-align:"+align)
						valign := apxes.Valign
						if len(valign) == 0 {
							valign = "middle"
						}
						style = append(style, "vertical-align:"+valign)
					} else {
						class = "blank"
					}
					sstyle := ` style="` + strings.Join(style, ";") + `"`
					sclass := ` class="` + class + `"`
					html += "<td" + sclass + sstyle + ">"
					if k < n {
						html += `<a id="` + base.StrMD5(apx.Src) + `" href="` + apx.Src + `">` + apx.Tag + `</a>`
					}
					html += "</td>"
				}
				html += "</tr>"
			}
			html += "</table>"
		}
	}
	return
}

func Illustration2html(b []byte) (html string, e error) {
	type imageT struct { //same as entity
		Align  string `json:"align"`
		Valign string `json:"valign"`
		Tag    string `json:"tag"`
		Src    string `json:"src"`
	}
	type illustrationT struct {
		Columns int      `json:"columns"`
		Align   string   `json:"align"`
		Valign  string   `json:"valign"`
		Images  []imageT `json:"images"`
	}
	var illustration illustrationT
	e = json.Unmarshal(b, &illustration)
	if e == nil {
		var ncols = illustration.Columns
		if ncols == 0 {
			ncols = 1
		}
		var n = len(illustration.Images)
		if n > 0 {
			colwidths := base.CalcColumnWidth(ncols)
			nrows := (n + ncols - 1) / ncols
			html = `<div class="ht_gridblock"><table width="100%">`
			for j := 0; j < nrows; j++ {
				html += "<tr>"
				for i := 0; i < ncols; i++ {
					style := []string{"width:" + colwidths[i] + "%"}
					class := ""
					var image imageT
					k := j*ncols + i
					if k < n {
						class = "used"
						image = illustration.Images[k]
						align := image.Align
						if len(align) == 0 {
							align = illustration.Align
						}
						style = append(style, "text-align:"+align)
						valign := image.Valign
						if len(valign) == 0 {
							valign = illustration.Valign
						}
						style = append(style, "vertical-align:"+valign)
					} else {
						class = "blank"
					}
					sstyle := ` style="` + strings.Join(style, ";") + `"`
					sclass := ` class="` + class + `"`
					html += "<td" + sclass + sstyle + ">"
					if k < n {
						html += `<img style="max-width:100%" id="` + base.StrMD5(image.Src) + `" src="` + image.Src + `" alt="` + image.Tag + `">`
					}
					html += "</td>"
				}
				html += "</tr>"
			}
			html += "</table></div>"
		}
	}
	return
}

type imageRT struct {
	Url string `json:"url"`
	Tag string `json:"tag"`
}

func AssembleRender(render, val string) (txt string, e error) {
	flag := false
	if strings.Contains(render, "?") && strings.Contains(render, ":") {
		r, err := regexp.Compile("^[a-z0-9]+\\?[a-z]+:[a-z]+$")
		if err == nil {
			if r.MatchString(render) {
				ss := strings.Split(render, "?")
				if len(ss) == 2 {
					vv := strings.Split(ss[1], ":")
					if len(vv) == 2 {
						if ss[0] == val {
							txt = vv[0]
						} else {
							txt = vv[1]
						}
						txt = "{{.t_" + txt + "}}"
					}
					flag = true
				}
			}
		} else {
			e = errors.New(render + ":" + err.Error())
		}
	}
	if !flag {
		k, v := base.SplitK_V(render, ":")
		switch k {
		case "decoder":
			txt = `<div class="decoder" dt="` + val + `"></div>`
		case "qrcode":
			txt = `<img src="/qrcode?`
			el := base.Str2int(v)
			if el > 0 { //edgelength
				txt += "el=" + strconv.Itoa(el) + "&"
			}
			txt += `dt=` + val + `">`
		case "codestructurediagram":
			if strings.Contains(val, "|") {
				txt = `<img src="/csd?dt=` + base.EncodeParam(val) + `">`
			}
		case "pictable":
			txt = `<table class="pictable" bs64dt="` + base.Encode("base64", val) + `" param="` + v + `" style="margin:0 auto;"></table>`
		case "mdtable":
			txt = `<div class="mdtable" dt="` + val + `" param="` + v + `" style="margin:0 auto;"></div>`
		case "fdtable":
			txt = `<div class="fdtable" dt="` + val + `" param="` + v + `" style="margin:0 auto;"></div>`
		case "embeddedview":
			txt = `<div class="embeddedview" dt="` + val + `" param="` + v + `" style="margin:0 auto;"></div>`
		case "tianditu":
			width, height := base.SplitK_V(v, "*")
			wh := "width: "
			if len(width) > 0 {
				wh += width
			} else {
				wh += "640"
			}
			wh += "px; height: "
			if len(height) > 0 {
				wh += height
			} else {
				wh += "480"
			}
			wh += "px;"
			txt = `<div id="tianditu" class="tianditu" address="` + val + `" style="` + wh + ` margin:0 auto; border: 1px solid gray;"></div>`
		case "appendixes":
			if strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}") {
				txt, e = Appendix2html([]byte(val))
			}
		case "image":
			width, height := base.SplitK_V(v, "*")
			wh := ""
			if len(width) > 0 {
				wh += ` width="` + width + `"`
			}
			if len(height) > 0 {
				wh += ` height="` + height + `"`
			}

			//{"columns":1,"images":[{"tag":"bigdata.gif","src":"/u?n=illu_d6b34d4d6feedc93f6c614030d18d28f-d0f49d88c867aef9f907e745ba83449b.gif","extension":".gif"}]}
			if strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}") {
				txt, e = Illustration2html([]byte(val))
			} else {
				ss := strings.Split(val, ",")
				m := len(ss)
				for j := 0; j < m; j++ {
					data := ss[j]
					if len(data) > 0 {
						if strings.Contains(data, ".") {
							txt += `<img src="` + data + `"` + wh + `>`
						} else {
							b, e := base64.StdEncoding.DecodeString(data)
							if e == nil {
								var img imageRT
								e = json.Unmarshal(b, &img)
								if e == nil {
									txt += `<img src="` + img.Url + `" alt="` + img.Tag + `"` + wh + `>`
								}
							}
						}
					}
				}
			}
		default:
			txt = val
		}
	}
	return
}
func AssembleLeverTxt(lever, param string) (txt string) {
	switch lever {
	case "spinner":
		txt = `<div class="` + lever + `" param="` + param + `"></div>`
	case "gallery":
		txt = `<div class="` + lever + `" param="` + param + `"></div>`
	}
	return
}
func AssembleLever(lever, param, val string) (txt string, e error) {
	flag := false
	/*if strings.Contains(render, "?") && strings.Contains(render, ":") {
		r, err := regexp.Compile("^[a-z0-9]+\\?[a-z]+:[a-z]+$")
		if err == nil {
			if r.MatchString(render) {
				ss := strings.Split(render, "?")
				if len(ss) == 2 {
					vv := strings.Split(ss[1], ":")
					if len(vv) == 2 {
						if ss[0] == val {
							txt = vv[0]
						} else {
							txt = vv[1]
						}
						txt = "{{.t_" + txt + "}}"
					}
					flag = true
				}
			}
		} else {
			e = errors.New(render + ":" + err.Error())
		}
	}*/
	if !flag {
		switch lever {
		case "embeddedview":
			bs64 := base64.StdEncoding.EncodeToString([]byte(val))
			txt = `<div class="` + lever + `" dt="` + bs64 + `" param="` + param + `"></div>`
		case "dimensionradio":
			txt = `<div class="` + lever + `" dt="` + val + `" param="` + param + `"></div>`

		/*
			case "pictable":
				txt = `<table class="pictable" bs64dt="` + base.Encode("base64", val) + `" param="` + v + `" style="margin:0 auto;"></table>`
			case "mdtable":
				txt = `<div class="mdtable" dt="` + val + `" param="` + v + `" style="margin:0 auto;"></div>`
			case "fdtable":
				txt = `<div class="fdtable" dt="` + val + `" param="` + v + `" style="margin:0 auto;"></div>`
		*/
		default:
			txt = val
		}
	}
	return
}

func trimJson(o *gjson.Result, flag bool) (txt string) {
	lnln := []string{}
	o.ForEach(func(k, v gjson.Result) bool {
		ln := ""
		if flag {
			ln += quote(k.String()) + ":"
		}
		switch v.Type.String() {
		case "String":
			ln += quote(v.String())
		case "JSON":
			if v.IsObject() {
				ln += "{" + trimJson(&v, true) + "}"
			} else if v.IsArray() {
				ln += "[" + trimJson(&v, false) + "]"
			}
		default:
			ln += v.String()
		}
		lnln = append(lnln, ln)
		return true
	})
	txt = strings.Join(lnln, ",")
	return
}

func TrimJSON(jsontxt string) (txt string) {
	o := gjson.Parse(jsontxt)
	if o.Exists() {
		txt = "{" + trimJson(&o, true) + "}"
	}
	return
}
