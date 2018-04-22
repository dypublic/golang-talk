package golang_exp

import (
	"fmt"
	"sync"
	"time"
	"bytes"
)

/*
Using "nil" Slices and Maps
level: beginner
It's OK to add items to a "nil" slice, but doing the same with a map will produce a runtime panic.
*/
//Works:
func sliceNil() {
	var s []int
	s = append(s,1)
}

//Fails:
func mapNil() {
	var m map[string]int
	m["one"] = 1 //error

}

/*
Unexpected Values in Slice and Array "range" Clauses
level: beginner
This can happen if you are used to the "for-in" or "foreach" statements in other languages. 
The "range" clause in Go is different. 
It generates two values: the first value is the item index while the second value is the item data.
*/
//Bad:

func main1() {
	x := []string{"a","b","c"}

	for v := range x {
		fmt.Println(v) //prints 0, 1, 2
	}
}
//Good:

func main2() {
	x := []string{"a","b","c"}

	for _, v := range x {
		fmt.Println(v) //prints a, b, c
	}
}

/*
Goroutines pattern

One of the most common solutions is to use a "WaitGroup" variable. 
It will allow the main goroutine to wait until all worker goroutines are done. 
If your app has long running workers with message processing loops you'll also need a way to signal those goroutines that it's time to exit. 
You can send a "kill" message to each worker. 
Another option is to close a channel all workers are receiving from. It's a simple way to signal all goroutines at once.
*/

package main

import (
"fmt"
"sync"
)

func main() {
	var wg sync.WaitGroup
	done := make(chan struct{})
	wq := make(chan interface{})
	workerCount := 2

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go doit(i,wq,done,&wg)
	}

	for i := 0; i < workerCount; i++ {
		wq <- i
	}

	close(done)
	wg.Wait()
	fmt.Println("all done!")
}

func doit(workerId int, wq <-chan interface{},done <-chan struct{},wg *sync.WaitGroup) {
	fmt.Printf("[%v] is running\n",workerId)
	defer wg.Done()
	for {
		select {
		case m := <- wq:
			fmt.Printf("[%v] m => %v\n",workerId,m)
		case <- done:
			fmt.Printf("[%v] is done\n",workerId)
			return
		}
	}
}

/*Sending to an Closed Channel Causes a Panic
Receiving from a closed channel is safe. 
The ok return value in a receive statement will be set to false indicating that no data was received. 
If you are receiving from a buffered channel you'll get the buffered data first and once it's empty the ok return value will be false.

Sending data to a closed channel causes a panic. 
It is a documented behavior, but it's not very intuitive for new Go developers who might expect the send behavior to be similar to the receive behavior.
*/
package main

import (
"fmt"
"time"
)

func main() {
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			ch <- (idx + 1) * 2
		}(i)
	}

	//get the first result
	fmt.Println(<-ch)
	close(ch) //not ok (you still have other senders)
	//do other work
	time.Sleep(2 * time.Second)
}
/*Depending on your application the fix will be different.
It might be a minor code change or it might require a change in your application design.
Either way, you'll need to make sure your application doesn't try to send data to a closed channel.
 */

package main

import (
"fmt"
"time"
)

func main() {
	ch := make(chan int)
	done := make(chan struct{})
	for i := 0; i < 3; i++ {
		go func(idx int) {
			select {
			case ch <- (idx + 1) * 2: fmt.Println(idx,"sent result")
			case <- done: fmt.Println(idx,"exiting")
			}
		}(i)
	}

	//get first result
	fmt.Println("result:",<-ch)
	close(done)
	//do other work
	time.Sleep(3 * time.Second)
}

/*
Methods with Value Receivers Can't Change the Original Value

Method receivers are like regular function arguments. 
If it's declared to be a value then your function/method gets a copy of your receiver argument. 
This means making changes to the receiver will not affect the original value unless your receiver is a map or slice variable and you are updating the items in the collection or the fields you are updating in the receiver are pointers.
*/
package main

import "fmt"

type data struct {
	num int
	key *string
	items map[string]bool
}

func (this *data) pmethod() {
	this.num = 7
}

func (this data) vmethod() {
	this.num = 8
	*this.key = "v.key"
	this.items["vmethod"] = true
}

func main() {
	key := "key.1"
	d := data{1,&key,make(map[string]bool)}

	fmt.Printf("num=%v key=%v items=%v\n",d.num,*d.key,d.items)
	//prints num=1 key=key.1 items=map[]

	d.pmethod()
	fmt.Printf("num=%v key=%v items=%v\n",d.num,*d.key,d.items)
	//prints num=7 key=key.1 items=map[]

	d.vmethod()
	fmt.Printf("num=%v key=%v items=%v\n",d.num,*d.key,d.items)
	//prints num=7 key=v.key items=map[vmethod:true]
}
/*
Slice Data "Corruption"
level: intermediate
Let's say you need to rewrite a path (stored in a slice). 
You reslice the path to reference each directory modifying the first folder name and then you combine the names to create a new path.
*/
package main

import (
"fmt"
"bytes"
)

