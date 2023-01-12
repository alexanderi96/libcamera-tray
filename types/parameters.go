package types

import (
	"encoding/json"
	"log"
)

type Parameter struct {
	Command     	string 	`json:command`
	Value       	string 	`json:value`
	Enabled			bool	`json:enabled`
	StillSpecific	bool	`json:stillSpecific`
	Description 	string 	`json:description`
}

func (p *Parameter) Toggle() {
	p.Enabled = !p.Enabled
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
