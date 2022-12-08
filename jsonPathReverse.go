package jpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type JsonPathInverse struct {
	exps *exps
}

func (j *JsonPathInverse) Set(dst interface{}, val interface{}) error {
	return marshalInterface(j.exps, dst, val)
}

func (j *JsonPathInverse) UnmarshalJSON(b []byte) error {
	strval := ""
	err := json.Unmarshal(b, &strval)
	if err != nil {
		return fmt.Errorf("invalid json rule:%w  :%s", err, string(b))
	}
	j.exps, err = parseExps(strval)
	if err != nil {
		return fmt.Errorf("invalid rule :%w   :%s", err, strval)
	}
	return err
}

const (
	TYPE_KEY = -1
)

// src must be map[string]interface{}

//
//resolve the token is a field or array
//return -1 and field name if token is field
//return the idx of token and array name if token is array
func yyp(token string) (string, int, error) {
	numidx_start := 0
	numidx_end := 0
	for k, v := range token {
		t := string(v)
		if t == "[" {
			numidx_start = k
		}
		if t == "]" {
			numidx_end = k
		}
	}
	if numidx_end > 0 && numidx_start >= 0 {
		num, err := strconv.Atoi(token[numidx_start+1 : numidx_end])
		if err != nil {
			return "", TYPE_KEY, err
		}
		return token[:numidx_start], num, nil
	}
	return token, TYPE_KEY, nil
}

type token struct {
	idx int
	key string
}

type exps struct {
	isroot bool
	tokens []*token
}

func parseExps(query string) (*exps, error) {
	rawtkns := strings.Split(strings.TrimLeft(query, "$."), ".")
	parsedTkns := make([]*token, 0, len(rawtkns))
	for _, rawtkn := range rawtkns {
		key, idx, err := yyp(rawtkn)
		if err != nil {
			return nil, err
		}
		parsedTkns = append(parsedTkns, &token{
			idx: idx,
			key: key,
		})
	}

	exp := &exps{
		tokens: parsedTkns,
	}
	if strings.TrimLeft(query, "$.") == "" {
		exp.isroot = true
	}
	return exp, nil

}

//marshal and set the value to interface{}
//MarshalInterface() is power than Marshal().
// MarshalInterface() support expression such as $ ,$[0] which  Marshal() doesn't support
//attention that $dst must be *interface{}

func MarshalInterface(dst interface{}, query string, value interface{}) error {
	return marshalInterface(query, dst, value)
}

func marshalInterface(exps *exps, dst interface{}, value interface{}) error {
	//fmt.Println(reflect.TypeOf(dst))
	//exps, err := parseExps(query)
	//if err != nil {
	//	return err
	//}
	var cp = dst
	if cpi, ok := dst.(*interface{}); ok {
		if exps.isroot {
			*cpi = value
			return nil
		}
		if _, ok := (*cpi).(map[string]interface{}); ok {
			// do not handle
		} else if _, ok := (*cpi).([]interface{}); ok {
			cp = cpi
			goto done
		} else {
			//yp, idx, err := yyp(tks[0])
			//if err != nil {
			//	return err
			//}
			tkn := exps.tokens[0]
			if tkn.idx == TYPE_KEY {
				*cpi = map[string]interface{}{}
			} else {
				if tkn.key != "" {
					*cpi = map[string]interface{}{}
				} else {
					*cpi = make([]interface{}, tkn.idx+1)
					cp = cpi
					goto done
				}

			}
		}
		cp = *cpi

	}
done:
	return parserToken(exps, cp, value)

}

