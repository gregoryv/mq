<a name="top"></a>

# MQTT - exploring alternatives

<div id="about">
Gregory Vin&ccaron;i&cacute;<br>
xx January 2023
</div>

<img src="logo.svg" alt="logo" />


On Aug 3, 2022 efforts began to implement <a
href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">mqtt-v5.0</a>
in Go. [github.com/gregoryv/mq](https://github.com/gregoryv/mq) is the
result and this article documents the thoughts around it's design and
efforts of writing it.

<a name="toc"></a>
<span class="anchored">Table of contents <a class="link" href="#toc">§</a></span>
<nav>
	<ul>
		<li><a href="#background">Background</a></li>
		<li><a href="#goal">Goal</a></li>
		<li><a href="#approach">Approach</a></li>
		<li><a href="#thespec">The specification</a></li>
		<li><a href="#design">Design</a></li>
		<li><a href="#performance">Performance</a></li>
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
[eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang).
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

<a name="approach"></a>
## Approach <a class="link" href="#approach">§</a>

How to implement a protocol specification, such as the [mqtt-v5]? With a top-down
approach , writing the client first and then add on what's needed as
you go along. I've never been a fan of such top down development, felt
that it always resulted in more refactoring and hard to do test driven
developent(TDD).

First thing is to actually read the specification and get a feel for
it. The requirements are spread out within the document which gives
them context and they follow RFC2119 for phrasing, which is nice.

As I read on, navigation became a bit of a hassle. Having to jump up
and down to the table of contents and then into a specific section was
cumbersome with that many sections. This issue became even more
prominent once development started. To remedy that, I saved a local
copy and added anchors with proper names where needed, ie. #connect to
directly jump to the Connect packet section. This enabled me to save
some readable bookmarks.

<a name="design"></a>
## Design <a class="link" href="#design">§</a>

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
<b>BenchmarkPublish/our            586962       1974 ns/op      688 B/op      31 allocs/op</b>
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

<a name="references"></a>
## References <a class="link" href="#references">§</a>

<ol>
	<li><a href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">mqtt-v5.0</a> specification</li>
	<li><a href="https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang">paho.mqtt.golang</a> implementation</li>
	<li><a href="http://www.rfc-editor.org/info/rfc2119">RFC2119</a></li>
</ol>
