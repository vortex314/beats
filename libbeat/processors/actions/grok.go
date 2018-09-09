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
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"
	"github.com/vjeantet/grok"
)

type grokPatterns struct {
	Patterns []string
	grok     *grok.Grok
}

func init() {
	processors.RegisterPlugin("grok",
		configChecked(newGrokPatterns,
			requireFields("patterns"),
			allowedFields("patterns", "timestamps", "when")))

}

func newGrokPatterns(c *common.Config) (processors.Processor, error) {
	config := struct {
		Patterns   []string `config:"patterns"`
		Timestamps []string `config:"timestamps"`
	}{}
	err := c.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the grok_Patterns configuration: %s", err)
	}

	/* remove read only Patterns */
	for _, readOnly := range processors.MandatoryExportedFields {
		for i, field := range config.Patterns {
			if readOnly == field {
				config.Patterns = append(config.Patterns[:i], config.Patterns[i+1:]...)
			}
		}
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})

	f := &grokPatterns{Patterns: config.Patterns, grok: g}
	return f, nil
}

func (f *grokPatterns) Run(event *beat.Event) (*beat.Event, error) {

	var errors []string

	for _, field := range f.Patterns {
		err := event.Delete(field)
		if err != nil {
			errors = append(errors, err.Error())
		}

	}
	message, _ := event.Fields.GetValue("message")
	values, _ := f.grok.Parse(f.Patterns[0], message.(string))
	for k, v := range values {
		event.PutValue(k, v)
		//		fmt.Printf("%+15s =>  %s\n", k, v)
	}

	if len(errors) > 0 {
		return event, fmt.Errorf(strings.Join(errors, ", "))
	}
	return event, nil
}

func (f *grokPatterns) String() string {
	return "grok_Patterns=" + strings.Join(f.Patterns, ", ")
}
