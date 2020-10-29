benchmarks on 8 core vm on a mostly idle i7-4790k w/ plenty of ram:

```
% go test -v -bench String ./random

=== RUN   TestRandom
--- PASS: TestRandom (0.00s)
goos: linux
goarch: amd64
pkg: code.hollensbe.org/erikh/demo-code/random
BenchmarkString5Len
String generation at length: 5
String generation at length: 5
String generation at length: 5
String generation at length: 5
String generation at length: 5
BenchmarkString5Len-8            2645406               446 ns/op
BenchmarkString25Len
String generation at length: 25
String generation at length: 25
String generation at length: 25
String generation at length: 25
String generation at length: 25
BenchmarkString25Len-8           2413048               511 ns/op
BenchmarkString100Len
String generation at length: 100
String generation at length: 100
String generation at length: 100
String generation at length: 100
BenchmarkString100Len-8          1000000              1128 ns/op
BenchmarkString500Len
String generation at length: 500
String generation at length: 500
String generation at length: 500
String generation at length: 500
BenchmarkString500Len-8           377850              3169 ns/op
PASS
ok      code.hollensbe.org/erikh/demo-code/random       5.759s
```
