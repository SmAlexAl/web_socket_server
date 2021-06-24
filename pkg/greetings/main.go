package greetings

import "fmt"

func Hello(name string) string {
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message
}

func Buy(name string) string {
	message := fmt.Sprintf("Buy, %v", name)

	return message
}
