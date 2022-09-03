# Changelog
All notable changes to this project will be documented in this file.

The format is based on http://keepachangelog.com/en/1.0.0/
and this project adheres to http://semver.org/spec/v2.0.0.html.

## [unreleased]

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
