package random

import (
	"fmt"
	"regexp"
	"testing"
)

var stringRegex = regexp.MustCompile(`^[0-9A-Za-z]+$`)

func TestRandom(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := String(100, 5)
		if !stringRegex.MatchString(s) {
			t.Fatalf("string did not match: %q", s)
		}
	}
}

func makeLenBench(b *testing.B, l uint) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			String(l, l)
		}
	})
}

func BenchmarkString5Len(b *testing.B) {
	fmt.Println("String generation at length: 5")
	makeLenBench(b, 5)
}

func BenchmarkString25Len(b *testing.B) {
	fmt.Println("String generation at length: 25")
	makeLenBench(b, 25)
}

func BenchmarkString100Len(b *testing.B) {
	fmt.Println("String generation at length: 100")
	makeLenBench(b, 100)
}

func BenchmarkString500Len(b *testing.B) {
	fmt.Println("String generation at length: 500")
	makeLenBench(b, 500)
}
