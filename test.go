package main

import (
	"context"
	"log"

	"github.com/garyburd/redigo/redis"
)

type Tree struct {
	Left  *Tree
	Value Order
	Right *Tree
}

type Order struct {
	Side  string
	Price float32
}

var ctx = context.Background()

func insert(t *Tree, v Order) *Tree {
	if t == nil {
		return &Tree{nil, v, nil}
	}
	if v.Price < t.Value.Price {
		t.Left = insert(t.Left, v)
		return t
	}
	t.Right = insert(t.Right, v)
	return t
}

func Walk(t *Tree, ch chan float32) {
	if t == nil {
		return
	}
	ch <- t.Value.Price
	// fmt.Println(t.Value.Side, " ", t.Value.Price)
	if t.Left != nil {
		go Walk(t.Left, ch)
	}
	if t.Right != nil {
		go Walk(t.Right, ch)
	}
}

func delete(t *Tree, sell Order) *Tree {
	if t == nil {
		return nil
	}

	switch price := sell.Price; {

	case price < t.Value.Price:
		t.Left = delete(t.Left, sell)
	case price > t.Value.Price:
		t.Right = delete(t.Right, sell)
	default:
		if !(sell.Side == t.Value.Side) { //remove if unequal sides
			switch {
			case t.Left == nil && t.Right == nil:
				t = nil
			case t.Left == nil:
				t = t.Right
			case t.Right == nil:
				t = t.Left
			default:
				temp := findMin(t.Right)
				t.Value = temp
				t.Right = delete(t.Right, temp)
			}
		}
	}

	return t
}

func insertOrDelete(t *Tree, order Order) *Tree {
	if t == nil {
		return nil
	}
	if order.Side != t.Value.Side {
		delete(t, order)
	} else {
		insert(t, order)
	}
	return t
}

func findMin(t *Tree) Order {
	if t == nil {
		return t.Value
	}
	if t.Left != nil {
		return findMin(t.Left)
	}
	return t.Value
}

func Walker(t *Tree) <-chan float32 {
	ch := make(chan float32)
	go func() {
		Walk(t, ch)
		close(ch)
	}()
	return ch
}

func sendOrder(order Order, ch chan Order) {
	ch <- order
}

func sendOrders(order ...Order) (chan Order, int) {
	ch := make(chan Order)
	for i := range order {
		go sendOrder(order[i], ch)
	}
	return ch, len(order)
}

func recieveOrder(ch chan Order, ch2 chan *Tree, t *Tree, length int) {
	var temp *Tree
	for i := 0; i < length; i++ {
		temp = insertOrDelete(t, <-ch)
	}
	ch2 <- temp
}

func main() {
	var t *Tree
	t = insert(t, Order{"Buy", 20})
	t = insert(t, Order{"Buy", 10})
	t = insert(t, Order{"Buy", 14})
	t = insert(t, Order{"Buy", 8})
	t = insert(t, Order{"Buy", 40})

	tch := make(chan *Tree)

	ch, length := sendOrders(Order{"Buy", 42}, Order{"Sell", 8})

	go recieveOrder(ch, tch, t, length)

}
