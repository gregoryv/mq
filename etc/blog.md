<a name="top"></a>

# MQTT - Exploring Alternatives

<div id="about">
Gregory Vin&ccaron;i&cacute;<br>
xx January 2023
</div>

<img src="logo.svg" alt="logo" />


On Aug 3, 2022 efforts began to implement <a
href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">mqtt-v5.0</a>
in Go, resulting in
packages [gregoryv/mq](https://github.com/gregoryv/mq)
and [gregoryv/tt](https://github.com/gregoryv/tt). This article
documents efforts of writing it and design decisions taken along the
way.

<a name="toc"></a>
<div class="anchored">Table of contents <a class="link" href="#toc">§</a></div>
<nav>
	<ul>
		<li><a href="#background">Background</a></li>
		<li><a href="#goal">Goal</a></li>
		<li><a href="#approach">Approach</a></li>
		<li><a href="#design">Design</a></li>
		<ul>
			<li><a href="#packageName">Package Name</a></li>
		</ul>
		<li><a href="#performance">Performance</a></li>
		<li><a href="#conclusion">Conclusion</a></li>
		<li><a href="#references">References</a></li>
	</ul>
</nav>

<a name="background"></a>
## Background <a class="link" href="#background">§</a>

In a world of connected devices or internet of things (IoT), the MQTT
protocol is used for device to cloud communication.  Its compact size
makes it effective for asynchronous telemetry messaging.

In one project difficulties where encountered and I needed to learn
more about the protocol details. While digging into the specification
I realized there where to many abstractions on top of the wire format
and I needed a way to *simplify* things.  The wire format is concrete
though the behavior of clients and servers contains optional
capabilities. As such, maybe writing your own client, with only the
things you need, is *simpler* than using a generic one?

The specification is well written with requirements clearly stated.
You can divide the specification into two main areas, (1) wire format
and (2) behavior of clients and servers.  Looking into alternatives I
found that some
packages,
[huin/mqtt](https://pkg.go.dev/github.com/huin/mqtt),
[surgemq/message](https://pkg.go.dev/github.com/surgemq/message)
or [eclipse/paho](https://github.com/eclipse/paho.mqtt.golang), do
provided the wire format features, though most are for mqtt-v3.1.

I could have opted to used eclipse/paho, which has support for
mqtt-v5, though I also wanted to experience and learn more about the
process of implementing a package according to the specification,
something you rarely get to do these days.

Putting other projects on hold, I decided to do this thing and set my
goals accordingly.



<a name="goal"></a>
## Goal <a class="link" href="#goal">§</a>

*Provide a mqtt-v5.0 Go package for writing clients and
servers.*

Its design aims for

1. Correctness - difficult to make protocol mistakes
2. Performance - save the environment, conserve power
3. Simplicity - optional abstractions

With these goals in mind approaching the solution is a tricky thing,
though rewarding and fun.



<a name="approach"></a>
## Approach <a class="link" href="#approach">§</a>

The entire specification is 137 pages. This is quite a lot of
information to read before actually starting anything. Luckily the
first section is *terminology*. It provides the necessary vocabulary
and a general feel for what concepts are *large* and which are small.

Having poor experience with a top-down development approach I looked
for the smallest concepts to implement first, like constants and error
codes. However these are spread out throughout the specification so I
stopped after a while and decided to move on with the first control
packet, connect. 

Representing the control packet with public fields felt idiomic. This
approach is used in other packages and I was inclined to follow suit.
However I already knew that in some cases the fields where
related. Setting them from the outside would make it hard to fulfill
the *Correctness* goal. So I went with the getter/setter approach,
hiding all the fields. 

Having pahos implementation I opted to early on write benchmark tests
comparing my approach to theirs. Benchmarks between the initial
implementation and pahos showed poor results in several areas, I
needed a redesign.

<a name="design"></a>
## Design <a class="link" href="#design">§</a>

At this point my design was limited to performance of control packet
conversion to and from the wire format. But I really wanted to
have a design that was at least in par with pahos, performance wise.
The hidden fields with getter and setter methods had no affect on the
performance so they stayed. But a lot of effort went into designing
the wire types described in the specification in an efficient yet readable.

Reading and writing packets is deterministic as the length is
provided. This trait is used in all the wire types to minimize
allocations. Once the performance was adequate the remaining packets
where fairly quickly implemented.

### Package Name

The module started out as <code>mqtt</code>, obvious choice which
initially worked fine. Once the control packet types where implemented
focus shifted to clients and servers. This amount of code in package
mqtt was quite large already so I went with a subdirectory `mqtt/x`
and later renamed it to `mqtt/proto`. The more I worked on client
behavior the naming felt wrong. Not only the naming, also the day to
day work where working in a subdirectory of the package was not
optimal, at least not in my setup. I want to quickly select the
project I work on and stay in that directory for most of the
time. This lead to another round of package renaming. Finally I
decided it was time to split the packages

1. gregoryv/mq
1. gregoryv/tt

Short and concise names, that are related but do not have to
be. I.e. someone else may want to write a generic client of sorts
using `gregoryv/mq`. The packages also reflect the two major areas in
the specification (1) wire format and (2) behavior of clients and
servers.


<a name="performance"></a>
## Performance <a class="link" href="#performance">§</a>

Before I go on, let me first say thank you to the eclipse/paho
developers for their great work and I hope these results may give
ideas to improving their already great package.

Initial comparison was on creation, writing and reading as separate
tests.

<pre>
goos: linux
goarch: amd64
pkg: github.com/gregoryv/mq
cpu: Intel(R) Xeon(R) E-2288G CPU @ 3.70GHz
BenchmarkConnect/make/our     15082816        77.58 ns/op      24 B/op       3 allocs/op
BenchmarkConnect/make/their    3935006       279.30 ns/op     512 B/op       5 allocs/op
<em>BenchmarkConnect/write/our      483277      2096.00 ns/op      48 B/op       1 allocs/op</em>
BenchmarkConnect/write/their   2359382       862.40 ns/op     368 B/op      10 allocs/op
<em>BenchmarkConnect/read/our      1553311       859.40 ns/op     440 B/op       8 allocs/op</em>
BenchmarkConnect/read/their     549508      2507.00 ns/op    3288 B/op      24 allocs/op
</pre>

Writing a control packet uses one allocation but is still a lot slower
than their version when it comes to writing. Though in the reading the
roles are reversed, our version has fewer allocations and is quicker.


Using pprof I could find that the slowest part of writing a control
packet was when writing fields defined as properties. Replacing a loop
with direct access yielded quite an improvement

<pre>
BenchmarkConnect/write/our       483277     2096.00 ns/op      48 B/op       1 allocs/op
... after...
<em>BenchmarkConnect/write/our      7871455       150.6 ns/op      48 B/op       1 allocs/op</em>
</pre>


BenchmarkAuth is faster when successful in pahos favor, though when
including e.g. a reauthenticate with some user properties our
implementation is faster. In the successful case we could optimize it
even further I guess, though that could affect reading of other
control packages. 

At this point I decided to write a more complete benchmark that creates, writes and reads
a control packet. A more reasonable comparison

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

Finally let me group benchmarks related to the publish control packet
which can be argued is the control packet that will flow the most
between a client and server.

<pre>
BenchmarkPublish/our-16            813364    1667 ns/op      688 B/op       31 allocs/op
BenchmarkPublish/their-16          459866    6305 ns/op     5792 B/op       43 allocs/op
BenchmarkPublish/write/our-16     2817781     393 ns/op       80 B/op        1 allocs/op
BenchmarkPublish/write/their-16   1587978     711 ns/op      472 B/op       10 allocs/op
BenchmarkPublish/wqos0/our-16     9936145     120 ns/op       24 B/op        1 allocs/op
BenchmarkPublish/wqos0/their-16   2453695     481 ns/op      408 B/op        9 allocs/op
</pre>

<a name="conclusion"></a>
## Conclusion <a class="link" href="#conclusion">§</a>

After D days and C commits,
packages [gregoryv/mq](https://github.com/gregoryv/mq)
and [gregoryv/mq](https://github.com/gregoryv/mq) are ready for the
scrutiny of the community.

<a name="references"></a>
## References <a class="link" href="#references">§</a>

<ol>
	<li><a href="https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html">mqtt-v5.0</a> specification</li>
	<li><a href="https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang">paho.mqtt.golang</a> implementation</li>
	<li><a href="http://www.rfc-editor.org/info/rfc2119">RFC2119</a> - Key words for use in RFCs to Indicate Requirement Levels</li>
	<li><a href="https://pkg.go.dev/github.com/huin/mqtt">huin/mqtt</a> - wire format package for mqtt-v3.1</li>
</ol>
