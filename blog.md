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
