package helloWorld

import "testing"

func testHelloWorld(t *testing.T) {
	if helloWorld() != "Hello, world!" {
		t.Error("Incorrect hello world")
	}
}
