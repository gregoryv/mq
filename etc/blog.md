<a name="top"></a>

# MQTT - exploring alternatives

<div id="about">
Gregory Vin&ccaron;i&cacute;<br>
xx January 2023
</div>

<img src="logo.svg" alt="logo" />


On Aug 3, 2022 efforts began to implement <a
href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">mqtt-v5.0</a>
in Go. [gregoryv/mq](https://github.com/gregoryv/mq) is the
result and this article documents the thoughts around it's design and
efforts of writing it.

<a name="toc"></a>
<div class="anchored">Table of contents <a class="link" href="#toc">§</a></div>
<nav>
	<ul>
		<li><a href="#background">Background</a></li>
		<li><a href="#goal">Goal</a></li>
		<li><a href="#approach">Approach</a></li>
		<li><a href="#design">Design</a></li>
		<ul>
			<li><a href="#queues">Queues</a></li>
			<li><a href="#packageName">Package name</a></li>
		</ul>
		<li><a href="#performance">Performance</a></li>
		<li><a href="#conclusion">Conclusion</a></li>
		<li><a href="#references">References</a></li>
	</ul>
</nav>

<a name="background"></a>
## Background <a class="link" href="#background">§</a>

MQTT is widely used in connected devices or internet of things
(IoT). It is simple and of low bandwidth. In the Go community a few
options are available,
like
[huin/mqtt](https://pkg.go.dev/github.com/huin/mqtt),
[surgemq/message](https://pkg.go.dev/github.com/surgemq/message)
or
[eclipse/paho](https://github.com/eclipse/paho.mqtt.golang).
The latter comes up on top in number of imports and is the one I have
been using.

In one project difficulties where encountered and I needed to learn
more about the protocol details. There where simply to many
abstractions on top of the wire format and I wanted to *simplify*
things. 

The specification is well written with requirements clearly stated.
You can divide the specification into two main areas, 

- wire format and 
- behavior of clients and servers

The wire format is very concrete but the behavior has many optional
capabilities. As such, maybe writing your own client, with only the
things you need, is *simpler* than using a generic one?

Looking into alternatives I found that some packages do provided the
wire format features, though most of them are for mqtt-v3.1.  I could
have opted used e.g. the eclipse/paho package, which has support for
mqtt-v5, though I also wanted to experience and learn more about the
process of implementing a package according to the specification,
something you rarely get to do these days.



<a name="goal"></a>
## Goal <a class="link" href="#goal">§</a>

*Provide a mqtt-v5.0 Go package for writing clients and
servers.*

Its design aims for

1. Correctness - difficult to make protocol mistakes
2. Performance - save the environment, conserve power
3. Simplicity - optional abstractions

With these goals in mind approaching the solution is a tricky thing,
though rewarding.



<a name="approach"></a>
## Approach <a class="link" href="#approach">§</a>

The entire specification is 137 pages. This is quite a lot of
information to read before actually starting anything. Luckily the
first section is *terminology*. Especially useful as it provides the
necessary vocabulary and a general feel for what concepts are
*large* and which are small.

Having poor experience with a top-down development approach I looked
for the smallest concepts to implement first, like constants and error
codes. However these are spread out throughout the specification so I
stopped after a while and decided to move on with the first control
packet, connect. 

Representing the control packet with public fields felt wrong.  I
already knew that in some cases the fields where related and would
make it hard to fulfill the *Correctness*. So I went with the
getter/setter approach, hiding all the fields. The downside being the
documentation is now quite long. Benchmarks between the initial
implementation and pahos showed poor results in several areas, I
needed a redesign.



<a name="design"></a>
## Design <a class="link" href="#design">§</a>

At this point design was limited to performance of control packet
conversion to and from the wire format. But I really wanted to
have a design that was at least in par with pahos, performance wise.
The hidden fields with getter and setter methods had no affect on the
performance so they stayed. But a lot of effort went into designing
the wire types described in the specification in an efficient way but
also somewhat readable. 

Reading and writing packets is deterministic as the length is
provided.  This trait is used in all the wire types to minimize
allocations.

Once the performance was adequate the remaining packets where fairly
quickly implemented.

<a name="queues"></a>
### Queues <a class="link" href="#queues">§</a>

### Package name

The module started out as <code>mqtt</code>, obvious choice which initially worked fine.
Focusing on implementing the control packets.

<pre>
mqtt
mqtt/x
mqtt/proto

mq/x
mq/tt
</pre>

<a name="performance"></a>
## Performance <a class="link" href="#performance">§</a>

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
<em>BenchmarkPublish/our            586962       1974 ns/op      688 B/op      31 allocs/op</em>
BenchmarkPublish/their          479336       2846 ns/op     4064 B/op      41 allocs/op
</pre>

Not a huge difference, but still in the right direction.



## benchmark tt.Client

Initial benchmark, the QoS1 is a lot less efficient but that is due to
the fact that we have to wait for an ack from the server. These
benchmarks do not include any network latencies.

<pre>
goos: linux
goarch: amd64
pkg: github.com/gregoryv/mq/tt
cpu: Intel(R) Xeon(R) E-2288G CPU @ 3.70GHz
BenchmarkClient_PubQoS0-16       1000000              1002 ns/op             560 B/op         11 allocs/op
BenchmarkClient_PubQoS1-16        105883             12934 ns/op            1072 B/op         24 allocs/op
PASS
ok      github.com/gregoryv/mq/tt       3.481s
</pre>

after improving the pool allocation of next packet id

<pre>
BenchmarkClient_PubQoS0-16       1304368              1016 ns/op             560 B/op         11 allocs/op
BenchmarkClient_PubQoS1-16        231313             11030 ns/op            1072 B/op         24 allocs/op
</pre>

<a name="conclusion"></a>
## Conclusion <a class="link" href="#conclusion">§</a>

After D days, C commits and R
releases [gregoryv/mq](https://github.com/gregoryv/mq) is ready for
the scrutiny of the community.

<a name="references"></a>
## References <a class="link" href="#references">§</a>

<ol>
	<li><a href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">mqtt-v5.0</a> specification</li>
	<li><a href="https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang">paho.mqtt.golang</a> implementation</li>
	<li><a href="http://www.rfc-editor.org/info/rfc2119">RFC2119</a></li>
	<li>[huin/mqtt](https://pkg.go.dev/github.com/huin/mqtt) - wire format package for mqtt-v3.1</li>
</ol>
