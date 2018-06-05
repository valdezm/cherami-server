// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type (
	TokenBucketSuite struct {
		*require.Assertions // override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test, not merely log an error
		suite.Suite
	}
)

func TestTokenBucketSuite(t *testing.T) {
	suite.Run(t, new(TokenBucketSuite))
}

func (s *TokenBucketSuite) SetupTest() {
	s.Assertions = require.New(s.T()) // Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
}

type mockTimeSource struct {
	currTime time.Time
}

func (ts *mockTimeSource) Now() time.Time {
	return ts.currTime
}

func (ts *mockTimeSource) advance(d time.Duration) {
	ts.currTime = ts.currTime.Add(d)
}

func (s *TokenBucketSuite) TestRpsEnforced() {
	ts := &mockTimeSource{currTime: time.Now()}
	tb := NewTokenBucket(99, ts)
	for i := 0; i < 2; i++ {
		total := 0
		attempts := 1
		for ; attempts < 11; attempts++ {
			for c := 0; c < 2; c++ {
				if ok, _ := tb.TryConsume(10); ok {
					total += 10
				}
			}

			if total >= 90 {
				break
			}
			ts.advance(time.Millisecond * 101)
		}
		s.Equal(90, total, "Token bucket failed to enforce limit")
		s.Equal(9, attempts, "Token bucket gave out tokens too quickly")
		ts.advance(time.Millisecond * 101)
		ok, _ := tb.TryConsume(9)
		s.True(ok, "Token bucket failed to enforce limit")
		ok, _ = tb.TryConsume(1)
		s.False(ok, "Token bucket failed to enforce limit")
		ts.advance(time.Second)
	}
}

func (s *TokenBucketSuite) TestLowRpsEnforced() {
	ts := &mockTimeSource{currTime: time.Now()}
	tb := NewTokenBucket(3, ts)

	total := 0
	attempts := 1
	for ; attempts < 10; attempts++ {
		for c := 0; c < 2; c++ {
			if ok, _ := tb.TryConsume(1); ok {
				total++
			}
		}
		if total >= 3 {
			break
		}
		ts.advance(time.Millisecond * 101)
	}
	s.Equal(3, total, "Token bucket failed to enforce limit")
	s.Equal(3, attempts, "Token bucket gave out tokens too quickly")
}
