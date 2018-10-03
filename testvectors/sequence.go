/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testvectors

import (
	"fmt"
	"testing"
)

// YAMLObject is an object encoded in YAML.
type YAMLObject string

// SchemaDefinition is an object schema. (TODO: get correct type; for now
// assume this is a yaml-formatted string that can be deserialized.)
type SchemaDefinition string

// Sequence is a testcase collection consisting of 1+n Vectors which are being
// run against an InitialState to verify the final result against ExpectedState
// TODO: Maybe switch the naming between sequence and vector?
type Sequence struct {
	Name string `yaml:"name"`

	// To allow multiple sequences to use the same schema.
	SchemaName string `yaml:"schemaName"`

	InitialState YAMLObject `yaml:"initialState"`
	Vectors      []Vector   `yaml:"vectors"`

	ExpectedState YAMLObject `yaml:"expectedState"`
}

// AppendTestVectors adds the given vectors to the global list.
func (s *Sequence) AppendTestVectors(vectors ...Vector) {
	for _, v := range vectors {
		// defend against typos, since I'm expecting people to define tests via YAML.
		if !v.Valid() {
			panic(fmt.Sprintf("Test case %#v is not complete", &v))
		}
		s.Vectors = append(s.Vectors, v)
	}
}

// Run all the sequences' vectors against the initial state and
// fails the test on unexpected conflicts or a mismatching final state
func (s *Sequence) Run(t *testing.T) {
	t.Run(s.Name, func(t *testing.T) {
		t.Parallel()

		state := s.InitialState
		schema, ok := Schemas[s.SchemaName]
		if !ok {
			t.Fatalf("Test %v references schema %v, but it is not defined", s.Name, s.SchemaName)
		}

		for _, v := range s.Vectors {
			t.Run(v.Name(), func(t *testing.T) {
				state = v.Run(t, state, schema)
			})
		}

		if s.ExpectedState != state {
			t.Errorf("Test did not result in the expected state\n-- expected state:\n%v\n-- result:\n%v", s.ExpectedState, state)
		}
	})
}

// Schemas keeps the schemas that may be referenced by the test vectors.
var Schemas = map[string]SchemaDefinition{}

// Sequences is a list of all the sequences; each file in this package can add one
// or more to the list.
var Sequences []*Sequence

// AppendTestSequences adds the given vectors to the global list.
func AppendTestSequences(sequences ...*Sequence) {
	for _, s := range sequences {
		// defend against typos, since I'm expecting people to define tests via YAML.
		if s.Name == "" ||
			s.SchemaName == "" ||
			s.InitialState == "" ||
			s.ExpectedState == "" ||
			len(s.Vectors) < 1 {
			panic(fmt.Sprintf("Test case %#v is not complete", *s))
		}
		Sequences = append(Sequences, s)
	}
}

// RunAllSequences runs all sequences.
func RunAllSequences(t *testing.T) {
	for _, s := range Sequences {
		s.Run(t)
	}
}
