package main

import (
	"context"
	"fmt"
	"time"

	behaviortree "github.com/joeycumines/go-behaviortree"
)

// ExampleNewTickerStopOnFailure_counter demonstrates the use of NewTickerStopOnFailure to implement more complex "run
// to completion" behavior using the simple modular building blocks provided by this package
func ExampleNewTickerStopOnFailure_counter() {
	var (
		// counter is the shared state used by this example
		counter = 0
		// printCounter returns a node that will print the counter prefixed with the given name then succeed
		printCounter = func(name string) behaviortree.Node {
			return behaviortree.New(
				func(children []behaviortree.Node) (behaviortree.Status, error) {
					fmt.Printf("%s: %d\n", name, counter)
					return behaviortree.Success, nil
				},
			)
		}
		// incrementCounter is a node that will increment counter then succeed
		incrementCounter = behaviortree.New(
			func(children []behaviortree.Node) (behaviortree.Status, error) {
				counter++
				return behaviortree.Success, nil
			},
		)
		// ticker is what actually runs this example and will tick the behavior tree defined by a single root node at
		// most once per millisecond and will stop after the first failed tick or error or context cancel
		ticker = behaviortree.NewTickerStopOnFailure(
			context.Background(),
			time.Millisecond,
			behaviortree.New(
				behaviortree.Selector, // runs each child sequentially until one succeeds (success) or all fail (failure)
				behaviortree.New(
					behaviortree.Sequence, // runs each child in order until one fails (failure) or they all succeed (success)
					behaviortree.New(
						func(children []behaviortree.Node) (behaviortree.Status, error) { // succeeds while counter is less than 10
							if counter < 10 {
								return behaviortree.Success, nil
							}
							return behaviortree.Failure, nil
						},
					),
					incrementCounter,
					printCounter("< 10"),
				),
				behaviortree.New(
					behaviortree.Sequence,
					behaviortree.New(
						func(children []behaviortree.Node) (behaviortree.Status, error) { // succeeds while counter is less than 20
							if counter < 20 {
								return behaviortree.Success, nil
							}
							return behaviortree.Failure, nil
						},
					),
					incrementCounter,
					printCounter("< 20"),
				),
			), // if both children failed (counter is >= 20) the root node will also fail
		)
	)
	// waits until ticker stops, which will be on the first failure of it's root node
	<-ticker.Done()
	// every Tick may return an error which would automatically cause a failure and propagation of the error
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	// Output:
	// < 10: 1
	// < 10: 2
	// < 10: 3
	// < 10: 4
	// < 10: 5
	// < 10: 6
	// < 10: 7
	// < 10: 8
	// < 10: 9
	// < 10: 10
	// < 20: 11
	// < 20: 12
	// < 20: 13
	// < 20: 14
	// < 20: 15
	// < 20: 16
	// < 20: 17
	// < 20: 18
	// < 20: 19
	// < 20: 20
}

func main() {
	ExampleNewTickerStopOnFailure_counter()
}
