/*
Copyright 2014 The Kubernetes Authors.

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

package buildlog

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
)

// PipeStreamer is a resource that streams the contents of a particular
// pipe
type PipeStreamer struct {
	In          *io.PipeWriter
	Out         *io.PipeReader
	Flush       bool
	ContentType string
}

// a PipeStreamer must implement a rest.ResourceStreamer
var _ rest.ResourceStreamer = &PipeStreamer{}

func (obj *PipeStreamer) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// InputStream returns a stream with the contents of the embedded pipe.
func (s *PipeStreamer) InputStream(apiVersion, acceptHeader string) (stream io.ReadCloser, flush bool, contentType string, err error) {
	flush = s.Flush
	stream = s.Out
	return
}
