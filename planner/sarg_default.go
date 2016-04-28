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
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
)

var _SELF_SPANS plan.Spans
var _FULL_SPANS plan.Spans
var _VALUED_SPANS plan.Spans
var _EMPTY_SPANS plan.Spans
var _EXACT_FULL_SPANS plan.Spans
var _EXACT_VALUED_SPANS plan.Spans

func init() {
	sspan := &plan.Span{}
	sspan.Range.Low = expression.Expressions{expression.TRUE_EXPR}
	sspan.Range.Inclusion = datastore.LOW
	_SELF_SPANS = plan.Spans{sspan}

	fspan := &plan.Span{}
	fspan.Range.Low = expression.Expressions{expression.NULL_EXPR}
	fspan.Range.Inclusion = datastore.LOW
	_FULL_SPANS = plan.Spans{fspan}

	vspan := &plan.Span{}
	vspan.Range.Low = expression.Expressions{expression.NULL_EXPR}
	vspan.Range.Inclusion = datastore.NEITHER
	_VALUED_SPANS = plan.Spans{vspan}

	espan := &plan.Span{}
	espan.Range.High = expression.Expressions{expression.NULL_EXPR}
	espan.Range.Inclusion = datastore.NEITHER
	_EMPTY_SPANS = plan.Spans{espan}

	_EXACT_FULL_SPANS = _FULL_SPANS.Copy()
	_EXACT_FULL_SPANS[0].Exact = true

	_EXACT_VALUED_SPANS = _VALUED_SPANS.Copy()
	_EXACT_VALUED_SPANS[0].Exact = true
}

type sargDefault struct {
	sargBase
}

func newSargDefault(pred expression.Expression) *sargDefault {
	var spans plan.Spans
	if pred.PropagatesNull() {
		if _, ok := pred.(*expression.Not); ok {
			spans = _VALUED_SPANS
		} else {
			spans = _EXACT_VALUED_SPANS
		}
	} else if pred.PropagatesMissing() {
		if _, ok := pred.(*expression.Not); ok {
			spans = _FULL_SPANS
		} else {
			spans = _EXACT_FULL_SPANS
		}
	}

	rv := &sargDefault{}
	rv.sarger = func(expr2 expression.Expression) (plan.Spans, error) {
		if SubsetOf(pred, expr2) {
			return _SELF_SPANS, nil
		}

		if spans != nil && pred.DependsOn(expr2) {
			return spans, nil
		}

		return nil, nil
	}

	return rv
}
