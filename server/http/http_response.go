//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/couchbaselabs/query/errors"
	"github.com/couchbaselabs/query/execution"
	"github.com/couchbaselabs/query/server"
	"github.com/couchbaselabs/query/value"
)

func (this *httpRequest) Output() execution.Output {
	return this
}

func (this *httpRequest) Fail(err errors.Error) {
	defer this.Stop(server.FATAL)

	this.resp.WriteHeader(http.StatusInternalServerError)
	this.writeString(err.Error())
}

func (this *httpRequest) Execute(stopNotify chan bool, metrics bool) {
	defer this.Stop(server.COMPLETED)

	this.NotifyStop(stopNotify)

	this.resp.WriteHeader(http.StatusOK)
	_ = this.writePrefix() &&
		this.writeResults() &&
		this.writeSuffix(metrics, server.COMPLETED)
}

func (this *httpRequest) Expire() {
	defer this.Stop(server.TIMEOUT)

	this.writeSuffix(true, server.TIMEOUT)
}

func (this *httpRequest) writePrefix() bool {
	return this.writeString("{\n    \"results\": [")
}

func (this *httpRequest) writeResults() bool {
	var item value.Value

	ok := true
	for ok {
		select {
		case <-this.StopExecute():
			return true
		default:
		}

		select {
		case item, ok = <-this.Results():
			if ok {
				if !this.writeResult(item) {
					return false
				}
			}
		case <-this.StopExecute():
			return true
		}
	}

	return true
}

func (this *httpRequest) writeResult(item value.Value) bool {
	var rv bool
	if this.resultCount == 0 {
		rv = this.writeString("\n")
	} else {
		rv = this.writeString(",\n")
	}

	bytes, err := json.MarshalIndent(item.Actual(), "        ", "    ")
	if err != nil {
		return false
	}

	this.resultCount++

	return rv &&
		this.writeString("        ") &&
		this.writeString(string(bytes))
}

func (this *httpRequest) writeSuffix(metrics bool, state server.State) bool {
	return this.writeString("\n    ]") &&
		this.writeState(state) &&
		this.writeErrors() &&
		this.writeWarnings() &&
		this.writeMetrics(metrics) &&
		this.writeString("\n}\n")
}

func (this *httpRequest) writeString(s string) bool {
	_, err := io.WriteString(this.resp, s)
	this.Flush()
	return err == nil
}

func (this *httpRequest) writeState(state server.State) bool {
	if state == "" {
		state = this.State()
	}

	if state == "completed" {
		return true
	}

	return this.writeString(fmt.Sprintf(",\n    \"state\": \"%s\"", state))
}

func (this *httpRequest) writeErrors() bool {
	var err errors.Error
	ok := true
loop:
	for ok {
		select {
		case err, ok = <-this.Errors():
			if ok {
				if this.errorCount == 0 {
					this.writeString(",\n    \"errors\": [")
				}
				ok = this.writeError(err, this.errorCount)
				this.errorCount++
			}
		default:
			break loop
		}
	}

	return this.errorCount == 0 || this.writeString("\n    ]")
}

func (this *httpRequest) writeWarnings() bool {
	var err errors.Error
	ok := true
loop:
	for ok {
		select {
		case err, ok = <-this.Warnings():
			if ok {
				if this.warningCount == 0 {
					this.writeString(",\n    \"warnings\": [")
				}
				ok = this.writeError(err, this.warningCount)
				this.warningCount++
			}
		default:
			break loop
		}
	}

	return this.warningCount == 0 || this.writeString("\n    ]")
}

func (this *httpRequest) writeError(err errors.Error, count int) bool {
	var rv bool
	if count == 0 {
		rv = this.writeString("\n")
	} else {
		rv = this.writeString(",\n")
	}

	bytes, er := json.MarshalIndent(err, "        ", "    ")
	if er != nil {
		return false
	}

	return rv &&
		this.writeString("        ") &&
		this.writeString(string(bytes))
}

func (this *httpRequest) writeMetrics(metrics bool) bool {
	m := this.Metrics()
	if m == value.FALSE ||
		(m == value.NONE && !metrics) {
		return true
	}

	rv := this.writeString(",\n    \"metrics\": {") &&
		this.writeString(fmt.Sprintf("\n        elapsedTime: %v", time.Since(this.RequestTime()))) &&
		this.writeString(fmt.Sprintf(",\n        executionTime: %v", time.Since(this.ServiceTime()))) &&
		this.writeString(fmt.Sprintf(",\n        resultCount: %d", this.resultCount))

	if this.errorCount > 0 {
		rv = rv && this.writeString(fmt.Sprintf(",\n        errorCount: %d", this.errorCount))
	}

	if this.warningCount > 0 {
		rv = rv && this.writeString(fmt.Sprintf(",\n        warningCount: %d", this.warningCount))
	}

	return rv && this.writeString("\n    }")
}

func (this *httpRequest) Flush() {
	this.resp.(http.Flusher).Flush()
}