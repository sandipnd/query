//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"time"

	"github.com/couchbaselabs/query/err"
	"github.com/couchbaselabs/query/expression"
	"github.com/couchbaselabs/query/value"
)

type Context interface {
	expression.Context
	Now() time.Time
	Argument(parameter string) (value.Value, error)
	WarningChannel() err.ErrorChannel
	ErrorChannel() err.ErrorChannel
	EvaluateSubquery(query *Select, parent value.Value) (value.Value, error)
}
