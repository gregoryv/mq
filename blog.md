# Writing the MQTT module

I've mainly used
https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang for systems
that required MQTT as a communication protocol. In one such project
difficulties where encountered and I needed to learn more about the
protocol details.

The specification has detailed instructions and requirements on most
areas such as the wire format and behavior of clients and
servers. Some things are optional and made me think that maybe it's
actually more benefitial to write your own clients and servers over
using a generic one. 

I could have opted for reusing components of e.g. the paho module but
also wanted to experience what it's like to implement the requirements
of the specification on my own.

As the protocols intention is to be small and efficient, I set my
goals accordingly. Having a ready implementation in the paho module
made it easy for me to compare performance. After some thought I
didn't only want performance to be the main goal which could have
resulted in really convoluted code that was hard to understand.
The idea was also to be able to write my own clients and servers so
the three main goals where defined as

1. Correctness
2. Performance
3. Ease of use

Correctness: came on top and means it should be difficult, though not
impossible, to make protocol mistakes when using the module. I needed
the "not impossible" remark to be able to test error cases.

Performance: this is where I would get a chance to explore more of the
memory alignment and allocation optimizations I've read about but
never tried. Hopefully with benchmarks in place I can provide some
useful insights to the community about either my own improvements or
possible ones in the paho module.



The inverse initial optimization

<pre>
goos: linux
goarch: amd64
pkg: github.com/gregoryv/mqtt
cpu: Intel(R) Xeon(R) E-2288G CPU @ 3.70GHz
BenchmarkConnect/make/our-16     15082816        77.58 ns/op      24 B/op       3 allocs/op
BenchmarkConnect/make/their-16    3935006       279.30 ns/op     512 B/op       5 allocs/op
<b>BenchmarkConnect/write/our-16      483277      2096.00 ns/op      48 B/op       1 allocs/op</b>
BenchmarkConnect/write/their-16   2359382       862.40 ns/op     368 B/op      10 allocs/op
<b>BenchmarkConnect/read/our-16      1553311       859.40 ns/op     440 B/op       8 allocs/op</b>
BenchmarkConnect/read/their-16     549508      2507.00 ns/op    3288 B/op      24 allocs/op
</pre>

Writing a control packet uses one allocation but is still a lot slower
than their version when it comes to writing. Though in the reading the
roles are inversed, our version has fewer allocations and is quicker.
We'll have to do an overall test, i.e. reading And writing messages,
and maybe focus on the Publish message.


Using pprof I could find that the slowest part of writing a control
packet was when writing fields defined in the propertyMap. Replacing
the loop with direct access yielded quite an improvement

<pre>
BenchmarkConnect/write/our-16      <b>7871455       150.6 ns/op</b>      48 B/op       1 allocs/op
BenchmarkConnect/write/their-16    2347669       629.5 ns/op     368 B/op      10 allocs/op
</pre>
