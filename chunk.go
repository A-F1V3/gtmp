package main

import (
  "io"
)

type Chunk struct {
  fmt int
  csid int
  ts int
  tsdelta int
  msid int
  mlen int
  mtypeid int
  size int
  reader io.Reader
}
