package types

import (
    "encoding/json"
    "log"
)

type Parameter struct {
    Command string `json:command`
    DefaultValue string `json:defaultValue`
    Value string `json:value`
    Description string `json:description`
}

func (p Parameter) IsCustom() bool {
    if p.Value != "" && p.Value != p.DefaultValue {
        return true
    }
    return false
}

type ParamsMap map[string]Parameter

func (p *ParamsMap) LoadParamsMap(bytes []byte) {
    err := json.Unmarshal(bytes, &p)

    if err != nil {
        log.Fatal(err)
    }
}

func (p *ParamsMap) GetKeysList() (list []string) {
    for key, _ := range *p {
        list = append(list, key)
    }
    return
}