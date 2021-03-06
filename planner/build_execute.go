//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package planner

import (
	"fmt"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

var errBadFormat = fmt.Errorf("prepared must be a json string or object")

func (this *builder) VisitExecute(stmt *algebra.Execute) (interface{}, error) {
	expr := stmt.Prepared()
	if (expr == nil) || expr.Type() != value.STRING && expr.Type() != value.OBJECT {
		return nil, errors.NewUnrecognizedPreparedError(errBadFormat)
	}
	return plan.GetPrepared(expr)
}
