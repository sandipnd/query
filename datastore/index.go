//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package datastore

import (
	"github.com/couchbaselabs/query/errors"
	"github.com/couchbaselabs/query/expression"
	"github.com/couchbaselabs/query/value"
)

type IndexType string
type IndexKey expression.Expressions

const (
	DEFAULT     IndexType = "default"     // default may vary per backend
	UNSPECIFIED IndexType = "unspecified" // used by non-view primary_indexes
	VIEW        IndexType = "view"
	LSM         IndexType = "lsm"
)

// Index is the base type for indexes.
type Index interface {
	KeyspaceId() string // Id of the keyspace to which this index belongs
	Id() string         // Id of this index
	Name() string       // Name of this index
	Type() IndexType    // Type of this index
	Drop() errors.Error // Drop / delete this index
	//Rename(newName string) errors.Error                                 // Rename this index
	EqualKey() expression.Expressions                                   // Equality keys
	RangeKey() expression.Expressions                                   // Range keys
	Condition() expression.Expression                                   // Condition, if any
	Statistics(span *Span) (Statistics, errors.Error)                   // Obtain statistics for this index
	Scan(span *Span, distinct bool, limit int64, conn *IndexConnection) // Perform a scan on this index. Distinct and limit are hints.
}

// PrimaryIndex represents primary key indexes.
type PrimaryIndex interface {
	ScanEntries(limit int64, conn *IndexConnection) // Perform a scan of all the entries in this index
}

// SecondaryIndex is the base type secondary indexes.
type SecondaryIndex interface {
	KeyspaceId() string                                                 // Id of the keyspace to which this index belongs
	Id() string                                                         // Id of this index
	Name() string                                                       // Name of this index
	Type() IndexType                                                    // Type of this index
	Drop() errors.Error                                                 // Drop / delete this index
	Rename(newName string) errors.Error                                 // Rename this index
	EqualKey() expression.Expressions                                   // Equality keys
	RangeKey() expression.Expressions                                   // Range keys
	Condition() expression.Expression                                   // Condition, if any
	Statistics(span *Span) (Statistics, errors.Error)                   // Obtain statistics for this index
	Scan(span *Span, distinct bool, limit int64, conn *IndexConnection) // Perform a scan on this index. Distinct and limit are hints.
}

type Range struct {
	Low       value.Values
	High      value.Values
	Inclusion Inclusion
}

type Ranges []*Range

// Inclusion controls how the boundary values of a range are treated.
type Inclusion int

const (
	NEITHER Inclusion = 0x00
	LOW               = 0x01
	HIGH              = 0x10
	BOTH              = LOW | HIGH
)

type Span struct {
	Equal value.Values
	Range *Range
}

type Spans []*Span

type IndexEntry struct {
	EntryKey   value.Values
	PrimaryKey string
}

type EntryChannel chan *IndexEntry
type StopChannel chan bool

// Statistics captures statistics for a range.
type Statistics interface {
	Count() (int64, errors.Error)
	Min() (value.Values, errors.Error)
	Max() (value.Values, errors.Error)
	DistinctCount(int64, errors.Error)
	Bins() ([]Statistics, errors.Error)
}

type Context interface {
	Fatal(errors.Error)
	Error(errors.Error)
	Warning(errors.Error)
}

type IndexConnection struct {
	entryChannel EntryChannel // Closed by the index when the scan is completed or aborted.
	stopChannel  StopChannel  // Notifies index to stop scanning. Never closed, just garbage-collected.
	context      Context
}

const _ENTRY_CAP = 1024

func NewIndexConnection(context Context) *IndexConnection {
	return &IndexConnection{
		entryChannel: make(EntryChannel, _ENTRY_CAP),
		stopChannel:  make(StopChannel, 1),
		context:      context,
	}
}

func (this *IndexConnection) EntryChannel() EntryChannel {
	return this.entryChannel
}

func (this *IndexConnection) StopChannel() StopChannel {
	return this.stopChannel
}

func (this *IndexConnection) Fatal(err errors.Error) {
	this.context.Fatal(err)
}

func (this *IndexConnection) Error(err errors.Error) {
	this.context.Error(err)
}

func (this *IndexConnection) Warning(wrn errors.Error) {
	this.context.Warning(wrn)
}