# Deepequal

Package with protobuf-friendly deepequal functionality.

As you probably know, `reflect.Deepequal` should not work for structures generated with protoc-gen-go because of their
hidden service fields. This modified version of protoc `reflect.DeepEqual` takes care of this and call `proto.Equal`
for values of these types.

## Installation

```shell
go get github.com/sirkon/deepequal
```