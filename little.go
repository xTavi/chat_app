// You can edit this code!
// Click here and start typing.
package main

type obj struct {
	followers int
}

func main() {
	a := obj{followers: 10}
	b := &a

	b.followers = a.followers + 1
	println(a.followers)
	println(b.followers)
}
