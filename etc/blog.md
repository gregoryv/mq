<a name="top"></a>
# Writing the MQ/TT packages <a class="link" href="blog.html">§</a>

This article documents development efforts of
[mq](https://github.com/gregoryv/mq)/[mq](https://github.com/gregoryv/mq/tt),
an alternative MQTT v5 implementation in Golang.

<nav>
	<ul>
		<li><a href="#background">Background</a></li>
		<li><a href="#goal">Goal</a></li>
		<li><a href="#angleofattack">Angle of attack</a></li>
		<li><a href="#design">Design</a></li>
		<li><a href="#references">References</a></li>
	</ul>
</nav>

<a name="background"></a>
## Background

I've mainly used <a
href="https://github.com/eclipse/paho.mqtt.golang">github.com/eclipse/paho.mqtt.golang</a>
for systems that required MQTT as a communication protocol. In one
such project difficulties where encountered and I needed to learn more
about the protocol details.

The specification has detailed instructions and requirements on most
areas such as the wire format and behavior of clients and
servers. Some things are optional and made me think that maybe it's
actually more benefitial to write your own clients and servers over
using a generic one. 

I could have opted for reusing components of e.g. the paho module but
also wanted to experience what it's like to implement the requirements
of the specification on my own.


<a name="goal"></a>
## Goal

MQTT as a protocol is meant to be small and efficient, I set my goals
accordingly. Having a ready implementation in the paho module made it
easy for me to compare performance. After some thought I didn't only
want performance to be the main goal which could have resulted in
really convoluted code that was hard to understand.  The idea was also
to be able to write my own clients and servers so the three main goals
where defined as

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

<a name="angleofattack"></a>
<h2>Angle of attack</h2>

How do you go about starting to implement a specification of a
protocol? One way would be to start writing the client and then add on
what you need as you go along. I've never been a fan of such top down
development, hard to do test driven developent(TDD) in that scenario.

<a name="design"></a>
<h2>Design</h2>

Package naming ...


<h2>Initial benchmarks and optimization</h2>

<pre>
goos: linux
goarch: amd64
pkg: github.com/gregoryv/mq
cpu: Intel(R) Xeon(R) E-2288G CPU @ 3.70GHz
BenchmarkConnect/make/our     15082816        77.58 ns/op      24 B/op       3 allocs/op
BenchmarkConnect/make/their    3935006       279.30 ns/op     512 B/op       5 allocs/op
<em>BenchmarkConnect/write/our      483277      2096.00 ns/op     48 B/op       1 allocs/op</em>
BenchmarkConnect/write/their   2359382       862.40 ns/op     368 B/op      10 allocs/op
<em>BenchmarkConnect/read/our      1553311       859.40 ns/op    440 B/op       8 allocs/op</em>
BenchmarkConnect/read/their     549508      2507.00 ns/op    3288 B/op      24 allocs/op
</pre>

Writing a control packet uses one allocation but is still a lot slower
than their version when it comes to writing. Though in the reading the
roles are reversed, our version has fewer allocations and is quicker.
We'll have to do an overall test, i.e. reading And writing messages,
and maybe focus on the Publish message.


Using pprof I could find that the slowest part of writing a control
packet was when writing fields defined in the propertyMap. Replacing
the loop with direct access yielded quite an improvement

<pre>
BenchmarkConnect/write/our       483277     2096.00 ns/op      48 B/op       1 allocs/op
... after...
<em>BenchmarkConnect/write/our      7871455       150.6 ns/op      48 B/op       1 allocs/op</em>
</pre>


BenchmarkAuth is faster when successful in pahos favor, though when
including e.g. a reauthenticate with some user properties our
implementation is faster. In the successful case we could optimize it
even further I guess, though that could affect reading of other
control packages. FixedHeader.ReadRemaining was optimised for this
case, though the one allocation in difference was actually incorrectly
calculated as ReadRemaining creates the packet during testing whereas
in their case it was already instantiated outside.

To compare our and their side I'll have to use a more complete test
where a control packet is created, written on the wire and then read
back as another packet.

Weird result when writing the same packet, this could be signifficant
later as pahos implementation may require a new packet each time,
though unlikely.

<pre>
BenchmarkCompare/Auth/our     1789131           798 ns/op       232 B/op   16 allocs/op
BenchmarkCompare/Auth/their    120936        197728 ns/op   1063672 B/op   22 allocs/op
</pre>


A more reasonable comparison

<pre>
Benchmark/Auth/our            1595908         850 ns/op       296 B/op     18 allocs/op
Benchmark/Auth/their           396902        5372 ns/op      4208 B/op     43 allocs/op
Benchmark/Connect/our          675033        1586 ns/op       880 B/op     16 allocs/op
Benchmark/Connect/their        207224        5237 ns/op      5552 B/op     50 allocs/op
<em>Benchmark/Publish/our          504354        1990 ns/op       880 B/op     32 allocs/op</em>
Benchmark/Publish/their        609014        4074 ns/op      4064 B/op     41 allocs/op
</pre>

The most important package Publish is still slower than
pahos. Inlining the creation of packets as would be done in a real
client we should get different results.

<pre>
BenchmarkAuth/our              1808374        682 ns/op      264 B/op      17 allocs/op
BenchmarkAuth/their             513357       4823 ns/op     4208 B/op      43 allocs/op
BenchmarkConnect/our            785091       1311 ns/op      880 B/op      16 allocs/op
BenchmarkConnect/their          205426       6685 ns/op     5552 B/op      50 allocs/op
<b>BenchmarkPublish/our            586962       1974 ns/op      688 B/op      31 allocs/op</b>
BenchmarkPublish/their          479336       2846 ns/op     4064 B/op      41 allocs/op
</pre>

Not a huge difference, but still in the right direction.


<h2>Package naming</h2>

mqtt
mqtt/x
mqtt/proto

mq/x
mq/tt

<a name="references"></a>
<h2>References</h2>

<ol>
	<li><a href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">MQTT Version 5.0</a> specification</li>
	<li><a href="https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang">paho.mqtt.golang</a> implementation</li>
</ol>
