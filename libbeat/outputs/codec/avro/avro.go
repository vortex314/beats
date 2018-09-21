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

package avro

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/elastic/go-structform/gotype"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/outputs/codec"
	"github.com/linkedin/goavro"
)

// Encoder for serializing a beat.Event to avro.
type Encoder struct {
	buf    bytes.Buffer
	folder *gotype.Iterator

	version     string
	config      config
	AvroEncoder *goavro.Codec
	logger      *logp.Logger
}

type config struct {
	File string
}

var defaultConfig = config{
	File: "avro.json",
}

func init() {

	codec.RegisterType("avro", func(info beat.Info, cfg *common.Config) (codec.Codec, error) {
		config := defaultConfig
		if cfg != nil {
			if err := cfg.Unpack(&config); err != nil {
				return nil, err
			}
		}

		return New(config.File, info.Version), nil
	})
}

// New creates a new avro Encoder.
func New(file, version string) *Encoder {

	buffer, er := ioutil.ReadFile(file)
	if er != nil {
		fmt.Println(er)
	}
	avroCodec, err := goavro.NewCodec(string(buffer))
	if err != nil {
		fmt.Println("AVRO Codec not loaded", err)
	}
	//	fmt.Println("AVRO SCHEMA ", avroCodec.Schema())

	e := &Encoder{version: version, config: config{
		File: file,
	}, AvroEncoder: avroCodec, logger: logp.NewLogger("avro-codec")}
	e.reset()
	return e
}

func (e *Encoder) reset() {
	/*	visitor := avro.NewVisitor(&e.buf)
		visitor.SetEscapeHTML(e.config.EscapeHTML)*/

	var err error
	/*
		// create new encoder with custom time.Time encoding
		e.folder, err = gotype.NewIterator(visitor,
			gotype.Folders(
				codec.MakeTimestampEncoder(),
				codec.MakeBCTimestampEncoder(),
			),
		)*/
	if err != nil {
		panic(err)
	}
}

// Encode serializes a beat event to avro. It adds additional metadata in the
// `@metadata` namespace.
func (e *Encoder) Encode(index string, event *beat.Event) ([]byte, error) {
	e.logger.Debug(" encoding ", event.Fields)

	e.buf.Reset()
	/*
		textual := []byte(`{"timestamp":1000000,"component":"avro","message":"hello world"}`)

		// Convert textual Avro data (in Avro JSON format) to native Go form
		_, _, err := e.AvroEncoder.NativeFromTextual(textual)
		if err != nil {
			e.logger.Warn("NativeFromTextual", err)
			return nil, err
		}
	*/
	m := make(map[string]interface{})

	found, _ := event.Fields.HasKey("timestamp")
	if !found {
//		event.PutValue("timestamp", time.Now().UnixNano()/1000000)
		event.PutValue("timestamp", time.Now())
	}

	hasHost, _ := event.Fields.HasKey("host")
	if hasHost {
		valueHost, _ := event.GetValue("host")
		switch valueHost.(type) {
		case string: // do nothing it's ok
			{
			}
		default:
			{
				hostName, _ := os.Hostname()
				event.PutValue("host", hostName)
			}
		}
	} else {
		hostName, _ := os.Hostname()
		event.PutValue("host", hostName)
	}

	for k, v := range event.Fields {
		m[k] = v
	}

	buf, er := e.AvroEncoder.BinaryFromNative(nil, m)
	if er != nil {
		e.logger.Warn("BinaryFromNative", er)
		e.logger.Warn(" content ", m)
		return nil, er
	}

	if e.logger.IsDebug() {
		n, _, _ := e.AvroEncoder.NativeFromBinary(buf)
		e.logger.Debug(" returning ", n)
	}
	return buf, nil
}