func parserToken(tks *exps, cp, value interface{}) error {
	for k, v := range tks.tokens {
		//field, idx, err := yyp(v)
		//if err != nil {
		//	return err
		//}
		//
		field := v.key
		idx := v.idx
		if idx == TYPE_KEY {
			cpm, ok := cp.(map[string]interface{})
			if !ok {
				return errors.New(fmt.Sprintf("create field failed ,%s->parent cannot convert_ to map", field))
			}

			if k < len(tks.tokens)-1 {
				if cpm[field] == nil {
					cpm[field] = map[string]interface{}{}
				}
				cpm, ok = cpm[field].(map[string]interface{})
				if !ok {
					return errors.New(fmt.Sprintf("create field failed ,%s cannot convert_ to map", field))
				}
				cp = cpm
			} else {
				//filed is last token ,set value to interface
				cpm[field] = value
			}
		} else { //array
			if field == "" && k == 0 { //root array
				cpi, ok := cp.(*interface{})
				//	fmt.Println(reflect.TypeOf(cp))
				if !ok {
					return errors.New("root is not pointer")
				}
				if _, ok := (*cpi).([]interface{}); !ok {
					return errors.New("root is not array")
				}
				if len((*cpi).([]interface{})) < idx+1 {
					for i := len((*cpi).([]interface{})); i < idx+1; i++ {
						*cpi = append((*cpi).([]interface{}), nil)
					}
				}

				if k < len(tks.tokens)-1 {
					for i := 0; i < idx+1; i++ {
						if (*cpi).([]interface{})[i] == nil {
							(*cpi).([]interface{})[i] = map[string]interface{}{}
						}
					}
					cp = (*cpi).([]interface{})[idx]
				} else {
					//filed is last token ,set value to interface
					(*cpi).([]interface{})[idx] = value
				}
				//fmt.Println((*cpi).([]interface{}))
				continue
			}

			cpm, ok := cp.(map[string]interface{})
			if !ok {
				return errors.New("nil array child")
			}
			if cpm[field] == nil {
				cpm[field] = make([]interface{}, idx+1)
			}
			cps, ok := cpm[field].([]interface{})
			if !ok {
				return errors.New(fmt.Sprintf("create array failed ,%s cannot convert2 to array", field))
			}
			lenmap := len(cps)
			if lenmap < idx+1 {
				for i := lenmap; i < idx+1; i++ {
					cpm[field] = append(cpm[field].([]interface{}), nil)
				}
			}
			if k < len(tks.tokens)-1 {
				for i := 0; i < idx+1; i++ {
					if cpm[field].([]interface{})[i] == nil {
						cpm[field].([]interface{})[i] = map[string]interface{}{}
					}
				}
				cp = cpm[field].([]interface{})[idx]
			} else {
				//filed is last token ,set value to interface
				cpm[field].([]interface{})[idx] = value
			}
		}
	}

	return nil
}

/*
SwitchJson() can switch format of json from $data to $dst by specific expression strings.
attention that $dst must be type of *interface{}
*/

//
func tokenize(query string) ([]string, error) {
	tokens := []string{}
	//	token_start := false
	//	token_end := false
	token := ""

	// fmt.Println("-------------------------------------------------- start")
	for idx, x := range query {
		token += string(x)
		// //fmt.Printf("idx: %d, x: %s, token: %s, tokens: %v\n", idx, string(x), token, tokens)
		if idx == 0 {
			if token == "$" || token == "@" {
				tokens = append(tokens, token[:])
				token = ""
				continue
			} else {
				return nil, fmt.Errorf("should start with '$'")
			}
		}
		if token == "." {
			continue
		} else if token == ".." {
			if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, "*")
			}
			token = "."
			continue
		} else {
			// fmt.Println("else: ", string(x), token)
			if strings.Contains(token, "[") {
				// fmt.Println(" contains [ ")
				if x == ']' && !strings.HasSuffix(token, "\\]") {
					if token[0] == '.' {
						tokens = append(tokens, token[1:])
					} else {
						tokens = append(tokens, token[:])
					}
					token = ""
					continue
				}
			} else {
				// fmt.Println(" doesn't contains [ ")
				if x == '.' {
					if token[0] == '.' {
						tokens = append(tokens, token[1:len(token)-1])
					} else {
						tokens = append(tokens, token[:len(token)-1])
					}
					token = "."
					continue
				}
			}
		}
	}
	if len(token) > 0 {
		if token[0] == '.' {
			token = token[1:]
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		} else {
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		}
	}
	// fmt.Println("finished tokens: ", tokens)
	// fmt.Println("================================================= done ")
	return tokens, nil
}

//check that if rule is a valid rule
var ruleRegexp = regexp.MustCompile(`\$(\[\d+\])?(\.\w+(\[\d+\])?)*$`)

func checkRule(rule string) bool {
	return ruleRegexp.MatchString(rule)
}
