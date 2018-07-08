
> Started off trying to specify what _not_ to shade, but that blew up gloriously.
> So the current state (2018-5-16) is to shade just the problematic bits, with the
> hope/goal of increasing the "shaded surface area" until transitive dependency
> conflicts are a thing of the past.


## Currently shaded things
- `com.google.protobuf`
- `io.netty`


## Things to _NOT_ shade
- `org.apache.spark`.*
- `org.slf4j`.*
- `scala`.*