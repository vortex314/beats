// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package actions

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/processors"
	"github.com/robertkrimen/otto"
)

type javaScriptEngine struct {
	File   string
	Engine *otto.Otto
	logger *logp.Logger
}

func init() {
	processors.RegisterPlugin("javascript",
		configChecked(newJavaScriptEngine,
			requireFields("file"),
			allowedFields("file", "when")))

}

func newJavaScriptEngine(c *common.Config) (processors.Processor, error) {
	config := struct {
		File string `config:"file"`
	}{}
	logger := logp.NewLogger("javascript")
	err := c.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the grok_Patterns configuration: %s", err)
	}
	buffer, erc := ioutil.ReadFile(config.File)
	if erc != nil {
		logger.Warn(" javascript file read  error occured ", erc)
	}
	engine, value, er := otto.Run(string(buffer))
	if er != nil {
		logger.Warn(" javascript initial run  error occured ", er, value)
	}

	/* remove read only Patterns */
	/*	for _, readOnly := range processors.MandatoryExportedFields {
			for i, field := range config.File {
				if readOnly == field {
					config.Patterns = append(config.Patterns[:i], config.Patterns[i+1:]...)
				}
			}
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	*/
	f := &javaScriptEngine{File: config.File,
		Engine: engine,
		logger: logger}
	return f, nil
}

func (f *javaScriptEngine) Run(event *beat.Event) (*beat.Event, error) {
	var errors []string
	//	message, _ := event.Fields.GetValue("message")
	// call the JS process function
	_, er := f.Engine.Call("process", nil, event)
	//	fmt.Println(fields)
	if er != nil {
		f.logger.Warn("javascript error occured ", er)
	}
	// convert added fields from JS type to native Go
	for k, v := range event.Fields {
		switch v.(type) {
		case otto.Value:
			val, err := v.(otto.Value).Export()
			if err != nil {
				f.logger.Warn("cannot convert variable ", k, v, er)
			} else {
				event.PutValue(k, val)
			}
		default:
		}
	}

	if len(errors) > 0 {
		return event, fmt.Errorf(strings.Join(errors, ", "))
	}
	return event, nil
}

func (f *javaScriptEngine) String() string {
	return "javascript_file=" + f.File
}
