# Changelog
All notable changes to this project will be documented in this file.

The format is based on http://keepachangelog.com/en/1.0.0/
and this project adheres to http://semver.org/spec/v2.0.0.html.

## [0.29.0] 2024-12-28

- Fix missing payload of Connect.Will() message
- Connect.Will only sets the will, default delay interval is 0
- Add Connect.SetWillDelayInterval

## [0.28.0] 2024-09-21

- Fix Subscribe.subscriptionID is optional
- Add func Unsubscribe.Filters
- Update dependencies

## [0.27.0] 2023-01-15

- Fix Publish.WriteTo - encoding payload as raw data, not binary data
- Add wiretype rawdata

## [0.26.0] 2022-12-13

- Subscribe is malformed on bad subscriptionID
- Replace Subscribe.AddFilter with AddFilters
- Add method Subscribe.Wellformed

## [0.25.0] 2022-11-26

- Rename methods Reason to ReasonString
- Rename methods SetReason to SetReasonString
- Add reason to Disconnect.String
- Add type HasReason interface

## [0.24.0] 2022-11-25

- Update logo
- Add func Dump
- Include reason code in ConnAck.String
- Add missing Auth properties
- Update blog

## [0.23.0] 2022-11-19

- Add missing reason codes
- Remove method AckType
- Add separate types PubComp, PubRec, PubRel
- Returned *Malformed from Publish.WellFormed
- Remove types Handler, PubHandler and Middleware
  Their counterparts are found in github.com/gregoryv/tt

## [0.22.0] 2022-11-06

- Move tt to github.com/gregoryv/tt
- Remove tt.Interceptor

## [0.21.0] 2022-11-05

- cmd/tt pub and sub commands work with mosquitto(anonymous)
- Change log prefixes to in/out
- Hide types Bits and FixedHeader
- Remove Connect.Flag and ConnAck.Flag methods
- Add Connect.SetWill replacing all will related methods
- Add type PubRel
- Add type UserProperties

## [0.20.0] 2022-11-01

- Return pointer to packet types when using New constructor funcs
- Improve tt.Logger performance
- Move tt.TestServer to mq/tt/raven
- Add Logger.SetMaxIDLen to control length of id

## [0.19.0] 2022-10-21

- Add TestServer with manual responses
- Add types tt.InFlow and tt.OutFlow
- Replace NewQueue with NewInQueue and NewOutQueue
- Move tt/* packages to tt/

## [0.18.0] 2022-10-19

- Add tt/pakio with pakio.Sender and pakio.Receiver
- Remove type Client
- Remove type tt.Settings

## [0.17.0] 2022-10-17

- Add type tt.LogFeature
- Add tt/cmd/ttdemo
- Add types mq.PubHandler, tt.Router and tt.Route
- Remove Context alias

## [0.16.0] 2022-10-14

- Add context to type Handler
- Rename type Fop to FilterOption with alias Opt
- Simpler mq.Client interface with one Send method
- tt.Client uses stack of middlewares
- Add mq.Middleware interface
- Add tt.LogLevel with info and debug
- Remove tt.ackman

## [0.15.0] 2022-10-01

- Add protocol related interfaces in package mq
- Add alias mq.Packet for mq.ControlPacket for shorter receiver funcs
- tt.Client can connect, publish and receive published packets
- Rename package mqtt to mq and put one client implementation in
  subpackage mq/tt

## [0.14.0] 2022-09-18

- Add client/IDPool for reusing packet ids
- Remove unused constants
- Add missing reason codes

## [0.13.0] 2022-09-04

- Try connecting to mosquitto broker
- Optimise PubAck, ConnAck, Publish, Connect

## [0.12.0] 2022-09-03

- Add type Disconnect and Auth
- Add type PingReq and PingResp
- Add initial blog.md

## [0.11.0] 2022-09-03

- Add type Unsubscribe and UnsubAck

## [0.10.0] 2022-09-01

- Prefix all packetIDs in string output with character p
- Add type SubAck
- Fix ReadPacket to match properly for ControlPacket type

## [0.9.0] 2022-08-30

- Add type Subscribe
- Add type PubAck, same for PUBREL, PUBREC and PUBCOMP

## [0.8.0] 2022-08-28

- Optimize Connect memory alignment
- Implement ConnAck.UnmarshalBinary
- Fix Connect flags representation

## [0.7.0] 2022-08-27

- Add type ConnAck
- Make wstring alias of bindata, less memory allocation in
  UnmarshalBinary

## [0.6.0] 2022-08-21

- Add type Publish
- Simplify Connect.UnmarshalBinary

## [0.5.0] 2022-08-20

- Implement Connect.UnmarshalBinary
- Remove Connect.SetWillFlag, done automatically when setting will
  related fields

## [0.4.0] 2022-08-16

- Optimize Connect.WriteTo 

## [0.3.0] 2022-08-14

- Add type Connect, WriteTo matches pahos
- Add connect variable header properties

## [0.2.0] 2022-08-10

- More readable encoding errors
- FixedHeader supports marshal/unmarshal

## [0.1.0] 2022-08-08

- Add types for all data representation
