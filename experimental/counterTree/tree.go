package main

import (
	"fmt"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

const debug = true

type TreeCounterNode struct {
	Value  uint64
	Zero   *TreeCounterNode
	One    *TreeCounterNode
	Parent *TreeCounterNode
}

// MEGA EXPERIMENTAL. Start from the right-most bits in the integer and move leftwards
type TreeTopicViewCounter struct {
	root *TreeCounterNode
}

func newTreeTopicViewCounter() *TreeTopicViewCounter {
	return &TreeTopicViewCounter{
		&TreeCounterNode{0, nil, nil, nil},
	}
}

func (counter *TreeTopicViewCounter) Bump(signTopicID int64) {
	var topicID uint64 = uint64(signTopicID)
	var zeroCount = bits.LeadingZeros64(topicID)
	if debug {
		fmt.Printf("topicID int64: %d\n", signTopicID)
		fmt.Printf("topicID int64: %x\n", signTopicID)
		fmt.Printf("topicID int64: %b\n", signTopicID)
		fmt.Printf("topicID uint64: %b\n", topicID)
		fmt.Printf("leading zeroes: %d\n", zeroCount)

		var leadingZeroes = ""
		for i := 0; i < zeroCount; i++ {
			leadingZeroes += "0"
		}
		fmt.Printf("topicID lead uint64: %s%b\n", leadingZeroes, topicID)

		fmt.Printf("---\n")
	}

	var stopAt uint64 = 64 - uint64(zeroCount)
	var spot uint64 = 1
	var node = counter.root
	for {
		if debug {
			fmt.Printf("spot: %d\n", spot)
			fmt.Printf("topicID&spot: %d\n", topicID&spot)
		}
		if topicID&spot == 1 {
			if node.One == nil {
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(node.One)), nil, unsafe.Pointer(&TreeCounterNode{0, nil, nil, node}))
			}
			node = node.One
		} else {
			if node.Zero == nil {
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(node.Zero)), nil, unsafe.Pointer(&TreeCounterNode{0, nil, nil, node}))
			}
			node = node.Zero
		}

		spot++
		if spot >= stopAt {
			break
		}
	}

	atomic.AddUint64(&node.Value, 1)
}