func main() {
	path := []byte("AAAA/BBBBBBBBB")
	sepIndex := bytes.IndexByte(path,'/')
	dir1 := path[:sepIndex]
	dir2 := path[sepIndex+1:]
	fmt.Println("dir1 =>",string(dir1)) //prints: dir1 => AAAA
	fmt.Println("dir2 =>",string(dir2)) //prints: dir2 => BBBBBBBBB

	dir1 = append(dir1,"suffix"...)
	path = bytes.Join([][]byte{dir1,dir2},[]byte{'/'})

	fmt.Println("dir1 =>",string(dir1)) //prints: dir1 => AAAAsuffix
	fmt.Println("dir2 =>",string(dir2)) //prints: dir2 => uffixBBBB (not ok)

	fmt.Println("new path =>",string(path))
}
/*It didn't work as you expected. Instead of "AAAAsuffix/BBBBBBBBB" you ended up with "AAAAsuffix/uffixBBBB".
It happened because both directory slices referenced the same underlying array data from the original path slice.
This means that the original path is also modified. Depending on your application this might be a problem too.


This problem can fixed by allocating new slices and copying the data you need.
Another option is to use the full slice expression.
*/
package main

import (
"fmt"
"bytes"
)

func main() {
	path := []byte("AAAA/BBBBBBBBB")
	sepIndex := bytes.IndexByte(path,'/')
	dir1 := path[:sepIndex:sepIndex] //full slice expression
	dir2 := path[sepIndex+1:]
	fmt.Println("dir1 =>",string(dir1)) //prints: dir1 => AAAA
	fmt.Println("dir2 =>",string(dir2)) //prints: dir2 => BBBBBBBBB

	dir1 = append(dir1,"suffix"...)
	path = bytes.Join([][]byte{dir1,dir2},[]byte{'/'})

	fmt.Println("dir1 =>",string(dir1)) //prints: dir1 => AAAAsuffix
	fmt.Println("dir2 =>",string(dir2)) //prints: dir2 => BBBBBBBBB (ok now)

	fmt.Println("new path =>",string(path))
}
/*
Breaking Out of "for switch" and "for select" Code Blocks
level: intermediate
A "break" statement without a label only gets you out of the inner switch/select block. 
If using a "return" statement is not an option then defining a label for the outer loop is the next best thing.
*/
package main

import "fmt"

func main() {
loop:
	for {
		switch {
		case true:
			fmt.Println("breaking out...")
			break loop
		}
	}

	fmt.Println("out!")
}

/*
Iteration Variables and Closures in "for" Statements

This is the most common gotcha in Go.
The iteration variables in for statements are reused in each iteration.
This means that each closure (aka function literal) created in your for loop will reference the same variable (and they'll get that variable's value at the time those goroutines start executing).

Incorrect:
*/
package main

import (
"fmt"
"time"
)

func main() {
	data := []string{"one","two","three"}

	for _,v := range data {
		go func() {
			fmt.Println(v)
		}()
	}

	time.Sleep(3 * time.Second)
	//goroutines print: three, three, three
}

/*Works:*/

package main

import (
"fmt"
"time"
)

func main() {
	data := []string{"one","two","three"}

	for _,v := range data {
		go func(in string) {
			fmt.Println(in)
		}(v)
	}

	time.Sleep(3 * time.Second)
	//goroutines print: one, two, three
}

/*
"nil" Interfaces and "nil" Interfaces Values

level: advanced
This is the second most common gotcha in Go because interfaces are not pointers even though they may look like pointers.
Interface variables will be "nil" only when their type and value fields are "nil".

The interface type and value fields are populated based on the type and value of the variable used to create the corresponding interface variable.
This can lead to unexpected behavior when you are trying to check if an interface variable equals to "nil".
*/
package main

import "fmt"

func main() {
	var data *byte
	var in interface{}

	fmt.Println(data,data == nil) //prints: <nil> true
	fmt.Println(in,in == nil)     //prints: <nil> true

	in = data
	fmt.Println(in,in == nil)     //prints: <nil> false
	//'data' is 'nil', but 'in' is not 'nil'
}

/*Watch out for this trap when you have a function that returns interfaces.

Incorrect:*/

package main

import "fmt"

func main() {
	doit := func(arg int) interface{} {
		var result *struct{} = nil

		if(arg > 0) {
			result = &struct{}{}
		}

		return result
	}

	if res := doit(-1); res != nil {
		fmt.Println("good result:",res) //prints: good result: <nil>
		//'res' is not 'nil', but its value is 'nil'
	}
}
/*Works:*/

package main

import "fmt"

func main() {
	doit := func(arg int) interface{} {
		var result *struct{} = nil

		if(arg > 0) {
			result = &struct{}{}
		} else {
			return nil //return an explicit 'nil'
		}

		return result
	}

	if res := doit(-1); res != nil {
		fmt.Println("good result:",res)
	} else {
		fmt.Println("bad result (res is nil)") //here as expected
	}
}

/*
Time out, move on
*/

select {
case msg := <-ch:
    // a read from ch has occurred
	process(msg)
case <-time.After(time.Second):
    // the read from ch has timed out
}

/*Let's look at another variation of this pattern. 
In this example we have a program that reads from multiple replicated databases simultaneously. 
The program needs only one of the answers, and it should accept the answer that arrives first.

The function Query takes a slice of database connections and a query string. 
It queries each of the databases in parallel and returns the first response it receives:
*/
func Query(conns []Conn, query string) Result {
    ch := make(chan Result)
    for _, conn := range conns {
        go func(c Conn) {
            select {
            case ch <- c.DoQuery(query):
            default:
            }
        }(conn)
    }
    return <-ch
}